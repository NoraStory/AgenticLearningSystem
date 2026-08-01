package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"codeforge/backend/internal/model"
)

const (
	maxToolRetries = 5
)

// promptCache 缓存从文件读取的系统提示词，避免每次请求都读磁盘。
var promptCache struct {
	content string
	modTime time.Time
}

// loadSystemPrompt 从 backend/prompts/system_prompt.md 读取系统提示词，
// 文件变更时自动重新加载。读取失败时回退到内置默认。
func loadSystemPrompt() string {
	paths := []string{
		"prompts/system_prompt.md",
		"backend/prompts/system_prompt.md",
		filepath.Join("backend", "prompts", "system_prompt.md"),
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !promptCache.modTime.IsZero() && info.ModTime().Equal(promptCache.modTime) && promptCache.content != "" {
			return promptCache.content
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		promptCache.content = strings.TrimSpace(string(data))
		promptCache.modTime = info.ModTime()
		return promptCache.content
	}
	return "你是 CodeForge Academy 的中文技术学习助手。回答必须完整、充分，把用户的问题解释清楚。如果工具调用失败，诚实告知后使用已有知识回答。"
}

// buildSystemPrompt 构建系统提示词：从文件读取基础内容，注入当前时间。
func buildSystemPrompt(imageMode bool) string {
	base := loadSystemPrompt()
	now := time.Now().Format("2006-01-02 15:04:05 Monday")
	var b strings.Builder
	if imageMode {
		b.WriteString("你是 CodeForge Academy 的视觉技术助手。仔细识别图片中的文字、界面、图表、代码和错误信息。先直接回答用户问题，再补充详细依据和解读——完整描述图片内容、指出关键信息、给出后续建议。如果图片是代码或报错，逐行分析并给出修复方案。不要只用一两句话敷衍。\n\n")
	} else {
		b.WriteString(base)
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "当前时间：%s（Asia/Shanghai）。回答中涉及时间相关内容时请以此为准。", now)
	return b.String()
}

// buildSystemPromptForUser 构建带用户画像的系统提示词：保留热加载基础内容与时间注入，
// 追加当前用户的学习画像摘要（弱项/掌握度/学习时长），供模型个性化回答。
func (s *Server) buildSystemPromptForUser(imageMode bool, uid string) string {
	base := buildSystemPrompt(imageMode)
	digest := s.profileDigest(uid)
	if digest == "" {
		return base
	}
	return base + "\n\n—— 当前用户学习画像 ——\n" + digest +
		"\n（以上数据来自用户学习记录，具体以工具返回为准，不要编造数字）"
}

// profileDigest 汇总当前用户画像摘要：等级、累计学习、连续打卡、各方向掌握度、待加强 top3。
// 无画像数据时返回空串（调用方不注入，保持旧行为）。
func (s *Server) profileDigest(uid string) string {
	var p model.UserProfile
	if err := s.services.DB.Where("user_id = ?", uid).First(&p).Error; err != nil {
		return ""
	}
	var states []model.KnowledgeState
	s.services.DB.Where("user_id = ?", uid).Limit(200).Find(&states)
	if len(states) == 0 && p.TotalStudyTime == 0 {
		return ""
	}
	// 各方向平均掌握度
	catMastery := map[string]float64{}
	catCount := map[string]int{}
	for _, st := range states {
		catMastery[st.Category] += st.Mastery
		catCount[st.Category]++
	}
	// 待加强 top3（掌握度升序，取有练习记录的）
	type catAvg struct {
		key     string
		mastery float64
	}
	var weak []catAvg
	for k, sum := range catMastery {
		if catCount[k] == 0 {
			continue
		}
		weak = append(weak, catAvg{k, sum / float64(catCount[k])})
	}
	sort.Slice(weak, func(i, j int) bool { return weak[i].mastery < weak[j].mastery })
	if len(weak) > 3 {
		weak = weak[:3]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "等级：%s\n", p.Level)
	fmt.Fprintf(&b, "累计学习：%d 小时 · 连续打卡：%d 天\n", p.TotalStudyTime, p.Streak)
	if len(weak) > 0 {
		parts := make([]string, 0, len(weak))
		for _, w := range weak {
			label := w.key
			if v, ok := categoryLabelCN[w.key]; ok {
				label = v
			}
			parts = append(parts, label+" "+strconv.Itoa(int(w.mastery*100))+"%")
		}
		b.WriteString("各方向掌握度：" + strings.Join(parts, " / ") + "\n")
		b.WriteString("待加强：" + strings.Join(parts, "、"))
	}
	return b.String()
}

// extractSearchKeywords 从用户消息中提取搜索关键词。
// 去掉常见的无意义词，保留核心内容，限制长度避免 SearXNG 超时。
func extractSearchKeywords(message string) string {
	text := strings.TrimSpace(message)
	if text == "" {
		return text
	}
	// 去掉常见的对话性前缀
	stopwords := []string{"帮我", "请帮", "请问", "我想", "我要", "能不能", "可以", "帮我分析一下", "帮我分析", "帮我看看", "帮我查", "什么是", "介绍一下", "解释一下", "详细解释一下", "详细说一说", "说一说", "聊一聊"}
	for _, sw := range stopwords {
		text = strings.TrimPrefix(text, sw)
	}
	text = strings.TrimSpace(text)
	// 限制搜索关键词长度，避免超长查询导致 SearXNG 超时
	runes := []rune(text)
	if len(runes) > 60 {
		text = string(runes[:60])
	}
	return text
}

// executeWithRetry 对工具调用做最多 maxToolRetries 次重试。
// 每次重试间隔递增（1s, 2s, 3s...）。返回最终结果和尝试次数。
func (s *Server) executeWithRetry(ctx context.Context, uid string, plan plannedTool, in agentChatInput, maxRetries int) (toolExecution, int, error) {
	var lastErr error
	var exec toolExecution
	attempts := 0
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attempts = attempt
		exec, lastErr = s.executeAgentTool(ctx, uid, plan, in)
		if lastErr == nil {
			return exec, attempts, nil
		}
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return exec, attempts, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
	}
	return exec, attempts, lastErr
}

// toolFailureContext 在工具全部失败后，构建告知模型的上下文文本。
func toolFailureContext(toolName string, attempts int, errMsg string) string {
	return fmt.Sprintf("工具 %s 调用失败（已重试 %d 次，上限 %d 次）。错误：%s。请使用你已有的知识回答用户问题。", toolName, attempts, maxToolRetries, errMsg)
}

var _ = model.SessionMessage{}
