package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"codeforge/backend/internal/model"
	"codeforge/backend/internal/sandbox"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) listCourses(c *gin.Context) {
	q := s.services.DB.Model(&model.Course{})
	if v := c.Query("category"); v != "" && v != "all" {
		q = q.Where("category = ?", v)
	}
	if v := c.Query("difficulty"); v != "" && v != "all" {
		q = q.Where("difficulty = ? OR level = ?", v, v)
	}
	if v := c.Query("level"); v != "" && v != "all" && v != "全部" {
		q = q.Where("level = ?", v)
	}
	if v := c.Query("status"); v != "" && v != "all" && v != "全部" {
		q = q.Where("status = ?", strings.ToLower(v))
	}
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 50)
	if pageSize > 100 {
		pageSize = 100
	}
	var total int64
	q.Count(&total)
	var rows []model.Course
	q.Order("id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, v := range rows {
		items = append(items, courseListItem(v))
	}
	success(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items})
}
func courseListItem(v model.Course) gin.H {
	return gin.H{"id": v.ID, "course_id": v.ID, "title": v.Title, "category": v.Category, "categoryLabel": v.CategoryLabel, "category_label": v.CategoryLabel, "categoryColor": categoryColor(v.Category), "difficulty": v.Difficulty, "level": v.Level, "status": v.Status, "cover_image": v.CoverImage, "image": v.CoverImage, "summary": v.Summary, "description": v.Description, "author": v.Author, "publish_date": v.PublishDate, "date": v.PublishDate, "read_time": v.ReadTime, "readTime": v.ReadTime, "views": v.Views, "tags": v.Tags, "progress": v.Progress, "lessons": v.LessonsCount, "lessons_count": v.LessonsCount, "estimatedHours": v.EstimatedHours, "estimated_hours": v.EstimatedHours}
}
func categoryColor(v string) string {
	m := map[string]string{"python": "bg-blue-500/10 text-blue-600", "cpp": "bg-purple-500/10 text-purple-600", "database": "bg-cyan-500/10 text-cyan-600", "algorithm": "bg-orange-500/10 text-orange-600", "agent": "bg-rose-500/10 text-rose-600"}
	return m[v]
}
func queryInt(c *gin.Context, key string, def int) int {
	v, e := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(def)))
	if e != nil || v < 1 {
		return def
	}
	return v
}
func (s *Server) recommendedCourses(c *gin.Context) {
	var rows []model.Course
	s.services.DB.Order("views desc").Limit(4).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, v := range rows {
		items = append(items, gin.H{"id": v.ID, "title": v.Title, "category": v.CategoryLabel})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) courseResources(c *gin.Context) {
	cat := c.DefaultQuery("category", "python")
	resources := map[string][]gin.H{"python": {{"name": "Python 官方文档", "type": "文档", "url": "https://docs.python.org/zh-cn/3/"}, {"name": "Real Python", "type": "教程", "url": "https://realpython.com/"}}, "cpp": {{"name": "cppreference", "type": "文档", "url": "https://en.cppreference.com/"}, {"name": "C++ Core Guidelines", "type": "规范", "url": "https://isocpp.github.io/CppCoreGuidelines/"}}, "database": {{"name": "PostgreSQL Documentation", "type": "文档", "url": "https://www.postgresql.org/docs/"}, {"name": "Redis Documentation", "type": "文档", "url": "https://redis.io/docs/"}}, "agent": {{"name": "LangChain Docs", "type": "文档", "url": "https://python.langchain.com/"}, {"name": "OpenAI Cookbook", "type": "示例", "url": "https://cookbook.openai.com/"}}}
	success(c, gin.H{"items": resources[cat]})
}
func (s *Server) courseTags(c *gin.Context) {
	cat := c.Query("category")
	var rows []model.Course
	q := s.services.DB
	if cat != "" {
		q = q.Where("category = ?", cat)
	}
	q.Find(&rows)
	counts := map[string]int{}
	for _, v := range rows {
		for _, t := range v.Tags {
			counts[t]++
		}
	}
	items := make([]gin.H, 0, len(counts))
	for n, count := range counts {
		items = append(items, gin.H{"name": n, "count": count})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) hotTags(c *gin.Context) {
	var rows []model.Course
	s.services.DB.Find(&rows)
	counts := map[string]int{}
	for _, v := range rows {
		for _, t := range v.Tags {
			counts[t]++
		}
	}
	tags := make([]string, 0, len(counts))
	for k := range counts {
		tags = append(tags, k)
	}
	if len(tags) > 12 {
		tags = tags[:12]
	}
	success(c, gin.H{"tags": tags})
}
func (s *Server) courseDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fail(c, 400, 400, "课程 ID 无效")
		return
	}
	var v model.Course
	if s.services.DB.First(&v, id).Error != nil {
		fail(c, 404, 404, "课程不存在")
		return
	}
	var prev, next model.Course
	s.services.DB.Where("id < ?", v.ID).Order("id desc").First(&prev)
	s.services.DB.Where("id > ?", v.ID).Order("id asc").First(&next)
	data := courseListItem(v)
	data["sections"] = v.Sections
	data["author"] = v.Author
	data["prev_course"] = func() any {
		if prev.ID == 0 {
			return nil
		}
		return gin.H{"id": prev.ID, "title": prev.Title}
	}()
	data["next_course"] = func() any {
		if next.ID == 0 {
			return nil
		}
		return gin.H{"id": next.ID, "title": next.Title}
	}()
	success(c, data)
}
func (s *Server) comments(c *gin.Context) {
	courseID, _ := strconv.Atoi(c.Param("id"))
	var rows []model.Comment
	s.services.DB.Where("course_id = ?", courseID).Order("created_at desc").Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, v := range rows {
		var u model.User
		s.services.DB.First(&u, "id = ?", v.UserID)
		items = append(items, gin.H{"comment_id": v.ID, "user": gin.H{"user_id": u.ID, "username": u.Username, "avatar": u.Avatar}, "content": v.Content, "created_at": v.CreatedAt, "likes": v.Likes})
	}
	success(c, gin.H{"total": len(items), "items": items})
}
func (s *Server) createComment(c *gin.Context) {
	courseID, e := strconv.Atoi(c.Param("id"))
	var in struct {
		Content string `json:"content"`
	}
	if e != nil || c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Content) == "" {
		fail(c, 400, 400, "评论内容不能为空")
		return
	}
	v := model.Comment{ID: uuid.NewString(), CourseID: uint(courseID), UserID: userID(c), Content: strings.TrimSpace(in.Content)}
	if s.services.DB.Create(&v).Error != nil {
		fail(c, 500, 500, "评论失败")
		return
	}
	s.addActivity(userID(c), "comment", "发表了课程评论")
	success(c, gin.H{"comment_id": v.ID, "content": v.Content, "created_at": v.CreatedAt})
}
func (s *Server) toggleLike(c *gin.Context) {
	courseID, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		fail(c, 400, 400, "课程 ID 无效")
		return
	}
	var row model.CourseLike
	err := s.services.DB.Where("course_id = ? AND user_id = ?", courseID, userID(c)).First(&row).Error
	liked := false
	if err == nil {
		s.services.DB.Delete(&row)
	} else {
		row = model.CourseLike{ID: uuid.NewString(), CourseID: uint(courseID), UserID: userID(c)}
		s.services.DB.Create(&row)
		liked = true
	}
	var count int64
	s.services.DB.Model(&model.CourseLike{}).Where("course_id = ?", courseID).Count(&count)
	if liked {
		s.trackEvent(userID(c), "like", "点赞了一门课程", gin.H{"course_id": courseID})
	}
	success(c, gin.H{"liked": liked, "like_count": count})
}
func (s *Server) toggleBookmark(c *gin.Context) {
	courseID, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		fail(c, 400, 400, "课程 ID 无效")
		return
	}
	bookmarked := toggleFavorite(s, userID(c), uint(courseID))
	s.trackEvent(userID(c), "bookmark", "收藏了一门课程", gin.H{"course_id": courseID, "bookmarked": bookmarked})
	success(c, gin.H{"bookmarked": bookmarked})
}
func toggleFavorite(s *Server, uid string, courseID uint) bool {
	var row model.Favorite
	err := s.services.DB.Where("course_id = ? AND user_id = ?", courseID, uid).First(&row).Error
	if err == nil {
		s.services.DB.Delete(&row)
		return false
	}
	s.services.DB.Create(&model.Favorite{ID: uuid.NewString(), CourseID: courseID, UserID: uid})
	return true
}
func (s *Server) addFavorite(c *gin.Context) {
	var in struct {
		CourseID uint `json:"course_id"`
	}
	if c.ShouldBindJSON(&in) != nil || in.CourseID == 0 {
		fail(c, 400, 400, "课程 ID 无效")
		return
	}
	var row model.Favorite
	if s.services.DB.Where("course_id = ? AND user_id = ?", in.CourseID, userID(c)).First(&row).Error != nil {
		s.services.DB.Create(&model.Favorite{ID: uuid.NewString(), CourseID: in.CourseID, UserID: userID(c)})
	}
	success(c, gin.H{"favorited": true})
}
func (s *Server) deleteFavorite(c *gin.Context) {
	courseID, e := strconv.Atoi(c.Param("course_id"))
	if e != nil {
		fail(c, 400, 400, "课程 ID 无效")
		return
	}
	s.services.DB.Where("course_id = ? AND user_id = ?", courseID, userID(c)).Delete(&model.Favorite{})
	success(c, gin.H{})
}
func (s *Server) createNote(c *gin.Context) {
	var in struct {
		CourseID uint   `json:"course_id"`
		Title    string `json:"title"`
		Content  string `json:"content"`
	}
	if c.ShouldBindJSON(&in) != nil || in.CourseID == 0 || strings.TrimSpace(in.Title) == "" {
		fail(c, 400, 400, "笔记参数无效")
		return
	}
	n := model.Note{ID: uuid.NewString(), CourseID: in.CourseID, UserID: userID(c), Title: in.Title, Content: in.Content}
	if s.services.DB.Create(&n).Error != nil {
		fail(c, 500, 500, "保存笔记失败")
		return
	}
	s.trackEvent(userID(c), "note_create", "创建了一条笔记", gin.H{"course_id": in.CourseID})
	success(c, gin.H{"note_id": n.ID})
}
func (s *Server) listProblems(c *gin.Context) {
	q := s.services.DB.Model(&model.Problem{})
	for key, col := range map[string]string{"category": "category", "difficulty": "difficulty", "status": "status"} {
		if v := c.Query(key); v != "" && v != "all" && v != "全部" {
			q = q.Where(col+" = ?", v)
		}
	}
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 50)
	var total int64
	q.Count(&total)
	var rows []model.Problem
	q.Order("id").Offset((page - 1) * size).Limit(size).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	solved := 0
	for _, p := range rows {
		if p.Status == "solved" {
			solved++
		}
		items = append(items, gin.H{"id": p.ID, "problem_id": p.ID, "title": p.Title, "category": p.Category, "difficulty": p.Difficulty, "pass_rate": p.PassRate, "passRate": p.PassRate, "submissions": p.Submissions, "status": p.Status, "tags": p.Tags})
	}
	success(c, gin.H{"total": total, "solved": solved, "items": items})
}
func (s *Server) dailyProblem(c *gin.Context) {
	var p model.Problem
	day := int(time.Now().Unix() / 86400)
	var count int64
	s.services.DB.Model(&model.Problem{}).Count(&count)
	if count == 0 {
		fail(c, 404, 404, "暂无题目")
		return
	}
	s.services.DB.Order("id").Offset(day % int(count)).First(&p)
	success(c, gin.H{"problem_id": p.ID, "id": p.ID, "title": p.Title, "difficulty": p.Difficulty, "date": time.Now().Format("2006-01-02")})
}
func (s *Server) problemDetail(c *gin.Context) {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		fail(c, 400, 400, "题目 ID 无效")
		return
	}
	var p model.Problem
	if s.services.DB.First(&p, id).Error != nil {
		fail(c, 404, 404, "题目不存在")
		return
	}
	success(c, gin.H{"id": p.ID, "problem_id": p.ID, "title": p.Title, "category": p.Category, "difficulty": p.Difficulty, "description": p.Description, "examples": p.Examples, "constraints": p.Constraints, "tags": p.Tags, "pass_rate": p.PassRate, "submissions": p.Submissions})
}
func (s *Server) problemTemplates(c *gin.Context) {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		fail(c, 400, 400, "题目 ID 无效")
		return
	}
	var p model.Problem
	if s.services.DB.First(&p, id).Error != nil {
		fail(c, 404, 404, "题目不存在")
		return
	}
	success(c, p.Templates)
}

type codeInput struct {
	ProblemID uint   `json:"problem_id"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

func (s *Server) runCode(c *gin.Context) {
	var in codeInput
	if c.ShouldBindJSON(&in) != nil || in.Language == "" {
		fail(c, 400, 400, "代码参数无效")
		return
	}
	result, err := sandbox.Run(in.Language, in.Code)
	if err != nil {
		fail(c, 500, 1003, "沙箱执行失败")
		return
	}
	success(c, gin.H{"status": result.Status, "test_results": []gin.H{}, "execution_time_ms": result.ExecutionTimeMS, "memory_kb": result.MemoryKB, "stdout": result.Stdout, "stderr": result.Stderr})
}
func (s *Server) submitCode(c *gin.Context) {
	s.submitCodeV2(c)
}
func (s *Server) addActivity(uid, kind, text string) {
	s.services.DB.Create(&model.UserActivity{ID: uuid.NewString(), UserID: uid, Type: kind, Text: text})
}

var _ = http.StatusOK
