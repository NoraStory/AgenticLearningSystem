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
	analysis := gin.H{"score": 78, "atsScore": 74, "keywordMatch": []gin.H{{"keyword": "React", "found": true}, {"keyword": "TypeScript", "found": true}, {"keyword": "Node.js", "found": true}, {"keyword": "Python", "found": true}, {"keyword": "Docker", "found": false}, {"keyword": "云服务", "found": false}, {"keyword": "Git", "found": true}, {"keyword": "SQL", "found": true}}, "strengths": []string{"经历结构清晰，关键职责可快速定位", "技术栈与软件开发岗位匹配", "项目描述包含一定的结果信息"}, "weaknesses": []string{"部分成果缺少量化指标", "项目技术难点与个人贡献区分不够", "缺少针对目标岗位的关键词"}, "suggestions": []string{"使用 STAR 法则重写最近两段项目经历", "补充性能、用户量、成本或效率等量化结果", "按编程语言、框架、数据库、工程工具重组技能", "根据目标职位描述补充 Docker 与云服务关键词"}, "overall_score": 78, "dimensions": gin.H{"content_completeness": 82, "format_layout": 80, "keyword_match": 74, "professionalism": 79, "expression_clarity": 76}, "highlights": []string{"技术经历完整", "项目方向明确"}}
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
	sections := make([]gin.H, 0, len(t.Sections))
	var text strings.Builder
	for _, section := range t.Sections {
		content := optimizedSection(section)
		sections = append(sections, gin.H{"title": section, "content": content})
		text.WriteString("## " + section + "\n" + content + "\n\n")
	}
	s.services.DB.Model(&r).Update("optimized_content", text.String())
	success(c, gin.H{"success": true, "optimized_content": gin.H{"sections": sections, "template_used": t.ID, "text": text.String()}})
}
func optimizedSection(section string) string {
	m := map[string]string{"基本信息": "张三｜软件工程师｜上海｜zhangsan@example.com｜github.com/example", "个人简介": "5 年软件开发经验，擅长 TypeScript、Python 与云原生工程。能够从业务目标出发完成架构设计、性能优化和团队协作。", "技能清单": "编程语言：TypeScript、Python、Go、SQL\n框架：React、Next.js、Gin\n工程工具：Git、Docker、CI/CD、PostgreSQL、Redis", "技术栈": "前端：React、Next.js、TypeScript、Tailwind CSS\n后端：Go、Gin、PostgreSQL、Redis\n工程：Docker、GitHub Actions、监控与日志", "工作经历": "高级软件工程师｜示例科技｜2022.03 至今\n- 主导核心平台重构，接口平均延迟降低 42%\n- 建立组件库与代码审查规范，交付效率提升 30%\n- 指导 3 名工程师完成复杂模块交付", "项目经验": "CodeForge 学习平台｜技术负责人\n- 设计前后端分离架构与 Agent 工作流\n- 实现课程、题库、代码执行、简历与项目分析闭环\n- 通过缓存和索引优化将主要查询稳定在 100ms 内", "教育背景": "示例大学｜计算机科学与技术｜本科｜2016-2020"}
	if v := m[section]; v != "" {
		return v
	}
	return "请根据目标岗位补充可验证、可量化的 " + section + " 内容。"
}
func (s *Server) resumeExport(c *gin.Context) {
	var in struct {
		Format  string `json:"format"`
		Content any    `json:"content"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "导出参数无效")
		return
	}
	raw := contentString(in.Content)
	switch strings.ToLower(in.Format) {
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
