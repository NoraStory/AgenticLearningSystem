package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"codeforge/backend/internal/config"
	"codeforge/backend/internal/model"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/agenticark"
	"github.com/eino-contrib/jsonschema"
	"github.com/gin-gonic/gin"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// ---- Eino 工具循环：用模型自带工具调用替代关键词规划 ----

// einoTool 把平台工具目录里的一个工具包装成 eino 可注册的工具。
type einoTool struct {
	meta toolMeta
	call func(ctx context.Context, in agentChatInput) (toolExecution, error)
}

// agenticModel 是加了工具后的模型实例，仅在 LLM 可用时创建。
var (
	agenticOnce sync.Once
	agenticMod  einomodel.AgenticModel
	agenticErr  error
)

// newAgenticModel 构建绑定平台工具目录的 AgenticModel。
// 通过 agenticOnce 只构建一次，避免每次请求重建 HTTP 客户端。
func newAgenticModel(cfg config.Config) (einomodel.AgenticModel, error) {
	agenticOnce.Do(func() {
		if cfg.ArkAPIKey == "" || cfg.ArkModel == "" {
			agenticErr = fmt.Errorf("LLM 未配置")
			return
		}
		modelInst, err := agenticark.New(context.Background(), &agenticark.Config{
			BaseURL: cfg.ArkBaseURL,
			APIKey:  cfg.ArkAPIKey,
			Model:   cfg.ArkModel,
			Timeout: pointer(120 * time.Second),
		})
		if err != nil {
			agenticErr = fmt.Errorf("创建 Ark AgenticModel 失败: %w", err)
			return
		}
		agenticMod, agenticErr = modelInst.WithTools(einoToolInfos())
	})
	return agenticMod, agenticErr
}

func pointer[T any](v T) *T { return &v }

// einoToolInfos 把工具目录转成 eino 的 ToolInfo 列表。
// 每个工具接收一个字符串参数 query，由模型决定传什么。
func einoToolInfos() []*schema.ToolInfo {
	infos := make([]*schema.ToolInfo, 0, len(toolCatalog))
	for _, t := range toolCatalog {
		props := orderedmap.New[string, *jsonschema.Schema]()
		props.Set("query", &jsonschema.Schema{
			Type:        "string",
			Description: "传递给该工具的具体指令或查询内容",
		})
		info := &schema.ToolInfo{
			Name: t.ID,
			Desc: t.Desc,
			ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&jsonschema.Schema{
				Type:       "object",
				Properties: props,
				Required:   []string{"query"},
			}),
		}
		infos = append(infos, info)
	}
	return infos
}

// executeEinoTool 执行模型请求的工具调用，返回工具结果文本。
// 复用现有 executeAgentTool 执行器，失败时通过 executeWithRetry 重试。
func (s *Server) executeEinoTool(ctx context.Context, uid string, name, args string, in agentChatInput) (string, error) {
	var meta *toolMeta
	for i := range toolCatalog {
		if toolCatalog[i].ID == name {
			meta = &toolCatalog[i]
			break
		}
	}
	if meta == nil {
		return "", fmt.Errorf("工具 %s 未注册", name)
	}
	var parsed struct {
		Query string `json:"query"`
	}
	if strings.TrimSpace(args) != "" {
		_ = json.Unmarshal([]byte(args), &parsed)
	}
	if parsed.Query == "" {
		parsed.Query = in.Message
	}
	toolInput := in
	toolInput.Message = parsed.Query
	execution, attempts, err := s.executeWithRetry(ctx, uid, plannedTool{Meta: *meta}, toolInput, maxToolRetries)
	if err != nil {
		return "", fmt.Errorf("工具 %s 执行失败（已重试 %d 次）: %v", name, attempts, err)
	}
	return execution.Context, nil
}

