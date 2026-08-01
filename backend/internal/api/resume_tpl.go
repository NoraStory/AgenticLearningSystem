package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// resumeTemplateUpload 上传用户自己的 docx 模板,解析结构并让 LLM 识别章节,返回待确认的章节列表(Status=draft)。
func (s *Server) resumeTemplateUpload(c *gin.Context) {
	h, err := c.FormFile("file")
	if err != nil {
		fail(c, 400, 400, "请选择 DOCX 简历模板")
		return
	}
	if h.Size > 15<<20 {
		fail(c, 400, 400, "文件不能超过 15MB")
		return
	}
	if !strings.EqualFold(filepath.Ext(h.Filename), ".docx") {
		fail(c, 400, 400, "仅支持 DOCX 模板")
		return
	}

	// 存 MinIO 原件
	key, err := s.store.Save(c, "templates/"+userID(c), h)
	if err != nil {
		fail(c, 500, 1007, "模板上传失败")
		return
	}
	// 落 draft 记录
	tpl := model.ResumeTemplate{
		ID:          uuid.NewString(),
		Name:        strings.TrimSuffix(h.Filename, filepath.Ext(h.Filename)),
		Category:    "custom",
		Description: "用户上传的自定义模板",
		Preview:     "📄",
		ObjectKey:   key,
		Status:      "draft",
		OwnerID:     userID(c),
	}
	if s.services.DB.Create(&tpl).Error != nil {
		fail(c, 500, 500, "保存模板记录失败")
		return
	}

	// 下载到本地临时文件供 Python 解析
	tmpDir, err := os.MkdirTemp("", "codeforge-tpl-*")
	if err != nil {
		fail(c, 500, 500, "模板处理失败")
		return
	}
	defer os.RemoveAll(tmpDir)
	srcPath := filepath.Join(tmpDir, h.Filename)
	rc, err := s.store.Open(c, key)
	if err != nil {
		fail(c, 500, 500, "模板读取失败")
		return
	}
	defer rc.Close()
	out, err := os.Create(srcPath)
	if err != nil {
		fail(c, 500, 500, "模板处理失败")
		return
	}
	if _, err := out.ReadFrom(rc); err != nil {
		out.Close()
		fail(c, 500, 500, "模板处理失败")
		return
	}
	out.Close()

	// Python 解析结构
	parsed, err := s.parseDocxTemplate(context.Background(), srcPath)
	if err != nil {
		fail(c, 500, 500, "模板解析失败: "+err.Error())
		return
	}

	// 组装解析结果给 LLM 章节识别
	paras := s.parseParagraphs(parsed)

	sections := s.llmIdentifySections(paras)

	// 模型 Sections 是 []string(章节名数组),只存章节标题;完整结构(标题+条目)由前端持有
	secNames := make([]string, 0, len(sections))
	for _, sec := range sections {
		if title, ok := sec["title"].(string); ok && title != "" {
			secNames = append(secNames, title)
		}
	}
	if raw, err := json.Marshal(secNames); err == nil {
		s.services.DB.Model(&tpl).Update("sections", string(raw))
	}
	success(c, gin.H{
		"template_id": tpl.ID,
		"name":        tpl.Name,
		"sections":    sections,
		"parsed":      gin.H{"paragraph_count": len(paras), "table_count": len(parsed["tables"].([]any))},
	})
}

// paraItem 表示解析出的一个模板段落。
type paraItem struct {
	Idx   int    `json:"idx"`
	Text  string `json:"text"`
	Style string `json:"style"`
}

// parseParagraphs 从 tpl_parse 的输出中提取非空段落列表。
func (s *Server) parseParagraphs(parsed map[string]any) []paraItem {
	var paras []paraItem
	if raw, ok := parsed["paragraphs"].([]any); ok {
		for _, p := range raw {
			m, _ := p.(map[string]any)
			text, _ := m["text"].(string)
			style, _ := m["style"].(string)
			idx, _ := m["idx"].(float64)
			if strings.TrimSpace(text) != "" {
				paras = append(paras, paraItem{Idx: int(idx), Text: text, Style: style})
			}
		}
	}
	return paras
}

