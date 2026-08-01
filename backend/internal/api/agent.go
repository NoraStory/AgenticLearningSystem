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
	imageDataURLs := []string{}
	prompt := strings.TrimSpace(in.Message)
	if prompt == "" && len(in.Attachments) > 0 {
		prompt = "\u8bf7\u7b80\u6d01\u8bc6\u522b\u5e76\u5206\u6790\u4e0a\u4f20\u7684\u56fe\u7247\u3002"
	}
	prompt = buildConversationPrompt(prompt, conversationHistory, nil)
	analysisRunning := gin.H{"id": "analyze", "name": "问题分析", "status": "running", "agent": in.AgentType}
	writeSSE(c, "workflow_step", analysisRunning)
	c.Writer.Flush()
	analysisDone := gin.H{"id": "analyze", "name": "问题分析", "status": "completed", "agent": in.AgentType}
	writeSSE(c, "workflow_step", analysisDone)
	answerRunning := gin.H{"id": "answer", "name": "生成回答", "status": "running", "agent": in.AgentType}
	writeSSE(c, "workflow_step", answerRunning)
	c.Writer.Flush()
	// 生成回答：优先走 eino 工具循环（模型自主决策工具调用），LLM 不可用时回退流式直答
	systemPrompt := buildSystemPrompt(false)
	if len(imageDataURLs) > 0 {
		systemPrompt = buildSystemPrompt(true)
	}
	var answer strings.Builder
	if s.llm.Enabled() {
		if _, loopErr := s.runAgentLoop(c, uid, systemPrompt, prompt, in, conversationHistory, func(delta string) {
			answer.WriteString(delta)
			writeSSE(c, "token", gin.H{"content": delta})
			c.Writer.Flush()
		}); loopErr != nil {
			answer.WriteString("（工具协作出错：" + loopErr.Error() + "）")
		}
	} else {
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
	}
	if answer.Len() == 0 {
		answer.WriteString("（未能生成回答，请稍后重试）")
		writeSSE(c, "token", gin.H{"content": answer.String()})
		c.Writer.Flush()
	}
	answerDone := gin.H{"id": "answer", "name": "生成回答", "status": "completed", "agent": in.AgentType}
	writeSSE(c, "workflow_step", answerDone)
	workflow = append(workflow, analysisDone, answerDone)
	wfJSON, _ := json.Marshal(workflow)
	s.services.DB.Create(&model.SessionMessage{ID: uuid.NewString(), UserID: uid, SessionID: in.SessionID, Role: "assistant", Agent: in.AgentType, Content: answer.String(), WorkflowJSON: string(wfJSON)})
	resultJSON, _ := json.Marshal(gin.H{"answer": answer.String(), "session_id": in.SessionID})
	s.services.DB.Model(&wf).Updates(map[string]any{"status": "completed", "current_node": "done", "result_json": string(resultJSON)})
	s.addActivity(uid, "agent", "与 "+in.AgentType+" 完成了一次动态工具协作")
	s.updateProfileFromActivity(uid)
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
	level, levelTitle := computeLevel(p.TotalStudyTime, 0, p.ProblemSolvedCount, p.Streak)
	success(c, gin.H{"level": p.Level, "computed_level": level, "level_title": levelTitle, "focus_areas": p.FocusAreas, "focusAreas": p.FocusAreas, "weak_areas": p.WeakAreas, "weakAreas": p.WeakAreas, "learning_style": p.LearningStyle, "learningStyle": p.LearningStyle, "preferred_difficulty": p.PreferredDifficulty, "preferredDifficulty": p.PreferredDifficulty, "preferred_time_slot": p.PreferredTimeSlot, "daily_goal": p.DailyGoal, "dailyGoal": p.DailyGoal, "total_study_time": p.TotalStudyTime, "totalStudyTime": p.TotalStudyTime, "streak": p.Streak, "session_count": p.SessionCount, "problem_solved_count": p.ProblemSolvedCount, "problem_accuracy": p.ProblemAccuracy, "last_active_at": p.LastActiveAt})
}
func (s *Server) updateProfile(c *gin.Context) {
	var in struct {
		LearningStyle       *string  `json:"learning_style"`
		PreferredDifficulty *string  `json:"preferred_difficulty"`
		PreferredTimeSlot   *string  `json:"preferred_time_slot"`
		DailyGoal           *int     `json:"daily_goal"`
		FocusAreas          []string `json:"focus_areas"`
		WeakAreas           []string `json:"weak_areas"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "请求格式错误")
		return
	}
	uid := userID(c)
	var p model.UserProfile
	if s.services.DB.Where("user_id = ?", uid).First(&p).Error != nil {
		p = model.UserProfile{ID: uuid.NewString(), UserID: uid, Level: "新手开发者", FocusAreas: []string{}, WeakAreas: []string{}, LearningStyle: "实践型", PreferredDifficulty: "简单", DailyGoal: 30}
		s.services.DB.Create(&p)
	}
	updates := map[string]any{}
	if in.LearningStyle != nil { updates["learning_style"] = *in.LearningStyle }
	if in.PreferredDifficulty != nil { updates["preferred_difficulty"] = *in.PreferredDifficulty }
	if in.PreferredTimeSlot != nil { updates["preferred_time_slot"] = *in.PreferredTimeSlot }
	if in.DailyGoal != nil { updates["daily_goal"] = *in.DailyGoal }
	if in.FocusAreas != nil { updates["focus_areas"] = in.FocusAreas }
	if in.WeakAreas != nil { updates["weak_areas"] = in.WeakAreas }
	if len(updates) > 0 {
		s.services.DB.Model(&p).Updates(updates)
	}
	s.services.DB.Where("user_id = ?", uid).First(&p)
	level, levelTitle := computeLevel(p.TotalStudyTime, 0, p.ProblemSolvedCount, p.Streak)
	success(c, gin.H{
		"level": p.Level, "computed_level": level, "level_title": levelTitle,
		"focus_areas": p.FocusAreas, "weak_areas": p.WeakAreas,
		"learning_style": p.LearningStyle, "preferred_difficulty": p.PreferredDifficulty,
		"preferred_time_slot": p.PreferredTimeSlot, "daily_goal": p.DailyGoal,
		"total_study_time": p.TotalStudyTime, "streak": p.Streak,
		"session_count": p.SessionCount, "problem_solved_count": p.ProblemSolvedCount,
		"problem_accuracy": p.ProblemAccuracy, "last_active_at": p.LastActiveAt,
	})
}

func (s *Server) knowledgeStates(c *gin.Context) {
	var rows []model.KnowledgeState
	q := s.services.DB.Where("user_id = ?", userID(c))
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	q.Order("updated_at desc").Limit(200).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		items = append(items, gin.H{
			"skill_name": r.SkillName, "category": r.Category,
			"mastery": r.Mastery, "attempts": r.Attempts,
			"correct_count": r.CorrectCount, "last_practiced_at": r.LastPracticedAt,
		})
	}
	success(c, gin.H{"items": items})
}

func (s *Server) agentDashboard(c *gin.Context) {
	uid := userID(c)
	// 1. 热力图：最近 30 天每日学习时长 + 对话次数
	type heatRow struct {
		Date     time.Time
		Minutes  int
		Sessions int
	}
	var heat []heatRow
	s.services.DB.Model(&model.DailyStudyTime{}).
		Select("study_date as date, sum(duration_minutes) as minutes, count(*) as sessions").
		Where("user_id = ? AND study_date >= ?", uid, time.Now().AddDate(0, 0, -30)).
		Group("study_date").Order("study_date").Scan(&heat)
	heatItems := make([]gin.H, 0, len(heat))
	for _, h := range heat {
		heatItems = append(heatItems, gin.H{"date": h.Date.Format("2006-01-02"), "minutes": h.Minutes, "sessions": h.Sessions})
	}

	// 2. 雷达图：按 5 大方向计算平均掌握度
	type catRow struct {
		Category string
		AvgMastery float64
	}
	var cats []catRow
	s.services.DB.Model(&model.KnowledgeState{}).
		Select("category, avg(mastery) as avg_mastery").
		Where("user_id = ?", uid).Group("category").Scan(&cats)
	radarItems := make([]gin.H, 0, len(cats))
	for _, c2 := range cats {
		radarItems = append(radarItems, gin.H{"category": c2.Category, "mastery": c2.AvgMastery})
	}

	// 3. 趋势线：最近 7 天每日学习时长
	type trendRow struct {
		Date    time.Time
		Minutes int
	}
	var trends []trendRow
	s.services.DB.Model(&model.DailyStudyTime{}).
		Select("study_date as date, sum(duration_minutes) as minutes").
		Where("user_id = ? AND study_date >= ?", uid, time.Now().AddDate(0, 0, -7)).
		Group("study_date").Order("study_date").Scan(&trends)
	trendItems := make([]gin.H, 0, len(trends))
	for _, t := range trends {
		trendItems = append(trendItems, gin.H{"date": t.Date.Format("2006-01-02"), "minutes": t.Minutes})
	}

	// 4. 统计汇总
	var totalMinutes int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", uid).Select("COALESCE(SUM(duration_minutes),0)").Scan(&totalMinutes)
	var totalSessions int64
	s.services.DB.Model(&model.SessionMessage{}).Where("user_id = ? AND role = ?", uid, "assistant").Count(&totalSessions)
	var totalProblems int64
	s.services.DB.Model(&model.Submission{}).Where("user_id = ? AND passed = ?", uid, true).Count(&totalProblems)
	var totalSub int64
	s.services.DB.Model(&model.Submission{}).Where("user_id = ?", uid).Count(&totalSub)
	var accuracy float64
	if totalSub > 0 {
		accuracy = float64(totalProblems) / float64(totalSub)
	}

	success(c, gin.H{
		"heatmap": heatItems,
		"radar":   radarItems,
		"trend":   trendItems,
		"stats": gin.H{
			"total_minutes":  int(totalMinutes),
			"total_sessions": int(totalSessions),
			"total_problems": int(totalProblems),
			"avg_accuracy":   accuracy,
		},
	})
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
