package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
)

// ---- 简历文本提取 ----

// extractResumeText 按文件扩展名提取简历文本:PDF 用 ledongthuc/pdf,DOCX 用 Python 子进程,txt 直读,DOC 不支持。
func (s *Server) extractResumeText(c *gin.Context, key, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	rc, err := s.store.Open(c, key)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	switch ext {
	case ".pdf":
		return s.extractPDF(rc)
	case ".docx":
		return s.extractDocxText(c, rc)
	case ".txt":
		raw, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	default:
		return "", fmt.Errorf("暂不支持 %s 格式,请上传 PDF、DOCX 或 TXT", ext)
	}
}

// extractPDF 落临时文件后用 ledongthuc/pdf 提取全文。
func (s *Server) extractPDF(rc io.Reader) (string, error) {
	tmp, err := os.CreateTemp("", "codeforge-resume-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, rc); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	f, reader, err := pdf.Open(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("PDF 解析失败: %w", err)
	}
	defer f.Close()
	tr, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("PDF 文本提取失败: %w", err)
	}
	raw, err := io.ReadAll(tr)
	if err != nil {
		return "", err
	}
	text := string(raw)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("PDF 未提取到文本,可能是扫描件,暂不支持 OCR")
	}
	return text, nil
}

// extractDocxText 复用 tpl_parse.py 解析 DOCX,把段落拼接成文本。
func (s *Server) extractDocxText(c *gin.Context, rc io.Reader) (string, error) {
	tmp, err := os.CreateTemp("", "codeforge-resume-*.docx")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, rc); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()

	parsed, err := s.parseDocxTemplate(context.Background(), tmp.Name())
	if err != nil {
		return "", fmt.Errorf("DOCX 解析失败: %w", err)
	}
	var b strings.Builder
	for _, p := range s.parseParagraphs(parsed) {
		b.WriteString(p.Text)
		b.WriteString("\n")
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", fmt.Errorf("DOCX 未提取到文本")
	}
	return b.String(), nil
}

// ---- LLM 分析 ----

