package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"codeforge/backend/internal/model"
	"codeforge/backend/internal/sandbox"
)

type plannedTool struct {
	Meta      toolMeta
	Reason    string
	Phase     string
	DependsOn []string
}

type toolExecution struct {
	Context   string
	Images    []string
	Available bool
}

var toolPhases = map[string]string{
	"web_search": "retrieve", "doc_reader": "retrieve", "code_search": "retrieve", "git_helper": "retrieve",
	"leetcode_fetch": "retrieve", "course_search": "retrieve", "progress_query": "retrieve",
	"code_execute": "validate", "sql_explain": "analyze", "self_heal": "analyze", "resume_review": "analyze", "project_review": "analyze",
	"diagram_gen": "generate", "quiz_gen": "generate", "mindmap_gen": "generate",
}

// toolCapability reports executor registration. Runtime dependencies such as network,
// a compiler, or a database are reported in tool_result when the tool is called.
func toolCapability(id string) (bool, string) {
	switch id {
	case "web_search":
		return true, "SearXNG 本地搜索已接入，服务未启动时回退 Wikipedia"
	case "code_execute":
		return true, "受限沙箱执行器已接入；语言运行时缺失时会返回 unavailable"
	case "doc_reader":
		return true, "支持图片、UTF-8 文本、Markdown、JSON、CSV 和代码附件"
	default:
		return true, "已实现真实执行器，可由动态工作流调度"
	}
}

