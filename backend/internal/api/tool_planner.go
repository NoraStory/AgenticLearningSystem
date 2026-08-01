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
	"sync"
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
	"quiz_gen": "generate",
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
		return s.explainSQL(in.Message)
	case "self_heal":
		return selfHeal(in.Message)
	case "resume_review":
		return s.reviewResume(ctx, in)
	case "project_review":
		return s.reviewProject(ctx, in)
	case "quiz_gen":
		return toolExecution{Context: generateQuiz(in.Message), Available: true}, nil
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

func (s *Server) explainSQL(message string) (toolExecution, error) {
	sql := extractSQL(message)
	if sql == "" {
		return toolExecution{Context: "未检测到 SQL。请提供 SELECT/INSERT/UPDATE/DELETE 或 EXPLAIN 语句。", Available: true}, nil
	}
	// 优先真执行计划：EXPLAIN (FORMAT JSON) 只做计划不做查询，无副作用、可安全执行
	if plan := s.explainPlanJSON(sql); plan != "" {
		// 提取执行计划关键指标,供模型与用户判断
		summary := summarizePlan(plan)
		return toolExecution{Context: "SQL 执行计划分析（PostgreSQL EXPLAIN）：\n- " + summary, Available: true}, nil
	}
	// 数据库不可达时回退静态规则
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
	return toolExecution{Context: "SQL 静态分析（数据库不可达时回退）：\n- " + strings.Join(advice, "\n- "), Available: true}, nil
}

// explainPlanJSON 用连接池执行 EXPLAIN (FORMAT JSON)，返回计划 JSON 字符串。
// 只做计划不执行语句，无副作用；出错返回空串由调用方回退。
func (s *Server) explainPlanJSON(sql string) string {
	if s.services == nil || s.services.DB == nil {
		return ""
	}
	var raw string
	// EXPLAIN (FORMAT JSON) 的输出是 json 列,pgx 对 json 类型 Scan 进 string
	// 需要驱动支持;用 QueryRow + 显式 ::text 最稳
	err := s.services.DB.Raw("EXPLAIN (FORMAT JSON) " + sql + "::text").Row().Scan(&raw)
	if err != nil {
		return ""
	}
	return raw
}

// summarizePlan 从 EXPLAIN JSON 中提取人可读的关键指标。
func summarizePlan(plan string) string {
	var parsed []struct {
		Plan struct {
			NodeType       string  `json:"Node Type"`
			RelationName   string  `json:"Relation Name"`
			TotalCost      float64 `json:"Total Cost"`
			StartupCost    float64 `json:"Startup Cost"`
			ActualRows     *int64  `json:"Actual Rows"`
			PlannedRows    float64 `json:"Plan Rows"`
			ScanDirection  string  `json:"Scan Direction"`
			Filter         string  `json:"Filter"`
			IndexName      string  `json:"Index Name"`
			IndexCond      string  `json:"Index Cond"`
			HashCond       string  `json:"Hash Cond"`
			JoinType       string  `json:"Join Type"`
			MergeCond      string  `json:"Merge Cond"`
			Plans          []struct {
				NodeType     string `json:"Node Type"`
				RelationName string `json:"Relation Name"`
				IndexName    string `json:"Index Name"`
				IndexCond    string `json:"Index Cond"`
				Filter       string `json:"Filter"`
			} `json:"Plans"`
		} `json:"Plan"`
	}
	if json.Unmarshal([]byte(plan), &parsed) != nil || len(parsed) == 0 {
		return "执行计划生成成功，但解析失败，请直接查看原始计划。\n" + compact(plan)
	}
	root := parsed[0].Plan
	lines := []string{
		fmt.Sprintf("顶层节点：%s（估算总成本 %.2f，预计返回 %.0f 行）", root.NodeType, root.TotalCost, root.PlannedRows),
	}
	if root.RelationName != "" {
		lines = append(lines, fmt.Sprintf("涉及表：%s", root.RelationName))
	}
	if root.IndexName != "" {
		lines = append(lines, fmt.Sprintf("索引：%s（条件 %s）", root.IndexName, root.IndexCond))
	}
	for _, sub := range root.Plans {
		switch sub.NodeType {
		case "Seq Scan":
			lines = append(lines, fmt.Sprintf("⚠️ 顺序扫描：%s（考虑为查询列建索引；%s）", sub.RelationName, condOr(sub.IndexCond, sub.Filter, "无过滤条件")))
		case "Index Scan":
			lines = append(lines, fmt.Sprintf("✅ 索引扫描：%s on %s", sub.IndexName, sub.RelationName))
		case "Index Only Scan":
			lines = append(lines, fmt.Sprintf("✅ 仅索引扫描：%s on %s", sub.IndexName, sub.RelationName))
		case "Bitmap Heap Scan":
			lines = append(lines, fmt.Sprintf("🔶 位图扫描：%s（%s）", sub.RelationName, condOr(sub.IndexCond, sub.Filter, "")))
		case "Sort":
			lines = append(lines, "🔶 排序节点：结果集需要排序，数据量大时考虑索引避免")
		case "Nested Loop", "Hash Join", "Merge Join":
			lines = append(lines, fmt.Sprintf("🔗 JOIN：%s（%s）", sub.NodeType, condOr(sub.IndexCond, sub.Filter, "")))
		}
	}
	if root.NodeType == "Seq Scan" {
		lines = append(lines, fmt.Sprintf("⚠️ 顶层就是顺序扫描：%s（考虑建索引）", condOr(root.IndexCond, root.Filter, "查询可能全表扫描")))
	}
	return strings.Join(lines, "\n- ")
}

