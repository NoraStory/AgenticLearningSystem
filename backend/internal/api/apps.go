package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) resumeTemplates(c *gin.Context) {
	var rows []model.ResumeTemplate
	q := s.services.DB
	if cat := c.Query("category"); cat != "" && cat != "all" {
		q = q.Where("category = ?", cat)
	}
	q.Order("created_at asc").Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, v := range rows {
		items = append(items, gin.H{"id": v.ID, "name": v.Name, "category": v.Category, "description": v.Description, "preview": v.Preview, "preview_url": "", "sections": v.Sections, "style": v.Style, "structure": gin.H{"sections": v.Sections, "style": v.Style}})
	}
	success(c, gin.H{"templates": items})
}
func (s *Server) resumeUpload(c *gin.Context) {
	h, err := c.FormFile("file")
	if err != nil {
		fail(c, 400, 400, "请选择 PDF、DOC 或 DOCX 简历")
		return
	}
	if h.Size > 15<<20 {
		fail(c, 400, 400, "文件不能超过 15MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(h.Filename))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" && ext != ".txt" {
		fail(c, 400, 400, "仅支持 PDF、DOC、DOCX、TXT")
		return
	}
	key, err := s.store.Save(c, "resumes/"+userID(c), h)
	if err != nil {
		fail(c, 500, 1007, "上传失败")
		return
	}
	r := model.Resume{ID: uuid.NewString(), UserID: userID(c), Filename: h.Filename, ObjectKey: key, FileSize: h.Size, MimeType: h.Header.Get("Content-Type"), AnalysisJSON: "{}"}
	if s.services.DB.Create(&r).Error != nil {
		fail(c, 500, 500, "保存文件记录失败")
		return
	}
	success(c, gin.H{"success": true, "file_id": r.ID, "filename": r.Filename, "file_size": r.FileSize})
}
func (s *Server) resumeAnalyze(c *gin.Context) {
	var in struct {
		FileID string `json:"file_id"`
	}
	if c.ShouldBindJSON(&in) != nil || in.FileID == "" {
		fail(c, 400, 400, "缺少 file_id")
		return
	}
	var r model.Resume
	if s.services.DB.Where("id = ? AND user_id = ?", in.FileID, userID(c)).First(&r).Error != nil {
		fail(c, 404, 404, "简历文件不存在")
		return
	}
	// 已分析过(且是有效结果)直接复用,避免重复调用慢速 LLM
	if r.AnalysisJSON != "" && r.AnalysisJSON != "{}" && r.AnalysisJSON != "null" {
		var cached gin.H
		if json.Unmarshal([]byte(r.AnalysisJSON), &cached) == nil {
			if fb, ok := cached["fallback"].(bool); !ok || !fb {
				success(c, gin.H{"success": true, "analysis": cached, "cached": true})
				return
			}
		}
	}
	text, err := s.extractResumeText(c, r.ObjectKey, r.Filename)
	if err != nil {
		fail(c, 422, 422, err.Error())
		return
	}
	analysis := s.llmAnalyzeResume(text)
	raw, _ := json.Marshal(analysis)
	s.services.DB.Model(&r).Update("analysis_json", string(raw))
	success(c, gin.H{"success": true, "analysis": analysis})
}
func (s *Server) resumeOptimize(c *gin.Context) {
	var in struct {
		FileID                 string   `json:"file_id"`
		TemplateID             string   `json:"template_id"`
		OptimizationDirections []string `json:"optimization_directions"`
	}
	if c.ShouldBindJSON(&in) != nil || in.FileID == "" || in.TemplateID == "" {
		fail(c, 400, 400, "优化参数不完整")
		return
	}
	var r model.Resume
	if s.services.DB.Where("id = ? AND user_id = ?", in.FileID, userID(c)).First(&r).Error != nil {
		fail(c, 404, 404, "简历文件不存在")
		return
	}
	var t model.ResumeTemplate
	if s.services.DB.First(&t, "id = ?", in.TemplateID).Error != nil {
		fail(c, 404, 404, "模板不存在")
		return
	}
	text, err := s.extractResumeText(c, r.ObjectKey, r.Filename)
	if err != nil {
		fail(c, 422, 422, err.Error())
		return
	}
	// 模板章节结构(内置模板 sections 存的是章节名数组,自定义模板同样)
	tplSections := t.Sections
	if len(tplSections) == 0 {
		tplSections = []string{"基本信息", "技能清单", "工作经历", "项目经验", "教育背景"}
	}
	optimized, dropped, err := s.llmOptimizeResume(text, tplSections, in.OptimizationDirections)
	if err != nil {
		optimized = text
	}
	// 按模板章节组织输出(展示用;前端实际取 text 全文)
	sections := make([]gin.H, 0, len(tplSections))
	var b strings.Builder
	b.WriteString(optimized + "\n")
	for _, section := range tplSections {
		sections = append(sections, gin.H{"title": section, "content": ""})
		b.WriteString("## " + section + "\n\n")
	}
	_ = dropped
	s.services.DB.Model(&r).Update("optimized_content", b.String())
	success(c, gin.H{"success": true, "fallback": err != nil, "dropped_count": dropped, "optimized_content": gin.H{"sections": sections, "template_used": t.ID, "text": b.String()}})
}
func (s *Server) resumeExport(c *gin.Context) {
	var in struct {
		Format     string `json:"format"`
		TemplateID string `json:"template_id"`
		Content    any    `json:"content"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "导出参数无效")
		return
	}
	format := strings.ToLower(in.Format)
	if format == "" {
		format = "pdf"
	}
	raw := contentString(in.Content)

	// 指定了已注册模板:docx 走 docxtpl 渲染,PDF 再经 LibreOffice 转换,保留模板样式
	if in.TemplateID != "" {
		var t model.ResumeTemplate
		if s.services.DB.First(&t, "id = ?", in.TemplateID).Error == nil && t.Status == "ready" && t.RegisteredPath != "" {
			sections, err := s.exportSectionsFromContent(in.Content)
			if err != nil {
				fail(c, 400, 400, "导出内容格式无效")
				return
			}
			sections = alignTemplateSections(sections, t.Sections)
			data, err := s.renderDocxTemplate(c.Request.Context(), t.RegisteredPath, sections)
			if err != nil {
				fail(c, 500, 1010, "DOCX 渲染失败: "+err.Error())
				return
			}
			switch format {
			case "docx":
				c.Header("Content-Disposition", "attachment; filename=resume.docx")
				c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
				return
			case "pdf":
				pdfData, err := s.docxConverter().convertDocxToPDF(c.Request.Context(), data)
				if err != nil {
					// 转换服务不可用时回退文本版 PDF
					c.Header("Content-Disposition", "attachment; filename=resume.pdf")
					c.Data(200, "application/pdf", makePDF(raw))
					return
				}
				c.Header("Content-Disposition", "attachment; filename=resume.pdf")
				c.Data(200, "application/pdf", pdfData)
				return
			}
			// 其他格式(html)走下面通用分支
		}
	}

	switch format {
	case "html":
		c.Header("Content-Disposition", "attachment; filename=resume.html")
		c.Data(200, "text/html; charset=utf-8", []byte("<!doctype html><html><meta charset=utf-8><body><pre>"+html.EscapeString(raw)+"</pre></body></html>"))
	case "docx":
		data, err := makeDocx(raw)
		if err != nil {
			fail(c, 500, 1010, "DOCX 导出失败")
			return
		}
		c.Header("Content-Disposition", "attachment; filename=resume.docx")
		c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
	default:
		data := makePDF(raw)
		c.Header("Content-Disposition", "attachment; filename=resume.pdf")
		c.Data(200, "application/pdf", data)
	}
}

// exportSectionsFromContent 从导出请求的 content 中提取 sections 数据(docxtpl 渲染用)。
func (s *Server) exportSectionsFromContent(content any) ([]map[string]any, error) {
	// 前端可能传 {sections: [...]} 或 {optimized_content: {sections: [...]}} 或 {text: "..."}
	raw, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	var probe map[string]any
	if json.Unmarshal(raw, &probe) != nil {
		return []map[string]any{{"title": "内容", "items": []string{contentString(content)}}}, nil
	}
	if secs, ok := probe["sections"].([]any); ok {
		return toSectionMaps(secs)
	}
	if oc, ok := probe["optimized_content"].(map[string]any); ok {
		if secs, ok := oc["sections"].([]any); ok {
			return toSectionMaps(secs)
		}
	}
	if text, ok := probe["text"].(string); ok {
		return sectionsFromMarkdown(text), nil
	}
	return []map[string]any{{"title": "内容", "items": []string{contentString(content)}}}, nil
}

// alignTemplateSections 按模板章节名对齐 sections:模板有而 content 没有的章节补空 items;
// content 有而模板没有的章节忽略(避免 docxtpl 对占位符越界抛错)。
func alignTemplateSections(content []map[string]any, templateSectionNames []string) []map[string]any {
	byTitle := map[string][]string{}
	for _, sec := range content {
		title, _ := sec["title"].(string)
		items := []string{}
		if raw, ok := sec["items"].([]any); ok {
			for _, it := range raw {
				if s, ok := it.(string); ok {
					items = append(items, s)
				}
			}
		} else if raw, ok := sec["items"].([]string); ok {
			items = raw
		}
		if title != "" {
			byTitle[title] = items
		}
	}
	out := make([]map[string]any, 0, len(templateSectionNames))
	for _, title := range templateSectionNames {
		items, ok := byTitle[title]
		if !ok {
			items = []string{}
		}
		out = append(out, map[string]any{"title": title, "items": items})
	}
	return out
}

// sectionsFromMarkdown 把 "## 章节名\n条目" 形式的 markdown 拆成 sections 结构。
func sectionsFromMarkdown(text string) []map[string]any {
	sections := []map[string]any{}
	var current string
	var items []string
	flush := func() {
		if current != "" {
			sections = append(sections, map[string]any{"title": current, "items": items})
		}
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "##") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "##"))
			items = []string{}
		} else if strings.HasPrefix(line, "#") {
			flush()
			current = strings.TrimSpace(strings.TrimLeft(line, "#"))
			items = []string{}
		} else {
			items = append(items, line)
		}
	}
	flush()
	if len(sections) == 0 {
		sections = append(sections, map[string]any{"title": "内容", "items": []string{text}})
	}
	return sections
}

func toSectionMaps(secs []any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(secs))
	for _, sec := range secs {
		m, ok := sec.(map[string]any)
		if !ok {
			continue
		}
		title, _ := m["title"].(string)
		items := []string{}
		if rawItems, ok := m["items"].([]any); ok {
			for _, it := range rawItems {
				if s, ok := it.(string); ok {
					items = append(items, s)
				}
			}
		}
		if title != "" || len(items) > 0 {
			out = append(out, map[string]any{"title": title, "items": items})
		}
	}
	return out, nil
}
func contentString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		b, _ := json.MarshalIndent(x, "", "  ")
		return string(b)
	}
}
func makeDocx(text string) ([]byte, error) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	files := map[string]string{"[Content_Types].xml": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><Types xmlns=\"http://schemas.openxmlformats.org/package/2006/content-types\"><Default Extension=\"rels\" ContentType=\"application/vnd.openxmlformats-package.relationships+xml\"/><Default Extension=\"xml\" ContentType=\"application/xml\"/><Override PartName=\"/word/document.xml\" ContentType=\"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml\"/></Types>", "_rels/.rels": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><Relationships xmlns=\"http://schemas.openxmlformats.org/package/2006/relationships\"><Relationship Id=\"rId1\" Type=\"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument\" Target=\"word/document.xml\"/></Relationships>", "word/document.xml": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><w:document xmlns:w=\"http://schemas.openxmlformats.org/wordprocessingml/2006/main\"><w:body><w:p><w:r><w:t xml:space=\"preserve\">" + html.EscapeString(text) + "</w:t></w:r></w:p></w:body></w:document>"}
	for name, body := range files {
		w, e := z.Create(name)
		if e != nil {
			return nil, e
		}
		w.Write([]byte(body))
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func makePDF(text string) []byte {
	safe := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)", "\r", "", "\n", " ").Replace(text)
	if len(safe) > 1500 {
		safe = safe[:1500]
	}
	stream := "BT /F1 11 Tf 50 780 Td (" + safe + ") Tj ET"
	parts := []string{"%PDF-1.4\n", "1 0 obj<< /Type /Catalog /Pages 2 0 R >>endobj\n", "2 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n", "3 0 obj<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources<< /Font<< /F1 4 0 R >> >> /Contents 5 0 R >>endobj\n", "4 0 obj<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>endobj\n", fmt.Sprintf("5 0 obj<< /Length %d >>stream\n%s\nendstream endobj\n", len(stream), stream)}
	var b strings.Builder
	offsets := []int{0}
	for _, p := range parts {
		offsets = append(offsets, b.Len())
		b.WriteString(p)
	}
	xref := b.Len()
	b.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		b.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	b.WriteString(fmt.Sprintf("trailer<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", xref))
	return []byte(b.String())
}

func (s *Server) listProjects(c *gin.Context) {
	var rows []model.Project
	s.services.DB.Preload("Tasks").Where("user_id = ?", userID(c)).Order("created_at desc").Find(&rows)
	success(c, gin.H{"projects": rows})
}
func (s *Server) createProject(c *gin.Context) {
	var in struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		TechStack      []string `json:"tech_stack"`
		TechStackCamel []string `json:"techStack"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" {
		fail(c, 400, 400, "项目名称不能为空")
		return
	}
	if len(in.TechStack) == 0 {
		in.TechStack = in.TechStackCamel
	}
	p := model.Project{ID: uuid.NewString(), UserID: userID(c), Name: in.Name, Description: in.Description, TechStack: in.TechStack}
	tasks := []model.ProjectTask{{ID: uuid.NewString(), ProjectID: p.ID, Title: "需求与架构设计", Description: "梳理用户故事、数据模型和核心接口", Status: "pending"}, {ID: uuid.NewString(), ProjectID: p.ID, Title: "核心功能开发", Description: "完成主要业务流程与异常处理", Status: "pending"}, {ID: uuid.NewString(), ProjectID: p.ID, Title: "测试与部署", Description: "补充测试、性能检查和部署文档", Status: "pending"}}
	if err := s.services.DB.Create(&p).Error; err != nil {
		fail(c, 500, 500, "创建项目失败")
		return
	}
	s.services.DB.Create(&tasks)
	p.Tasks = tasks
	success(c, gin.H{"success": true, "project": p})
}
func (s *Server) projectUpload(c *gin.Context) {
	projectID := c.PostForm("projectId")
	if projectID == "" {
		projectID = c.PostForm("project_id")
	}
	var project model.Project
	if s.services.DB.Where("id = ? AND user_id = ?", projectID, userID(c)).First(&project).Error != nil {
		fail(c, 404, 404, "项目不存在")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		fail(c, 400, 400, "上传格式错误")
		return
	}
	headers := form.File["files"]
	if len(headers) == 0 {
		fail(c, 400, 400, "请选择源码文件")
		return
	}
	saved := []gin.H{}
	keys := []string{}
	for _, h := range headers {
		if h.Size > 10<<20 {
			continue
		}
		key, err := s.store.Save(c, "projects/"+projectID, h)
		if err == nil {
			keys = append(keys, key)
			saved = append(saved, gin.H{"name": h.Filename, "size": h.Size, "key": key})
		}
	}
	var task model.ProjectTask
	s.services.DB.Where("project_id = ? AND status != ?", projectID, "completed").Order("created_at").First(&task)
	if task.ID == "" {
		task = model.ProjectTask{ID: uuid.NewString(), ProjectID: projectID, Title: "源码分析", Description: "分析用户上传的项目源码", Status: "in_progress", Files: keys}
		s.services.DB.Create(&task)
	} else {
		task.Files = keys
		task.Status = "in_progress"
		s.services.DB.Save(&task)
	}
	success(c, gin.H{"success": true, "task_id": task.ID, "files": saved})
}
func (s *Server) projectAnalyze(c *gin.Context) {
	var in struct {
		ProjectID string `json:"project_id"`
		TaskID    string `json:"task_id"`
	}
	if c.ShouldBindJSON(&in) != nil || in.ProjectID == "" {
		fail(c, 400, 400, "分析参数无效")
		return
	}
	var p model.Project
	if s.services.DB.Preload("Tasks").Where("id = ? AND user_id = ?", in.ProjectID, userID(c)).First(&p).Error != nil {
		fail(c, 404, 404, "项目不存在")
		return
	}
	files := []string{}
	completed := []string{}
	pending := []string{}
	for _, t := range p.Tasks {
		files = append(files, t.Files...)
		if t.Status == "completed" {
			completed = append(completed, t.Title)
		} else {
			pending = append(pending, t.Title)
		}
	}
	suggestions := []string{"为核心业务补充单元测试与接口测试", "统一错误响应和日志 trace_id", "补充环境变量示例与一键启动脚本", "为上传文件增加类型和大小校验"}
	analysis := gin.H{"files": files, "completed_tasks": completed, "pending_tasks": pending, "suggestions": suggestions}
	if in.TaskID != "" {
		raw, _ := json.Marshal(analysis)
		s.services.DB.Model(&model.ProjectTask{}).Where("id = ? AND project_id = ?", in.TaskID, in.ProjectID).Updates(map[string]any{"analysis": string(raw), "status": "completed"})
	}
	success(c, gin.H{"success": true, "analysis": analysis})
}

func (s *Server) generateExam(c *gin.Context) {
	s.generateExamV2(c)
}
func questionPayload(rows []model.InterviewQuestion) []gin.H {
	out := make([]gin.H, 0, len(rows))
	for _, q := range rows {
		out = append(out, gin.H{"id": q.ID, "type": q.Type, "category": q.Category, "difficulty": q.Difficulty, "title": q.Title, "description": q.Description, "constraints": q.Constraints, "time_limit": q.TimeLimit, "timeLimit": q.TimeLimit, "score": q.Score, "example": q.Example})
	}
	return out
}
func (s *Server) listExams(c *gin.Context) {
	var rows []model.InterviewExam
	s.services.DB.Where("user_id = ?", userID(c)).Order("created_at desc").Limit(30).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, e := range rows {
		items = append(items, gin.H{"id": e.ID, "exam_id": e.ID, "date": e.CreatedAt.Format("2006-01-02"), "score": e.Score, "direction": e.Direction, "difficulty": e.Difficulty})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) examDetail(c *gin.Context) {
	var e model.InterviewExam
	if s.services.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID(c)).First(&e).Error != nil {
		fail(c, 404, 404, "笔试不存在")
		return
	}
	success(c, gin.H{"exam_id": e.ID, "direction": e.Direction, "difficulty": e.Difficulty, "questions": questionPayload(e.Questions), "score": e.Score})
}
func (s *Server) runExamQuestion(c *gin.Context) {
	s.runExamQuestionV2(c)
}
func (s *Server) submitExam(c *gin.Context) {
	s.submitExamV2(c)
}
func (s *Server) search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		success(c, gin.H{"courses": []any{}, "problems": []any{}, "notes": []any{}})
		return
	}
	like := "%" + q + "%"
	var courses []model.Course
	var problems []model.Problem
	var notes []model.Note
	s.services.DB.Where("title ILIKE ? OR summary ILIKE ?", like, like).Limit(10).Find(&courses)
	s.services.DB.Where("title ILIKE ? OR description ILIKE ?", like, like).Limit(10).Find(&problems)
	s.services.DB.Where("user_id = ? AND (title ILIKE ? OR content ILIKE ?)", userID(c), like, like).Limit(10).Find(&notes)
	success(c, gin.H{"courses": courses, "problems": problems, "notes": notes})
}

var _ = strconv.Itoa
var _ = time.Now