func (s *Server) planAgentTools(uid string, in agentChatInput, history []model.SessionMessage) []plannedTool {
	enabled := map[string]bool{}
	for _, t := range toolCatalog {
		enabled[t.ID] = true
	}
	var settings []model.UserToolSetting
	s.services.DB.Where("user_id = ?", uid).Find(&settings)
	for _, setting := range settings {
		enabled[setting.ToolID] = setting.Enabled
	}
	for _, t := range toolCatalog {
		if t.Locked {
			enabled[t.ID] = true
		}
	}

	query := strings.ToLower(in.Message)
	page, _ := in.Context["current_page"].(string)
	selected := []plannedTool{}
	seen := map[string]bool{}
	add := func(id, reason string) {
		if seen[id] || !enabled[id] {
			return
		}
		for _, t := range toolCatalog {
			if t.ID == id {
				selected = append(selected, plannedTool{Meta: t, Reason: reason, Phase: toolPhases[id]})
				seen[id] = true
				return
			}
		}
	}
	// A short follow-up should reuse conversation memory and go straight to the answer.
	if len(history) > 0 && isContextualFollowUp(in.Message) && len(in.Attachments) == 0 && !containsAny(query,
		"\u4ee3\u7801", "bug", "sql", "\u6700\u65b0", "\u4eca\u5929", "\u65b0\u95fb", "\u8054\u7f51", "\u641c\u7d22", "\u8fd0\u884c", "\u6267\u884c", "\u4fee\u590d", "\u9898", "\u8bfe\u7a0b", "\u6559\u7a0b", "\u5b66\u4e60\u8d44\u6599", "\u7b80\u5386", "\u6c42\u804c", "\u9879\u76ee", "\u6d41\u7a0b\u56fe", "\u67b6\u6784\u56fe", "\u6d4b\u9a8c", "\u51fa\u9898", "\u601d\u7ef4\u5bfc\u56fe", "git", "commit", "quiz") {
		return selected
	}
	if len(in.Attachments) > 0 {
		add("doc_reader", "检测到图片或文件附件，先读取可识别内容")
	}
	if containsAny(query, "最新", "今天", "新闻", "联网", "搜索网页", "查资料", "look up", "search web", "实时", "当前", "现在", "2024", "2025", "2026", "更新", "变化", "动态", "热点", "发布", "版本", "release", "最新版", "最新动态", "资讯", "行情", "趋势", "上线", "公告", "什么是", "什么是", "介绍一下", "了解", "对比", "区别", "推荐", "best", "latest", "current", "news", "update", "version", "trend") {
		add("web_search", "问题需要联网获取最新资料")
	}
	if containsAny(query, "代码", "报错", "bug", "异常", "编译", "函数", "class ", "def ") {
		add("code_search", "问题涉及代码或错误定位")
	}
	if containsAny(query, "运行代码", "执行代码", "测试代码", "run code") {
		add("code_execute", "用户明确要求运行或测试代码")
	}
	if containsAny(query, "修复代码", "自动修复", "修 bug", "self heal") {
		if _, _, ok := extractCode(in.Message); ok {
			add("code_execute", "先运行代码收集可复现的错误信息")
		}
		add("self_heal", "用户要求分析并修复代码")
	}
	if containsAny(query, "leetcode", "力扣", "算法题", "题目") {
		add("leetcode_fetch", "问题涉及算法题或题目检索")
	}
	if containsAny(query, "sql", "数据库查询", "执行计划", "索引优化") {
		add("sql_explain", "问题涉及 SQL 或查询优化")
	}
	if containsAny(query, "课程", "教程", "学习资料", "知识点") || strings.Contains(page, "course") {
		add("course_search", "需要检索站内学习内容")
	}
	if containsAny(query, "学习进度", "完成度", "学习时长", "连续学习") || strings.Contains(page, "profile") {
		add("progress_query", "需要读取当前用户学习数据")
	}
	if containsAny(query, "简历", "求职", "ats") {
		add("resume_review", "问题涉及简历或求职")
	}
	if containsAny(query, "项目评审", "项目分析", "源码项目", "项目审阅") {
		add("project_review", "问题涉及项目评审")
	}
	if containsAny(query, "架构图", "流程图", "mermaid", "画图") {
		add("diagram_gen", "用户需要图表或流程图")
	}
	if containsAny(query, "测验", "出题", "练习题", "quiz") {
		add("quiz_gen", "用户需要生成测验")
	}
	if containsAny(query, "思维导图", "mindmap") {
		add("mindmap_gen", "用户需要思维导图")
	}
	if containsAny(query, "git", "提交信息", "commit", "分支", "工作区状态") {
		add("git_helper", "问题涉及 Git 或工作区变更")
	}
	// 关键词未命中时，用 LLM 判断是否需要联网搜索
	if !seen["web_search"] && s.llm != nil && s.llm.Enabled() {
		decision := s.llmShouldSearch(in.Message, history)
		if decision {
			add("web_search", "AI 判断该问题需要联网获取外部信息")
		}
	}
	return scheduleTools(selected)
}

// Keep execution sequential for deterministic SSE ordering, but expose phases and dependencies.
func scheduleTools(tools []plannedTool) []plannedTool {
	for i := range tools {
		if tools[i].Meta.ID == "self_heal" {
			for _, other := range tools {
				if other.Meta.ID == "code_execute" {
					tools[i].DependsOn = []string{"code_execute"}
				}
			}
		}
		if tools[i].Meta.ID == "resume_review" || tools[i].Meta.ID == "project_review" {
			for _, other := range tools {
				if other.Meta.ID == "doc_reader" {
					tools[i].DependsOn = append(tools[i].DependsOn, "doc_reader")
				}
			}
		}
	}
	phaseOrder := map[string]int{"retrieve": 0, "validate": 1, "analyze": 2, "generate": 3}
	sort.SliceStable(tools, func(i, j int) bool { return phaseOrder[tools[i].Phase] < phaseOrder[tools[j].Phase] })
	return tools
}

func containsAny(text string, values ...string) bool {
	for _, v := range values {
		if strings.Contains(text, strings.ToLower(v)) {
			return true
		}
	}
	return false
}