func condOr(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
	// 复用简历模块的真实 LLM 分析（评分/关键词匹配/优缺点/建议），不再输出模板话术
	if s.llm.Enabled() {
		analysis := s.llmAnalyzeResume(text)
		if fb, _ := analysis["fallback"].(bool); !fb {
			score, _ := analysis["score"].(float64)
			ats, _ := analysis["atsScore"].(float64)
			strengths, _ := analysis["strengths"].([]any)
			weaknesses, _ := analysis["weaknesses"].([]any)
			suggestions, _ := analysis["suggestions"].([]any)
			var b strings.Builder
			fmt.Fprintf(&b, "简历分析（AI）：\n- 综合评分 %d/100，ATS 匹配度 %d/100\n", int(score), int(ats))
			b.WriteString("- 优势：")
			for i, s := range strengths {
				if str, ok := s.(string); ok {
					if i > 0 {
						b.WriteString("；")
					}
					b.WriteString(str)
				}
			}
			b.WriteString("\n- 不足：")
			for i, w := range weaknesses {
				if str, ok := w.(string); ok {
					if i > 0 {
						b.WriteString("；")
					}
					b.WriteString(str)
				}
			}
			b.WriteString("\n- 改进建议：")
			for i, s := range suggestions {
				if str, ok := s.(string); ok {
					if i > 0 {
						b.WriteString("；")
					}
					b.WriteString(str)
				}
			}
			return toolExecution{Context: b.String(), Available: true}, nil
		}
	}
	// LLM 不可用或解析失败时回退基础分析
	return toolExecution{Context: fmt.Sprintf("简历初筛：共 %d 字；建议检查联系方式、项目成果量化、技术栈关键词与岗位匹配度。\n%s", len([]rune(text)), compact(text)), Available: true}, nil
}

func (s *Server) reviewProject(ctx context.Context, in agentChatInput) (toolExecution, error) {
	git, _ := gitSummary(ctx, "")
	code, _ := s.searchCode(ctx, in)
	base := "项目审阅：\n" + git.Context + "\n" + code.Context
	// 附件里有代码时，用 LLM 做真实结构点评；无 LLM 时保留静态摘要
	if s.llm.Enabled() && strings.Contains(code.Context, "代码检索结果") {
		answer, err := s.llm.Chat(ctx,
			"你是资深代码评审专家。基于给出的项目代码片段，输出简洁的项目审阅结论：先一句话概述项目做什么，再列出 3-5 条结构/风险/改进建议（每条一行，用 - 开头）。不要输出无关内容。",
			truncate(code.Context, 6000))
		if err == nil && strings.TrimSpace(answer) != "" {
			return toolExecution{Context: base + "\n\n代码评审（AI）：\n" + strings.TrimSpace(answer), Available: true}, nil
		}
	}
	return toolExecution{Context: base, Available: true}, nil
}
func generateQuiz(message string) string {
	topic := compact(strings.TrimSpace(message))
	if topic == "" {
		topic = "当前知识点"
	}
	return fmt.Sprintf("测验草案（%s）：\n1. 请解释核心概念及适用场景。\n2. 给出一个最小代码示例。\n3. 说明一个常见边界条件，并给出改进方案。", topic)
}
type searxResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
		Engine  string `json:"engine"`
	} `json:"results"`
}

// searchCache 是搜索结果的内存缓存：key = provider+query,10 分钟过期。
// 避免同一关键词短时间内重复打 SearXNG。
var searchCache struct {
	mu    sync.Mutex
	items map[string]searchCacheItem
}

type searchCacheItem struct {
	result    string
	expiresAt time.Time
}

func searchCacheGet(key string) (string, bool) {
	searchCache.mu.Lock()
	defer searchCache.mu.Unlock()
	item, ok := searchCache.items[key]
	if !ok {
		return "", false
	}
	if time.Now().After(item.expiresAt) {
		delete(searchCache.items, key)
		return "", false
	}
	return item.result, true
}

func searchCacheSet(key, result string) {
	searchCache.mu.Lock()
	defer searchCache.mu.Unlock()
	if searchCache.items == nil {
		searchCache.items = map[string]searchCacheItem{}
	}
	searchCache.items[key] = searchCacheItem{result: result, expiresAt: time.Now().Add(10 * time.Minute)}
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
	cacheKey := provider + ":" + q
	if cached, ok := searchCacheGet(cacheKey); ok {
		return cached, nil
	}
	var result string
	var err error
	switch provider {
	case "searxng":
		result, err = s.searxSearch(ctx, q)
		if err == nil {
			searchCacheSet(cacheKey, result)
			return result, nil
		}
		// SearXNG 失败后用 Wikipedia 回退（用独立 context，不受原超时影响）
		if fallback, fallbackErr := wikiSearch(context.Background(), q); fallbackErr == nil {
			searchCacheSet(cacheKey, fallback)
			return fallback, nil
		}
		return "", fmt.Errorf("SearXNG 搜索失败：%w", err)
	case "duckduckgo":
		result, err = duckSearch(ctx, q)
	default:
		return "", fmt.Errorf("不支持的搜索提供商：%s", provider)
	}
	if err == nil {
		searchCacheSet(cacheKey, result)
	}
	return result, err
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
		return "", err
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
