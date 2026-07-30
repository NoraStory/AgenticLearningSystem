package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type toolMeta struct {
	ID, Name, Desc, Category string
	Locked, Free             bool
}

var toolCatalog = []toolMeta{
	{ID: "web_search", Name: "联网搜索", Desc: "搜索最新技术资料与官方文档", Category: "信息获取", Free: true},
	{ID: "doc_reader", Name: "文档阅读", Desc: "解析并总结上传文档", Category: "信息获取", Free: true},
	{ID: "code_search", Name: "代码检索", Desc: "在附件或配置的工作区中查找定义与引用", Category: "开发工具", Free: true},
	{ID: "git_helper", Name: "Git 助手", Desc: "读取工作区状态并生成提交说明", Category: "开发工具", Free: true},
	{ID: "leetcode_fetch", Name: "题目获取", Desc: "从站内题库检索并整理算法题", Category: "学习工具", Free: true},
	{ID: "code_execute", Name: "代码执行", Desc: "在受限沙箱中运行代码", Category: "开发工具", Locked: true, Free: true},
	{ID: "sql_explain", Name: "SQL 分析", Desc: "静态解释 SQL 并给出索引优化建议", Category: "开发工具", Free: true},
	{ID: "diagram_gen", Name: "图表生成", Desc: "生成 Mermaid 架构图和流程图", Category: "内容生成", Free: true},
	{ID: "quiz_gen", Name: "测验生成", Desc: "按知识点生成练习题", Category: "学习工具", Free: true},
	{ID: "self_heal", Name: "代码自修复", Desc: "分析错误并提出最小修复", Category: "开发工具", Locked: true},
	{ID: "course_search", Name: "课程检索", Desc: "检索站内课程和章节", Category: "学习工具", Free: true},
	{ID: "progress_query", Name: "进度查询", Desc: "读取当前用户学习进度并生成建议", Category: "学习工具", Free: true},
	{ID: "resume_review", Name: "简历审阅", Desc: "分析简历结构与表达", Category: "职业工具"},
	{ID: "project_review", Name: "项目审阅", Desc: "分析项目源码和任务完成度", Category: "职业工具"},
	{ID: "mindmap_gen", Name: "思维导图", Desc: "把知识整理为 Mermaid 思维导图", Category: "内容生成", Free: true},
}

type agentChatInput struct {
	Message           string         `json:"message"`
	AgentType         string         `json:"agent_type"`
	SessionID         string         `json:"session_id"`
	CollaborationMode string         `json:"collaboration_mode"`
	Context           map[string]any `json:"context"`
	Attachments       []string       `json:"attachments"`
}