func (s *Server) executeAgentTool(ctx context.Context, uid string, plan plannedTool, in agentChatInput) (toolExecution, error) {
	switch plan.Meta.ID {
	case "doc_reader":
		return s.readAttachments(ctx, in.Attachments)
	case "course_search":
		var courses []model.Course
		q := strings.TrimSpace(in.Message)
		if len([]rune(q)) > 30 {
			q = string([]rune(q)[:30])
		}
		like := "%" + q + "%"
		s.services.DB.Where("title ILIKE ? OR summary ILIKE ?", like, like).Order("views desc").Limit(5).Find(&courses)
		if len(courses) == 0 {
			s.services.DB.Order("views desc").Limit(5).Find(&courses)
		}
		names := []string{}
		for _, course := range courses {
			names = append(names, course.Title)
		}
		if len(names) == 0 {
			return toolExecution{Context: "站内暂无匹配课程。", Available: true}, nil
		}
		return toolExecution{Context: "站内课程：" + strings.Join(names, "、"), Available: true}, nil
	case "progress_query":
		var total, completed int64
		s.services.DB.Model(&model.Course{}).Count(&total)
		s.services.DB.Model(&model.LearningProgress{}).Where("user_id = ? AND progress >= 100", uid).Count(&completed)
		return toolExecution{Context: fmt.Sprintf("当前用户课程完成 %d/%d。", completed, total), Available: true}, nil
	case "web_search":
		result, err := s.searchWeb(ctx, in.Message)
		if err != nil {
			return toolExecution{Context: "联网搜索失败：" + err.Error(), Available: false}, nil
		}
		return toolExecution{Context: result, Available: true}, nil
	case "leetcode_fetch":
		return s.searchProblems(in.Message)
	case "code_search":
		return s.searchCode(ctx, in)
	case "git_helper":
		return gitSummary(ctx, in.Message)
	case "code_execute":
		return executeCode(in.Message)
	case "sql_explain":
		return explainSQL(in.Message)
	case "self_heal":
		return selfHeal(in.Message)
	case "resume_review":
		return s.reviewResume(ctx, in)
	case "project_review":
		return s.reviewProject(ctx, in)
	case "diagram_gen":
		return toolExecution{Context: generateDiagram(in.Message), Available: true}, nil
	case "quiz_gen":
		return toolExecution{Context: generateQuiz(in.Message), Available: true}, nil
	case "mindmap_gen":
		return toolExecution{Context: generateMindmap(in.Message), Available: true}, nil
	default:
		return toolExecution{Context: plan.Meta.Name + "执行器未注册。", Available: false}, nil
	}
}

func (s *Server) readAttachments(ctx context.Context, keys []string) (toolExecution, error) {
	images, texts, unsupported := []string{}, []string{}, []string{}
	for _, key := range keys {
		if dataURL, err := s.store.ImageDataURL(ctx, key, 12<<20); err == nil {
			images = append(images, dataURL)
			continue
		}
		reader, err := s.store.Open(ctx, key)
		if err != nil {
			unsupported = append(unsupported, filepath.Base(key)+"（无法读取）")
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, 256<<10))
		_ = reader.Close()
		if readErr != nil {
			unsupported = append(unsupported, filepath.Base(key)+"（读取失败）")
			continue
		}
		if !isTextAttachment(key, data) {
			unsupported = append(unsupported, filepath.Base(key)+"（暂不支持该格式）")
			continue
		}
		if text := strings.TrimSpace(string(data)); text != "" {
			texts = append(texts, fmt.Sprintf("[%s]\n%s", filepath.Base(key), text))
		}
	}
	parts := []string{fmt.Sprintf("已读取附件：图片 %d 个，文本 %d 个。", len(images), len(texts))}
	if len(texts) > 0 {
		parts = append(parts, strings.Join(texts, "\n\n"))
	}
	if len(unsupported) > 0 {
		parts = append(parts, "未解析格式："+strings.Join(unsupported, "、"))
	}
	return toolExecution{Context: strings.Join(parts, "\n"), Images: images, Available: true}, nil
}