// runAgentLoop 用 eino AgenticModel 跑工具循环：
// 1. 把用户消息 + 系统提示作为输入，调 Stream（带工具，流式输出）
// 2. 流中聚合 FunctionToolCall；assistant_gen_text 增量实时 onToken（保持前端打字机效果）
// 3. 有工具调用时逐个执行（SSE 推送 workflow_step/tool_call/tool_result），结果回填后继续下一轮
// 4. 无工具调用时结束，返回完整回答
// 每轮固定禁用思考：带 reasoning 的模型会先输出大段思考过程，消耗输出预算。
func (s *Server) runAgentLoop(c *gin.Context, uid, systemPrompt, userPrompt string, in agentChatInput, history []model.SessionMessage, onToken func(string)) (string, error) {
	mod, err := newAgenticModel(s.cfg)
	if err != nil {
		return "", err
	}

	thinkingOpt := agenticark.WithThinking(&responses.ResponsesThinking{Type: responses.ThinkingType_disabled.Enum()})
	toolOpt := einomodel.WithTools(einoToolInfos())

	messages := []*schema.AgenticMessage{
		schema.SystemAgenticMessage(systemPrompt),
		schema.UserAgenticMessage(userPrompt),
	}

	var full strings.Builder
	var toolResultParts []string
	maxRounds := 5
	emptyRounds := 0
	for round := 0; round < maxRounds; round++ {
		stream, err := mod.Stream(c.Request.Context(), messages, thinkingOpt, toolOpt)
		if err != nil {
			return full.String(), err
		}
		var calls []*schema.FunctionToolCall
		for {
			chunk, recvErr := stream.Recv()
			if recvErr != nil {
				break
			}
			for _, block := range chunk.ContentBlocks {
				if block.FunctionToolCall != nil {
					calls = append(calls, block.FunctionToolCall)
				}
				if block.AssistantGenText != nil {
					text := block.AssistantGenText.Text
					if text != "" {
						full.WriteString(text)
						onToken(text)
					}
				}
			}
		}
		stream.Close()
		if len(calls) == 0 {
			return full.String(), nil
		}
		// 防御：过滤空名称/空调用的异常工具调用，连续出现则视为模型异常，停止循环
		validCalls := calls[:0]
		for _, call := range calls {
			if strings.TrimSpace(call.Name) != "" {
				validCalls = append(validCalls, call)
			}
		}
		if len(validCalls) == 0 {
			emptyRounds++
			if emptyRounds >= 2 {
				return full.String(), nil
			}
			continue
		}
		emptyRounds = 0

		// 把模型回合（含工具调用）回填对话，再为每个调用追加工具结果回合
		assistantMsg := &schema.AgenticMessage{
			Role: schema.AgenticRoleTypeAssistant,
			ContentBlocks: []*schema.ContentBlock{
				schema.NewContentBlock(&schema.AssistantGenText{Text: full.String()}),
			},
		}
		for _, call := range validCalls {
			if call.CallID == "" {
				call.CallID = fmt.Sprintf("call-%d", time.Now().UnixNano())
			}
			assistantMsg.ContentBlocks = append(assistantMsg.ContentBlocks, schema.NewContentBlock(call))
		}
		messages = append(messages, assistantMsg)

		for _, call := range validCalls {
			stepID := "tool-" + call.Name
			running := gin.H{"id": stepID, "name": toolDisplayName(call.Name), "status": "running", "tool": call.Name, "reason": "AI 决策:根据当前问题动态选择工具"}
			writeSSE(c, "workflow_step", running)
			writeSSE(c, "tool_call", gin.H{"tool": call.Name, "name": toolDisplayName(call.Name), "reason": "AI 决策:根据当前问题动态选择工具", "phase": "dynamic"})
			c.Writer.Flush()

			resultText, toolErr := s.executeEinoTool(c.Request.Context(), uid, call.Name, call.Arguments, in)
			status := "completed"
			if toolErr != nil {
				status = "failed"
				resultText = toolErr.Error()
				writeSSE(c, "tool_failure", gin.H{"tool": call.Name, "attempts": maxToolRetries, "max_retries": maxToolRetries, "error": toolErr.Error(), "total_failures": 1})
			}
			writeSSE(c, "tool_result", gin.H{"tool": call.Name, "status": status, "result": resultText, "available": true, "phase": "dynamic", "attempts": 1})
			completed := gin.H{"id": stepID, "name": toolDisplayName(call.Name), "status": status, "tool": call.Name, "result": resultText, "phase": "dynamic", "attempts": 1}
			writeSSE(c, "workflow_step", completed)
			c.Writer.Flush()

			resultBlock := &schema.FunctionToolResult{
				CallID:  call.CallID,
				Name:    call.Name,
				Content: []*schema.FunctionToolResultContentBlock{{Type: schema.FunctionToolResultContentBlockTypeText, Text: &schema.UserInputText{Text: resultText}}},
			}
			// AgenticMessage 无独立 tool role：工具结果以 FunctionToolResult block 挂在 user 消息里回填
			messages = append(messages, &schema.AgenticMessage{
				Role:          schema.AgenticRoleTypeUser,
				ContentBlocks: []*schema.ContentBlock{schema.NewContentBlock(resultBlock)},
			})
			toolResultParts = append(toolResultParts, resultText)
		}
	}

	return full.String(), fmt.Errorf("工具循环超过 %d 轮，已停止", maxRounds)
}

func toolDisplayName(id string) string {
	for _, t := range toolCatalog {
		if t.ID == id {
			return t.Name
		}
	}
	return id
}