// llmIdentifySections 用 LLM 把模板段落归类为章节(title + items)。LLM 不可用/解析失败时按标题段落粗分。
func (s *Server) llmIdentifySections(paras []paraItem) []gin.H {
	fallback := s.heuristicSections(paras)
	if !s.llm.Enabled() || len(paras) == 0 {
		return fallback
	}
	// 段落文本序列化给 LLM
	var b strings.Builder
	for _, p := range paras {
		fmt.Fprintf(&b, "[%d] %s\n", p.Idx, p.Text)
	}
	systemPrompt := "你是简历模板结构分析师。用户给出一份简历模板的段落列表(每行 [序号] 文本)。请识别出各章节(如基本信息/技能清单/工作经历/项目经验/教育背景),每个章节下包含哪些条目(段落)。必须返回 JSON 对象 {\"sections\":[{\"title\":\"章节名\",\"items\":[\"条目原文\"]}]}。要求:items 必须是模板中存在的原文段落,不要改写;不认识的杂项段落归入其最近的章节;只返回 JSON,不要其他内容。"
	answer, err := s.llm.Chat(context.Background(), systemPrompt, b.String())
	if err != nil {
		return fallback
	}
	answer = strings.TrimSpace(answer)
	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimSuffix(answer, "```")
	answer = strings.TrimSpace(answer)
	var out struct {
		Sections []struct {
			Title string   `json:"title"`
			Items []string `json:"items"`
		} `json:"sections"`
	}
	if json.Unmarshal([]byte(answer), &out) != nil || len(out.Sections) == 0 {
		return fallback
	}
	sections := make([]gin.H, 0, len(out.Sections))
	for _, sec := range out.Sections {
		title := strings.TrimSpace(sec.Title)
		if title == "" {
			continue
		}
		items := make([]string, 0, len(sec.Items))
		for _, it := range sec.Items {
			if t := strings.TrimSpace(it); t != "" {
				items = append(items, t)
			}
		}
		sections = append(sections, gin.H{"title": title, "items": items})
	}
	if len(sections) == 0 {
		return fallback
	}
	return sections
}

// heuristicSections 无 LLM 时的粗略章节识别:按标题样式(Heading)分段。
func (s *Server) heuristicSections(paras []paraItem) []gin.H {
	sections := []gin.H{}
	var current string
	var items []string
	flush := func() {
		if current != "" {
			sections = append(sections, gin.H{"title": current, "items": items})
		}
	}
	for _, p := range paras {
		if strings.Contains(p.Style, "Heading") || strings.Contains(p.Style, "标题") {
			flush()
			current = p.Text
			items = []string{}
		} else if current != "" {
			items = append(items, p.Text)
		}
	}
	flush()
	if len(sections) == 0 && len(paras) > 0 {
		items := make([]string, 0, len(paras))
		for _, p := range paras {
			items = append(items, p.Text)
		}
		sections = append(sections, gin.H{"title": "内容", "items": items})
	}
	return sections
}

// resumeTemplateConfirm 用户确认/修正章节结构后,注入占位符生成注册模板(Status=ready)。
func (s *Server) resumeTemplateConfirm(c *gin.Context) {
	id := c.Param("id")
	var tpl model.ResumeTemplate
	if s.services.DB.First(&tpl, "id = ?", id).Error != nil || tpl.OwnerID != userID(c) {
		fail(c, 404, 404, "模板不存在")
		return
	}
	var in struct {
		Name     string `json:"name"`
		Sections []struct {
			Title string   `json:"title"`
			Items []string `json:"items"`
		} `json:"sections"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "参数无效")
		return
	}
	if len(in.Sections) == 0 {
		fail(c, 400, 400, "至少需要一个章节")
		return
	}
	sections := make([]map[string]any, 0, len(in.Sections))
	for _, sec := range in.Sections {
		title := strings.TrimSpace(sec.Title)
		items := []string{}
		for _, it := range sec.Items {
			if t := strings.TrimSpace(it); t != "" {
				items = append(items, t)
			}
		}
		if title == "" {
			fail(c, 400, 400, "章节标题不能为空")
			return
		}
		sections = append(sections, map[string]any{"title": title, "items": items})
	}

	// 下载原件到本地,注入占位符
	tmpDir, err := os.MkdirTemp("", "codeforge-tpl-*")
	if err != nil {
		fail(c, 500, 500, "模板处理失败")
		return
	}
	defer os.RemoveAll(tmpDir)
	srcPath := filepath.Join(tmpDir, "source.docx")
	rc, err := s.store.Open(c, tpl.ObjectKey)
	if err != nil {
		fail(c, 500, 500, "模板读取失败")
		return
	}
	outFile, err := os.Create(srcPath)
	if err != nil {
		rc.Close()
		fail(c, 500, 500, "模板处理失败")
		return
	}
	if _, err := outFile.ReadFrom(rc); err != nil {
		rc.Close()
		outFile.Close()
		fail(c, 500, 500, "模板处理失败")
		return
	}
	rc.Close()
	outFile.Close()

	// 注册后的模板缓存到 backend/data/templates/{id}.docx
	regDir := filepath.Join(s.backendRoot(), "data", "templates")
	if err := os.MkdirAll(regDir, 0o755); err != nil {
		fail(c, 500, 500, "模板处理失败")
		return
	}
	regPath := filepath.Join(regDir, id+".docx")
	if err := s.registerDocxTemplate(context.Background(), srcPath, regPath, sections); err != nil {
		fail(c, 500, 500, "模板注册失败: "+err.Error())
		return
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = tpl.Name
	}
	// 章节列表转 string 存 Sections
	secNames := make([]string, 0, len(in.Sections))
	for _, sec := range in.Sections {
		secNames = append(secNames, strings.TrimSpace(sec.Title))
	}
	secJSON, _ := json.Marshal(secNames)
	s.services.DB.Model(&tpl).Updates(map[string]any{
		"name":            name,
		"sections":        string(secJSON),
		"style":           "用户自定义模板",
		"registered_path": regPath,
		"status":          "ready",
	})
	success(c, gin.H{"success": true, "template_id": tpl.ID, "status": "ready"})
}