func isTextAttachment(key string, data []byte) bool {
	ext := strings.ToLower(filepath.Ext(key))
	if strings.Contains(string(data[:minInt(len(data), 512)]), "\x00") {
		return false
	}
	for _, allowed := range []string{".txt", ".md", ".markdown", ".json", ".csv", ".log", ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".java", ".c", ".cpp", ".h", ".sql", ".yaml", ".yml", ".toml", ".xml", ".html", ".css"} {
		if ext == allowed {
			return true
		}
	}
	return len(data) > 0 && strings.TrimSpace(string(data)) != ""
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var fencedCodePattern = regexp.MustCompile("(?s)```([A-Za-z0-9_+#.-]*)\\s*\\r?\\n?(.*?)```")

func extractCode(message string) (language, code string, ok bool) {
	match := fencedCodePattern.FindStringSubmatch(message)
	if len(match) != 3 {
		return "", "", false
	}
	language = strings.TrimSpace(strings.ToLower(match[1]))
	if language == "" {
		language = inferLanguage(match[2])
	}
	code = strings.TrimSpace(match[2])
	return language, code, code != ""
}
func inferLanguage(code string) string {
	lower := strings.ToLower(code)
	switch {
	case strings.Contains(lower, "#include") || strings.Contains(lower, "std::"):
		return "cpp"
	case strings.Contains(lower, "console.log") || strings.Contains(lower, "function "):
		return "javascript"
	case strings.Contains(lower, "fn main") || strings.Contains(lower, "let mut "):
		return "rust"
	default:
		return "python"
	}
}
func executeCode(message string) (toolExecution, error) {
	language, code, ok := extractCode(message)
	if !ok {
		return toolExecution{Context: "已启用代码沙箱，但当前消息没有可执行的 Markdown 代码块。", Available: true}, nil
	}
	result, err := sandbox.Run(language, code)
	if err != nil {
		return toolExecution{}, err
	}
	payload, _ := json.Marshal(result)
	return toolExecution{Context: fmt.Sprintf("代码执行结果（%s）：%s", language, payload), Available: result.Status != "unavailable" && result.Status != "unsupported"}, nil
}
func selfHeal(message string) (toolExecution, error) {
	language, code, ok := extractCode(message)
	if !ok {
		return toolExecution{Context: "请提供带语言标记的 Markdown 代码块，我才能复现错误并给出最小修复。", Available: true}, nil
	}
	result, err := sandbox.Run(language, code)
	if err != nil {
		return toolExecution{}, err
	}
	if result.Status == "success" {
		return toolExecution{Context: "代码复现通过，未发现运行时错误；请继续提供失败用例或编译日志。", Available: true}, nil
	}
	return toolExecution{Context: fmt.Sprintf("已复现 %s；建议先根据 stderr/stdout 修复，再重新运行。诊断：%s", result.Status, compact(result.Stderr+" "+result.Stdout)), Available: true}, nil
}
func compact(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1200 {
		return s[:1200] + "…"
	}
	return s
}

func explainSQL(message string) (toolExecution, error) {
	sql := extractSQL(message)
	if sql == "" {
		return toolExecution{Context: "未检测到 SQL。请提供 SELECT/INSERT/UPDATE/DELETE 或 EXPLAIN 语句。", Available: true}, nil
	}
	upper := strings.ToUpper(sql)
	advice := []string{}
	if strings.Contains(upper, "SELECT *") {
		advice = append(advice, "避免 SELECT *，只选择需要的列")
	}
	if strings.Contains(upper, " LIKE '%") {
		advice = append(advice, "前缀为通配符的 LIKE 通常无法使用普通 B-tree 索引")
	}
	if strings.Contains(upper, " OR ") {
		advice = append(advice, "多个 OR 条件可评估 UNION ALL 或组合索引")
	}
	if !strings.Contains(upper, " WHERE ") && strings.HasPrefix(strings.TrimSpace(upper), "SELECT") {
		advice = append(advice, "缺少 WHERE 条件，确认是否会扫描整张表")
	}
	if strings.Contains(upper, "JOIN") && !strings.Contains(upper, " ON ") {
		advice = append(advice, "JOIN 缺少显式 ON 条件，请确认是否误用了笛卡尔积")
	}
	if len(advice) == 0 {
		advice = append(advice, "未发现明显的静态反模式；生产环境仍应使用目标数据库的 EXPLAIN ANALYZE 验证")
	}
	return toolExecution{Context: "SQL 静态分析：\n- " + strings.Join(advice, "\n- "), Available: true}, nil
}
func extractSQL(message string) string {
	for _, match := range fencedCodePattern.FindAllStringSubmatch(message, -1) {
		if strings.Contains(strings.ToLower(match[1]), "sql") {
			return strings.TrimSpace(match[2])
		}
	}
	for _, keyword := range []string{"SELECT ", "INSERT ", "UPDATE ", "DELETE ", "EXPLAIN "} {
		if i := strings.Index(strings.ToUpper(message), keyword); i >= 0 {
			return strings.TrimSpace(strings.Trim(message[i:], "`\n "))
		}
	}
	return ""
}

func (s *Server) searchProblems(message string) (toolExecution, error) {
	q := strings.TrimSpace(message)
	if len([]rune(q)) > 40 {
		q = string([]rune(q)[:40])
	}
	like := "%" + q + "%"
	var rows []model.Problem
	s.services.DB.Where("title ILIKE ? OR description ILIKE ? OR category ILIKE ?", like, like, like).Order("id asc").Limit(5).Find(&rows)
	if len(rows) == 0 {
		return toolExecution{Context: "站内题库未找到匹配题目；可以直接提供题号或题目标题，我会继续分析。", Available: true}, nil
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, fmt.Sprintf("#%d %s（%s）", row.ID, row.Title, row.Difficulty))
	}
	return toolExecution{Context: "站内算法题：" + strings.Join(items, "、"), Available: true}, nil
}