// looksLikeResume 粗略判断文本是否像简历(有姓名/联系方式/经历等特征)。
func looksLikeResume(text string) bool {
	if len(text) > 20000 {
		return false // 简历一般不会这么长,多半是教程/文档
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\bemail\b|@\w+\.\w+|\bphone\b|电话|手机|邮箱)`),
		regexp.MustCompile(`(?i)(工作经历|项目经验|教育背景|技能清单|实习经历|experience|education|skill)`),
		regexp.MustCompile(`(?i)(\bgithub\.com\b|\blinkedin\b|简历|resume|cv)`),
	}
	hits := 0
	for _, re := range patterns {
		if re.MatchString(text) {
			hits++
		}
	}
	return hits >= 2
}

// llmAnalyzeResume 用 LLM 分析简历文本,输出与前端 AnalysisResult 兼容的结构。LLM 不可用/解析失败/内容不像简历时返回基础结构。
func (s *Server) llmAnalyzeResume(text string) gin.H {
	fallback := gin.H{
		"score":        0,
		"atsScore":     0,
		"keywordMatch": []gin.H{},
		"strengths":    []string{},
		"weaknesses":   []string{},
		"suggestions":  []string{},
		"fallback":     true,
	}
	if !s.llm.Enabled() || strings.TrimSpace(text) == "" {
		return fallback
	}
	if !looksLikeResume(text) {
		fallback["message"] = "上传的文件看起来不是简历(缺少姓名/联系方式/经历等特征),请上传 PDF、DOCX 或 TXT 格式的个人简历。"
		return fallback
	}
	systemPrompt := "你是资深 HR 与 ATS 简历分析师。请分析简历并输出 JSON,结构为:{\"score\":0-100,\"atsScore\":0-100,\"keywordMatch\":[{\"keyword\":\"...\",\"found\":true/false}],\"strengths\":[\"...\"],\"weaknesses\":[\"...\"],\"suggestions\":[\"...\"]}。要求:分数基于简历真实内容;keywordMatch 覆盖常见技术关键词并标注是否命中;strengths/weaknesses/suggestions 各 3-5 条,具体且可操作。只返回 JSON。"
	answer, err := s.llm.Chat(context.Background(), systemPrompt, text)
	if err != nil {
		fmt.Printf("[LLM-ANALYZE-ERR] %v\n", err)
		return fallback
	}
	answer = cleanJSON(answer)
	var out gin.H
	if json.Unmarshal([]byte(answer), &out) == nil {
		out["fallback"] = false
		return out
	}
	fmt.Printf("[LLM-ANALYZE-PARSE-ERR] len=%d prefix=%q\n", len(answer), truncate(answer, 300))
	return fallback
}

// truncate 安全截取字符串(避免中文字符截断)。
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// cleanJSON 去掉 LLM 输出中的 ```json 包裹。
func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// ---- diff 式优化(Resume-Matcher 思路) ----

// resumeChange 是 LLM 返回的单条修改点。
type resumeChange struct {
	Path     string `json:"path"` // 形如 sections.0.items.0
	Action   string `json:"action"` // rewrite | insert | delete
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
}

// lockedFieldPatterns 禁止 LLM 修改的个人信息字段(防捏造)。
var lockedFieldPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(姓名|name|邮箱|email|电话|phone|手机|linkedin|github\.com|出生|生日|日期|date)`),
	regexp.MustCompile(`\d{4}[.\-/年]\d{1,2}`), // 日期:2022.03 / 2022-03 / 2022年3月
}

// applyChanges 把 LLM 返回的修改点应用到文本行列表,逐条校验,命中锁定字段/路径非法/action 非法的条目静默丢弃。
func applyChanges(lines []string, changes []resumeChange) ([]string, int) {
	dropped := 0
	valid := make([]resumeChange, 0, len(changes))
	for _, ch := range changes {
		idx, ok := parsePath(ch.Path)
		if !ok || idx < 0 || idx >= len(lines) {
			dropped++
			continue
		}
		if locked := containsLocked(lines[idx]); locked {
			dropped++
			continue
		}
		switch ch.Action {
		case "rewrite", "insert", "delete":
			valid = append(valid, ch)
		default:
			dropped++
		}
	}
	if len(valid) == 0 {
		return lines, dropped
	}
	// 按路径索引从大到小应用,避免索引偏移
	byIdx := map[int]resumeChange{}
	for _, ch := range valid {
		idx, _ := parsePath(ch.Path)
		if _, exists := byIdx[idx]; exists {
			// 同一行多条修改,取最后一条
			byIdx[idx] = ch
		} else {
			byIdx[idx] = ch
		}
	}
	// 收集待应用索引并排序
	indexes := make([]int, 0, len(byIdx))
	for i := range byIdx {
		indexes = append(indexes, i)
	}
	for i := 0; i < len(indexes); i++ {
		for j := i + 1; j < len(indexes); j++ {
			if indexes[i] < indexes[j] {
				indexes[i], indexes[j] = indexes[j], indexes[i]
			}
		}
	}
	for _, idx := range indexes {
		ch := byIdx[idx]
		switch ch.Action {
		case "rewrite":
			lines[idx] = ch.NewValue
		case "delete":
			lines = append(lines[:idx], lines[idx+1:]...)
		case "insert":
			lines = append(lines[:idx+1], append([]string{ch.NewValue}, lines[idx+1:]...)...)
		}
	}
	return lines, dropped
}

// parsePath 解析 "sections.N.items.M" 形式的 path,返回 items 行索引。
func parsePath(path string) (int, bool) {
	re := regexp.MustCompile(`^sections\.(\d+)\.items\.(\d+)$`)
	m := re.FindStringSubmatch(strings.TrimSpace(path))
	if len(m) != 3 {
		return 0, false
	}
	var item int
	if _, err := fmt.Sscanf(m[2], "%d", &item); err != nil {
		return 0, false
	}
	return item, true
}

// containsLocked 判断行文本是否包含锁定字段(个人信息/日期)。
func containsLocked(line string) bool {
	for _, re := range lockedFieldPatterns {
		if re.MatchString(line) {
			return true
		}
	}
	return false
}

// llmOptimizeResume 用 LLM 对简历文本按模板章节做 diff 式优化,返回优化后的行与丢弃数。
func (s *Server) llmOptimizeResume(text string, templateSections []string, directions []string) (string, int, error) {
	if !s.llm.Enabled() {
		return text, 0, nil
	}
	lines := splitLines(text)
	if len(lines) == 0 {
		return text, 0, nil
	}
	systemPrompt := `你是资深简历优化顾问。用户给出简历全文与模板章节结构,请输出结构化的修改建议,让简历更匹配目标岗位。必须返回 JSON:{"changes":[{"path":"sections.0.items.0","action":"rewrite|insert|delete","new_value":"...","reason":"..."}]}。要求:path 严格形如 sections.N.items.M,N 是模板章节序号(0 起),M 是简历行序号(0 起,按行顺序);rewrite 表示改写该行,insert 表示在该行后插入新行,delete 表示删除该行;严禁修改姓名/邮箱/电话/网址/日期等个人信息;量化数字(如性能提升百分比)只能基于原文,禁止虚构;每条修改给出简短 reason。只返回 JSON。`
	userPrompt := fmt.Sprintf("模板章节:\n%s\n\n优化方向:\n%s\n\n简历全文(每行一个条目):\n%s", strings.Join(templateSections, "\n"), strings.Join(directions, "\n"), strings.Join(lines, "\n"))
	answer, err := s.llm.Chat(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return text, 0, err
	}
	answer = cleanJSON(answer)
	var out struct {
		Changes []resumeChange `json:"changes"`
	}
	if json.Unmarshal([]byte(answer), &out) != nil || len(out.Changes) == 0 {
		return text, 0, fmt.Errorf("LLM 优化输出解析失败")
	}
	optimized, dropped := applyChanges(lines, out.Changes)
	return strings.Join(optimized, "\n"), dropped, nil
}

// splitLines 把文本按行拆分,忽略空行。
func splitLines(text string) []string {
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, l := range raw {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, strings.TrimRight(l, "\r"))
		}
	}
	return lines
}