func (s *Server) agentChat(c *gin.Context) {
	var in agentChatInput
	if c.ShouldBindJSON(&in) != nil || (strings.TrimSpace(in.Message) == "" && len(in.Attachments) == 0) {
		fail(c, 400, 400, "消息不能为空")
		return
	}
	if in.SessionID == "" {
		in.SessionID = uuid.NewString()
	}
	if in.CollaborationMode == "" {
		in.CollaborationMode = "dynamic"
	}
	uid := userID(c)
	conversationHistory := s.loadConversationMemory(uid, in.SessionID)
	followUp := isContextualFollowUp(in.Message) && len(conversationHistory) > 0
	if in.AgentType == "" {
		if inherited := lastConversationAgent(conversationHistory); followUp && inherited != "" {
			in.AgentType = inherited
		} else {
			in.AgentType = routeAgent(in.Message)
		}
	}
	workflowID := uuid.NewString()
	inputJSON, _ := json.Marshal(in)
	wf := model.WorkflowExecution{ID: workflowID, UserID: uid, Status: "running", CurrentNode: "route", InputJSON: string(inputJSON), ResultJSON: "{}"}
	s.services.DB.Create(&wf)
	s.services.DB.Create(&model.SessionMessage{ID: uuid.NewString(), UserID: uid, SessionID: in.SessionID, Role: "user", Agent: in.AgentType, Content: in.Message, WorkflowJSON: "{}"})
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
	c.Status(http.StatusOK)
	writeSSE(c, "agent_route", gin.H{"agent": in.AgentType, "reason": "\u6839\u636e\u95ee\u9898\u3001\u9875\u9762\u3001\u9644\u4ef6\u548c\u4f1a\u8bdd\u8bb0\u5fc6\u81ea\u52a8\u8def\u7531", "session_id": in.SessionID, "memory_messages": len(conversationHistory), "follow_up": followUp})
	writeSSE(c, "workflow_start", gin.H{"workflow_id": workflowID, "mode": in.CollaborationMode})
	routeStep := gin.H{"id": "route", "name": "智能路由", "status": "completed", "agent": "router"}
	writeSSE(c, "workflow_step", routeStep)
	c.Writer.Flush()
	workflow := []gin.H{routeStep}
	toolContexts := []string{}
	imageDataURLs := []string{}
	for _, plan := range s.planAgentTools(uid, in, conversationHistory) {
		stepID := "tool-" + plan.Meta.ID
		running := gin.H{"id": stepID, "name": plan.Meta.Name, "status": "running", "tool": plan.Meta.ID, "reason": plan.Reason, "phase": plan.Phase, "depends_on": plan.DependsOn}
		writeSSE(c, "workflow_step", running)
		writeSSE(c, "tool_call", gin.H{"tool": plan.Meta.ID, "name": plan.Meta.Name, "reason": plan.Reason, "phase": plan.Phase, "depends_on": plan.DependsOn})
		c.Writer.Flush()
		toolInput := in
		toolInput.Message = contextualToolMessage(in.Message, conversationHistory)
		execution, toolErr := s.executeAgentTool(c.Request.Context(), uid, plan, toolInput)
		status := "completed"
		result := execution.Context
		if toolErr != nil {
			status = "failed"
			result = toolErr.Error()
		} else if !execution.Available {
			status = "unavailable"
		}
		toolContexts = append(toolContexts, result)
		imageDataURLs = append(imageDataURLs, execution.Images...)
		writeSSE(c, "tool_result", gin.H{"tool": plan.Meta.ID, "status": status, "result": result, "available": execution.Available, "phase": plan.Phase})
		completed := gin.H{"id": stepID, "name": plan.Meta.Name, "status": status, "tool": plan.Meta.ID, "result": result, "phase": plan.Phase, "depends_on": plan.DependsOn}
		writeSSE(c, "workflow_step", completed)
		workflow = append(workflow, completed)
		c.Writer.Flush()
	}
	prompt := strings.TrimSpace(in.Message)
	if prompt == "" && len(in.Attachments) > 0 {
		prompt = "\u8bf7\u7b80\u6d01\u8bc6\u522b\u5e76\u5206\u6790\u4e0a\u4f20\u7684\u56fe\u7247\u3002"
	}
	prompt = buildConversationPrompt(prompt, conversationHistory, toolContexts)
	analysisRunning := gin.H{"id": "analyze", "name": "问题分析", "status": "running", "agent": in.AgentType}
	writeSSE(c, "workflow_step", analysisRunning)
	c.Writer.Flush()
	analysisDone := gin.H{"id": "analyze", "name": "问题分析", "status": "completed", "agent": in.AgentType}
	writeSSE(c, "workflow_step", analysisDone)
	answerRunning := gin.H{"id": "answer", "name": "生成回答", "status": "running", "agent": in.AgentType}
	writeSSE(c, "workflow_step", answerRunning)
	c.Writer.Flush()
	systemPrompt := "\u4f60\u662f CodeForge Academy \u7684\u4e2d\u6587\u6280\u672f\u5b66\u4e60\u52a9\u624b\uff0c\u5fc5\u987b\u4f18\u5148\u4f7f\u7528\u5bf9\u8bdd\u8bb0\u5fc6\u89e3\u6790\u6307\u4ee3\u5173\u7cfb\u3002\u6839\u636e\u5de5\u5177\u7ed3\u679c\u51c6\u786e\u56de\u7b54\uff0c\u4f18\u5148\u7ed9\u51fa\u7ed3\u8bba\u518d\u7ed9\u6b65\u9aa4\u3002\u5982\u679c\u5de5\u5177\u4e0a\u4e0b\u6587\u5305\u542b Mermaid \u4ee3\u7801\u5757\uff0c\u5fc5\u987b\u539f\u6837\u8f93\u51fa\u5b8c\u6574\u4ee3\u7801\u5757\uff0c\u4e0d\u8981\u53ea\u7528\u6587\u5b57\u6982\u8ff0\u66ff\u4ee3\u3002"
	if len(imageDataURLs) > 0 {
		systemPrompt = "你是 CodeForge Academy 的视觉技术助手。快速识别图片中的文字、界面、图表、代码和错误信息；先直接回答用户问题，再补充关键依据，控制在 6 个要点以内。"
	}
	var answer strings.Builder
	streamErr := s.llm.StreamChatWithImages(c.Request.Context(), systemPrompt, prompt, imageDataURLs, func(delta string) {
		answer.WriteString(delta)
		writeSSE(c, "token", gin.H{"content": delta})
		c.Writer.Flush()
	})
	if streamErr != nil {
		failure := "模型服务暂时不可用：" + streamErr.Error()
		if answer.Len() == 0 {
			answer.WriteString(failure)
			writeSSE(c, "token", gin.H{"content": failure})
			c.Writer.Flush()
		}
	}
	answerDone := gin.H{"id": "answer", "name": "生成回答", "status": "completed", "agent": in.AgentType}
	writeSSE(c, "workflow_step", answerDone)
	workflow = append(workflow, analysisDone, answerDone)
	wfJSON, _ := json.Marshal(workflow)
	s.services.DB.Create(&model.SessionMessage{ID: uuid.NewString(), UserID: uid, SessionID: in.SessionID, Role: "assistant", Agent: in.AgentType, Content: answer.String(), WorkflowJSON: string(wfJSON)})
	resultJSON, _ := json.Marshal(gin.H{"answer": answer.String(), "session_id": in.SessionID})
	s.services.DB.Model(&wf).Updates(map[string]any{"status": "completed", "current_node": "done", "result_json": string(resultJSON)})
	s.addActivity(uid, "agent", "与 "+in.AgentType+" 完成了一次动态工具协作")
	writeSSE(c, "done", gin.H{"session_id": in.SessionID, "workflow_id": workflowID, "agent": in.AgentType})
	c.Writer.Flush()
}
func writeSSE(c *gin.Context, event string, data any) {
	raw, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, raw)
}
func chunks(s string, size int) []string {
	r := []rune(s)
	out := []string{}
	for i := 0; i < len(r); i += size {
		end := i + size
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[i:end]))
	}
	return out
}
func routeAgent(q string) string {
	q = strings.ToLower(q)
	switch {
	case strings.Contains(q, "简历"):
		return "career"
	case strings.Contains(q, "报错") || strings.Contains(q, "代码") || strings.Contains(q, "bug"):
		return "code-review"
	case strings.Contains(q, "算法") || strings.Contains(q, "题目"):
		return "problem-explain"
	case strings.Contains(q, "计划") || strings.Contains(q, "路径"):
		return "planner"
	case strings.Contains(q, "项目") || strings.Contains(q, "架构"):
		return "project"
	default:
		return "learning-assistant"
	}
}
func (s *Server) agentHistory(c *gin.Context) {
	q := s.services.DB.Where("user_id = ?", userID(c))
	if v := c.Query("session_id"); v != "" {
		q = q.Where("session_id = ?", v)
	}
	var rows []model.SessionMessage
	q.Order("created_at asc").Limit(200).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, v := range rows {
		var workflow any
		if v.WorkflowJSON != "" {
			_ = json.Unmarshal([]byte(v.WorkflowJSON), &workflow)
		}
		items = append(items, gin.H{"id": v.ID, "session_id": v.SessionID, "role": v.Role, "agent": v.Agent, "content": v.Content, "workflow": workflow, "created_at": v.CreatedAt})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) agentUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		fail(c, 400, 400, "上传格式错误")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		fail(c, 400, 400, "请选择文件")
		return
	}
	urls := []string{}
	for _, h := range files {
		if h.Size > 20<<20 {
			continue
		}
		key, err := s.store.Save(c, "chats/"+userID(c), h)
		if err == nil {
			urls = append(urls, key)
		}
	}
	success(c, gin.H{"file_urls": urls})
}
func (s *Server) agentTools(c *gin.Context) {
	var settings []model.UserToolSetting
	s.services.DB.Where("user_id = ?", userID(c)).Find(&settings)
	set := map[string]bool{}
	seen := map[string]bool{}
	for _, v := range settings {
		set[v.ToolID] = v.Enabled
		seen[v.ToolID] = true
	}
	items := make([]gin.H, 0, len(toolCatalog))
	for _, t := range toolCatalog {
		enabled := true
		if seen[t.ID] {
			enabled = set[t.ID]
		}
		usable, reason := toolCapability(t.ID)
		items = append(items, gin.H{"id": t.ID, "name": t.Name, "desc": t.Desc, "category": t.Category, "enabled": enabled, "locked": t.Locked, "free": t.Free, "usable": usable, "reason": reason, "workflow_ready": usable})
	}
	success(c, gin.H{"tools": items})
}
func (s *Server) patchAgentTool(c *gin.Context) {
	id := c.Param("id")
	var meta *toolMeta
	for i := range toolCatalog {
		if toolCatalog[i].ID == id {
			meta = &toolCatalog[i]
			break
		}
	}
	if meta == nil {
		fail(c, 404, 404, "Tool 不存在")
		return
	}
	if meta.Locked {
		fail(c, 403, 403, "该 Tool 由系统锁定，不能关闭")
		return
	}
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "请求格式错误")
		return
	}
	var row model.UserToolSetting
	err := s.services.DB.Where("user_id = ? AND tool_id = ?", userID(c), id).First(&row).Error
	if err != nil {
		row = model.UserToolSetting{ID: uuid.NewString(), UserID: userID(c), ToolID: id, Enabled: in.Enabled}
		s.services.DB.Create(&row)
	} else {
		s.services.DB.Model(&row).Update("enabled", in.Enabled)
	}
	success(c, gin.H{"id": id, "enabled": in.Enabled})
}
func (s *Server) agentProfile(c *gin.Context) {
	var p model.UserProfile
	if s.services.DB.Where("user_id = ?", userID(c)).First(&p).Error != nil {
		p = model.UserProfile{ID: uuid.NewString(), UserID: userID(c), Level: "初级开发者", FocusAreas: []string{}, WeakAreas: []string{}, LearningStyle: "实践型", PreferredDifficulty: "简单", DailyGoal: 30}
		s.services.DB.Create(&p)
	}
	success(c, gin.H{"level": p.Level, "focus_areas": p.FocusAreas, "focusAreas": p.FocusAreas, "weak_areas": p.WeakAreas, "weakAreas": p.WeakAreas, "learning_style": p.LearningStyle, "learningStyle": p.LearningStyle, "preferred_difficulty": p.PreferredDifficulty, "preferredDifficulty": p.PreferredDifficulty, "daily_goal": p.DailyGoal, "dailyGoal": p.DailyGoal, "total_study_time": p.TotalStudyTime, "totalStudyTime": p.TotalStudyTime, "streak": p.Streak})
}
func (s *Server) agentKnowledge(c *gin.Context) {
	var k model.UserKnowledgeGraph
	if s.services.DB.Where("user_id = ?", userID(c)).First(&k).Error != nil {
		k = model.UserKnowledgeGraph{ID: uuid.NewString(), UserID: userID(c), Areas: []model.KnowledgeArea{}, RecentTopics: []model.Topic{}}
		s.services.DB.Create(&k)
	}
	success(c, gin.H{"areas": k.Areas, "recent_topics": k.RecentTopics, "recentTopics": k.RecentTopics})
}
func (s *Server) confirmWorkflow(c *gin.Context) {
	var in struct {
		WorkflowID string `json:"workflow_id"`
		NodeID     string `json:"node_id"`
		Approved   bool   `json:"approved"`
		Feedback   string `json:"feedback"`
	}
	if c.ShouldBindJSON(&in) != nil || in.WorkflowID == "" {
		fail(c, 400, 400, "工作流参数无效")
		return
	}
	var wf model.WorkflowExecution
	if s.services.DB.Where("id = ? AND user_id = ?", in.WorkflowID, userID(c)).First(&wf).Error != nil {
		fail(c, 404, 404, "工作流不存在")
		return
	}
	status := "cancelled"
	if in.Approved {
		status = "completed"
	}
	result, _ := json.Marshal(gin.H{"approved": in.Approved, "feedback": in.Feedback})
	s.services.DB.Model(&wf).Updates(map[string]any{"status": status, "current_node": in.NodeID, "result_json": string(result)})
	success(c, gin.H{"status": status, "result": gin.H{"approved": in.Approved, "feedback": in.Feedback}})
}
func (s *Server) workflowStatus(c *gin.Context) {
	var wf model.WorkflowExecution
	if s.services.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID(c)).First(&wf).Error != nil {
		fail(c, 404, 404, "工作流不存在")
		return
	}
	var input, result any
	json.Unmarshal([]byte(wf.InputJSON), &input)
	json.Unmarshal([]byte(wf.ResultJSON), &result)
	success(c, gin.H{"workflow_id": wf.ID, "status": wf.Status, "current_node": wf.CurrentNode, "input": input, "result": result, "created_at": wf.CreatedAt, "updated_at": wf.UpdatedAt})
}

var _ = time.Now