func (s *Server) searchCode(ctx context.Context, in agentChatInput) (toolExecution, error) {
	terms := codeSearchTerms(in.Message)
	if len(terms) == 0 {
		return toolExecution{Context: "未提取到明确代码符号；请提供函数名、类名或错误标识。", Available: true}, nil
	}
	matches := []string{}
	for _, key := range in.Attachments {
		reader, err := s.store.Open(ctx, key)
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(io.LimitReader(reader, 512<<10))
		_ = reader.Close()
		if !isTextAttachment(key, data) {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			lower := strings.ToLower(line)
			for _, term := range terms {
				if strings.Contains(lower, strings.ToLower(term)) {
					start, end := i-1, i+1
					if start < 0 {
						start = 0
					}
					if end >= len(lines) {
						end = len(lines) - 1
					}
					matches = append(matches, fmt.Sprintf("%s:%d\n%s", filepath.Base(key), i+1, strings.TrimSpace(strings.Join(lines[start:end+1], " "))))
					break
				}
			}
		}
	}
	if len(matches) == 0 {
		return toolExecution{Context: "未在附件中找到匹配符号。若要检索项目源码，请上传文件或配置 CODEFORGE_WORKSPACE。", Available: true}, nil
	}
	if len(matches) > 8 {
		matches = matches[:8]
	}
	return toolExecution{Context: "代码检索结果：\n- " + strings.Join(matches, "\n- "), Available: true}, nil
}
func codeSearchTerms(message string) []string {
	seen := map[string]bool{}
	terms := []string{}
	for _, term := range regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]{2,}`).FindAllString(message, -1) {
		lower := strings.ToLower(term)
		if !seen[lower] && !containsAny(lower, "the", "and", "code", "error", "please") {
			seen[lower] = true
			terms = append(terms, term)
		}
	}
	return terms
}

func gitSummary(ctx context.Context, message string) (toolExecution, error) {
	root := os.Getenv("CODEFORGE_WORKSPACE")
	if root == "" {
		root, _ = os.Getwd()
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		return toolExecution{Context: "未找到 Git 工作区；设置 CODEFORGE_WORKSPACE 后可读取 status/diff。", Available: true}, nil
	}
	status, err := runGit(ctx, root, "status", "--short")
	if err != nil {
		return toolExecution{Context: "Git 状态读取失败：" + err.Error(), Available: false}, nil
	}
	stat, _ := runGit(ctx, root, "diff", "--stat")
	result := "Git 工作区状态：clean"
	if strings.TrimSpace(status) != "" {
		result = "Git 工作区状态：\n" + compact(status)
	}
	if strings.TrimSpace(stat) != "" {
		result += "\n变更统计：\n" + compact(stat)
	}
	if containsAny(strings.ToLower(message), "commit", "提交", "提交信息") {
		result += "\n建议提交信息：" + suggestCommit(status)
	}
	return toolExecution{Context: result, Available: true}, nil
}
func runGit(ctx context.Context, root string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
func suggestCommit(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "test"):
		return "test: update automated tests"
	case strings.Contains(lower, "fix"):
		return "fix: resolve reported issue"
	case strings.TrimSpace(status) != "":
		return "chore: update learning project"
	default:
		return "chore: update project"
	}
}

func (s *Server) reviewResume(ctx context.Context, in agentChatInput) (toolExecution, error) {
	reader, _ := s.readAttachments(ctx, in.Attachments)
	text := reader.Context
	if len(in.Attachments) == 0 {
		text = in.Message
	}
	if strings.TrimSpace(text) == "" {
		return toolExecution{Context: "请上传简历文本或粘贴简历内容。", Available: true}, nil
	}
	return toolExecution{Context: fmt.Sprintf("简历初筛：共 %d 字；建议检查联系方式、项目成果量化、技术栈关键词与岗位匹配度。\n%s", len([]rune(text)), compact(text)), Available: true}, nil
}
func (s *Server) reviewProject(ctx context.Context, in agentChatInput) (toolExecution, error) {
	git, _ := gitSummary(ctx, "")
	code, _ := s.searchCode(ctx, in)
	return toolExecution{Context: "项目审阅：\n" + git.Context + "\n" + code.Context, Available: true}, nil
}
func generateDiagram(message string) string {
	topic := compact(strings.TrimSpace(message))
	lower := strings.ToLower(topic)
	if containsAny(lower, "\u524d\u7f00\u548c", "prefix sum", "scan", "hillis-steele", "\u53cc\u7f13\u51b2") {
		return "Mermaid \u6d41\u7a0b\u56fe\uff1a\n```mermaid\nflowchart TD\n  A[\u521d\u59cb\u53168\u5143\u7d20\u6570\u7ec4\uff1a1 2 3 4 5 6 7 8] --> B[\u8bbe\u7f6e step = 1\uff0c\u6e90\u7f13\u51b2\u533a\u53d1\u76ee\u6807\u7f13\u51b2\u533a]\n  B --> C{step < n ?}\n  C -->|\u662f| D[\u6bcf\u4e2a\u7ebf\u7a0b\u8bfb\u53d6 src[i] \u548c src[i-step]]\n  D --> E[\u5199\u5165 dst[i]\uff1a\u6709\u524d\u7f6e\u5143\u7d20\u5219\u76f8\u52a0\uff0c\u5426\u5219\u76f4\u63a5\u590d\u5236]\n  E --> F[\u540c\u6b65\u5e76\u4ea4\u6362 src \u548c dst \u53cc\u7f13\u51b2\u533a]\n  F --> G[step = step * 2]\n  G --> C\n  C -->|\u5426| H[\u8f93\u51fa\u524d\u7f00\u548c\u7ed3\u679c]\n```"
	}
	if topic == "" {
		topic = "\u5b66\u4e60\u6d41\u7a0b"
	}
	return fmt.Sprintf("Mermaid \u6d41\u7a0b\u56fe\uff1a\n```mermaid\nflowchart TD\n  A[\u5f00\u59cb\uff1a%s] --> B[\u5206\u6790\u9700\u6c42]\n  B --> C[\u6267\u884c\u5de5\u5177]\n  C --> D[\u751f\u6210\u7ed3\u679c]\n```", topic)
}
func generateQuiz(message string) string {
	topic := compact(strings.TrimSpace(message))
	if topic == "" {
		topic = "当前知识点"
	}
	return fmt.Sprintf("测验草案（%s）：\n1. 请解释核心概念及适用场景。\n2. 给出一个最小代码示例。\n3. 说明一个常见边界条件，并给出改进方案。", topic)
}
func generateMindmap(message string) string {
	topic := compact(strings.TrimSpace(message))
	if topic == "" {
		topic = "学习主题"
	}
	return fmt.Sprintf("Mermaid 思维导图：\n```mermaid\nmindmap\n  root((%s))\n    核心概念\n    示例\n    常见错误\n    练习与复盘\n```", topic)
}

type searxResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
		Engine  string `json:"engine"`
	} `json:"results"`
}

// searchWeb selects the configured provider. SearXNG is the default because it
// can run locally and expose a stable JSON endpoint without a vendor API key.
func (s *Server) searchWeb(ctx context.Context, q string) (string, error) {
	q = extractSearchKeywords(q)
	if q == "" {
		return "没有搜索关键词。", nil
	}
	provider := strings.ToLower(strings.TrimSpace(s.cfg.SearchProvider))
	if provider == "" {
		provider = "searxng"
	}
	switch provider {
	case "searxng":
		result, err := s.searxSearch(ctx, q)
		if err == nil {
			return result, nil
		}
		// SearXNG 失败后用 Wikipedia 回退（用独立 context，不受原超时影响）
		if fallback, fallbackErr := wikiSearch(context.Background(), q); fallbackErr == nil {
			return fallback, nil
		}
		return "", fmt.Errorf("SearXNG 搜索失败：%w", err)
	case "duckduckgo":
		return duckSearch(ctx, q)
	default:
		return "", fmt.Errorf("不支持的搜索提供商：%s", provider)
	}
}

func (s *Server) searxSearch(ctx context.Context, q string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(s.cfg.SearchBaseURL), "/")
	if base == "" {
		return "", fmt.Errorf("SEARCH_BASE_URL \u4e3a\u7a7a")
	}
	u, err := url.Parse(base + "/search")
	if err != nil {
		return "", err
	}
	values := u.Query()
	values.Set("q", q)
	values.Set("format", "json")
	values.Set("language", "all")
	values.Set("categories", "general")
	u.RawQuery = values.Encode()

	timeout := s.cfg.SearchTimeoutSeconds
	if timeout < 1 || timeout < 15 {
		timeout = 15
	}
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CodeForge-Academy/1.0")
	res, err := (&http.Client{Timeout: time.Duration(timeout+2) * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}
	var payload searxResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Results) == 0 {
		return "", fmt.Errorf("SearXNG \u641c\u7d22\u6ca1\u6709\u8fd4\u56de\u7ed3\u679c")
	}
	parts := make([]string, 0, minInt(len(payload.Results), 8))
	for _, item := range payload.Results {
		if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.URL) == "" {
			continue
		}
		line := item.Title + "?" + item.URL + "?"
		if strings.TrimSpace(item.Content) != "" {
			line += "?" + compact(item.Content)
		}
		parts = append(parts, line)
		if len(parts) >= 8 {
			break
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("SearXNG \u641c\u7d22\u6ca1\u6709\u8fd4\u56de\u7ed3\u679c")
	}
	return "\u8054\u7f51\u641c\u7d22\uff08SearXNG\uff09\uff1a\n- " + strings.Join(parts, "\n- "), nil
}

type duckResponse struct {
	AbstractText  string `json:"AbstractText"`
	AbstractURL   string `json:"AbstractURL"`
	RelatedTopics []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}

func duckSearch(ctx context.Context, q string) (string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "没有搜索关键词。", nil
	}
	u := "https://api.duckduckgo.com/?q=" + url.QueryEscape(q) + "&format=json&no_html=1&no_redirect=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	client := http.Client{Timeout: 8 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		if fallback, fallbackErr := wikiSearch(searchFallbackContext(), q); fallbackErr == nil {
			return fallback, nil
		}
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}
	var out duckResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	parts := []string{}
	if out.AbstractText != "" {
		part := out.AbstractText
		if out.AbstractURL != "" {
			part += "（" + out.AbstractURL + "）"
		}
		parts = append(parts, part)
	}
	for i, t := range out.RelatedTopics {
		if i >= 5 {
			break
		}
		if t.Text != "" {
			if t.FirstURL != "" {
				parts = append(parts, t.Text+"（"+t.FirstURL+"）")
			} else {
				parts = append(parts, t.Text)
			}
		}
	}
	if len(parts) == 0 {
		if fallback, fallbackErr := wikiSearch(searchFallbackContext(), q); fallbackErr == nil {
			return fallback, nil
		}
		return "搜索服务没有返回摘要；模型将基于已有知识回答。", nil
	}
	return "联网搜索摘要：\n- " + strings.Join(parts, "\n- "), nil
}

type wikiResponse struct {
	Query struct {
		Search []struct {
			Title   string `json:"title"`
			Snippet string `json:"snippet"`
		} `json:"search"`
	} `json:"query"`
}

func searchFallbackContext() context.Context {
	return context.Background()
}

func wikiSearch(ctx context.Context, q string) (string, error) {
	u := "https://en.wikipedia.org/w/api.php?action=query&list=search&srsearch=" + url.QueryEscape(q) + "&format=json&utf8=1&srlimit=5"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	res, err := (&http.Client{Timeout: 6 * time.Second}).Do(req)
	var out wikiResponse
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode < 300 {
			err = json.NewDecoder(res.Body).Decode(&out)
		} else {
			err = fmt.Errorf("Wikipedia HTTP %d", res.StatusCode)
		}
	}
	if err != nil {
		// Some Windows environments block Go's TLS transport while curl.exe can
		// still reach the same public endpoint. Keep this fallback constrained to
		// the fixed Wikipedia URL and a short timeout.
		cmd := exec.CommandContext(ctx, "curl.exe", "-fsSL", "--max-time", "6", u)
		data, curlErr := cmd.Output()
		if curlErr != nil {
			cmd = exec.CommandContext(ctx, "curl", "-fsSL", "--max-time", "6", u)
			data, curlErr = cmd.Output()
		}
		if curlErr != nil {
			return "", err
		}
		if decodeErr := json.Unmarshal(data, &out); decodeErr != nil {
			return "", decodeErr
		}
	}
	parts := []string{}
	for _, item := range out.Query.Search {
		snippet := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(item.Snippet, "")
		parts = append(parts, html.UnescapeString(item.Title+"："+snippet+"（https://en.wikipedia.org/wiki/"+url.PathEscape(strings.ReplaceAll(item.Title, " ", "_"))+"）"))
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("Wikipedia 没有返回结果")
	}
	return "联网搜索（Wikipedia 回退）：\n- " + strings.Join(parts, "\n- "), nil
}

// llmShouldSearch 用 LLM 快速判断当前问题是否需要联网搜索。
// 仅在关键词规则未命中时调用，避免每次请求都消耗 LLM 额度。
func (s *Server) llmShouldSearch(message string, history []model.SessionMessage) bool {
	systemPrompt := "你是一个搜索意图判断器。判断用户的问题是否需要联网搜索才能回答（例如：需要最新信息、实时数据、新闻、技术版本、具体事实查证、产品对比等）。只回答 JSON：{\"need_search\": true} 或 {\"need_search\": false}。不要回答其他内容。"
	userPrompt := "问题：" + message + "\n\n请判断是否需要联网搜索。"
	answer, err := s.llm.Chat(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return strings.Contains(answer, "true")
}
