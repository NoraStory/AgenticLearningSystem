package api

import (
	"strconv"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (s *Server) register(c *gin.Context) {
	var in authInput
	if c.ShouldBindJSON(&in) != nil || len(in.Username) < 2 || len(in.Password) < 8 || in.Email == "" {
		fail(c, 400, 400, "用户名、邮箱或密码不符合要求（密码至少 8 位）")
		return
	}
	var n int64
	s.services.DB.Model(&model.User{}).Where("email = ? OR username = ?", in.Email, in.Username).Count(&n)
	if n > 0 {
		fail(c, 409, 409, "邮箱或用户名已存在")
		return
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		fail(c, 500, 500, "密码加密失败")
		return
	}
	u := model.User{ID: uuid.NewString(), Username: in.Username, Email: in.Email, PasswordHash: hash, Level: 1, LevelTitle: "初级学习者", LearningDays: 1}
	if err := s.services.DB.Create(&u).Error; err != nil {
		fail(c, 500, 500, "创建用户失败")
		return
	}
	access, _ := s.signToken(u.ID, 24*time.Hour)
	// 7 天免密登录:refresh token 有效期与 DB 记录一致
	refresh, _ := s.signToken(u.ID, 7*24*time.Hour)
	s.services.DB.Create(&model.RefreshToken{ID: uuid.NewString(), UserID: u.ID, TokenHash: tokenHash(refresh), ExpiresAt: time.Now().Add(7 * 24 * time.Hour)})
	success(c, gin.H{"user_id": u.ID, "username": u.Username, "token": access, "refresh_token": refresh, "expires_in": 86400})
}
func (s *Server) login(c *gin.Context) {
	var in authInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "请求格式错误")
		return
	}
	var u model.User
	if err := s.services.DB.Where("email = ?", in.Email).First(&u).Error; err != nil || !verifyPassword(u.PasswordHash, in.Password) {
		fail(c, 401, 401, "邮箱或密码错误")
		return
	}
	// 旧 bcrypt 哈希登录成功后自动升级为 Argon2id,渐进迁移
	if !isArgon2Hash(u.PasswordHash) {
		if upgraded, err := hashPassword(in.Password); err == nil {
			s.services.DB.Model(&u).Update("password_hash", upgraded)
		}
	}
	access, _ := s.signToken(u.ID, 24*time.Hour)
	refresh, _ := s.signToken(u.ID, 7*24*time.Hour)
	s.services.DB.Create(&model.RefreshToken{ID: uuid.NewString(), UserID: u.ID, TokenHash: tokenHash(refresh), ExpiresAt: time.Now().Add(7 * 24 * time.Hour)})
	success(c, gin.H{"user_id": u.ID, "username": u.Username, "token": access, "refresh_token": refresh, "expires_in": 86400})
}
func (s *Server) refresh(c *gin.Context) {
	var in authInput
	if c.ShouldBindJSON(&in) != nil || in.Token == "" {
		fail(c, 400, 400, "缺少刷新令牌")
		return
	}
	var row model.RefreshToken
	if err := s.services.DB.Where("token_hash = ? AND expires_at > ?", tokenHash(in.Token), time.Now()).First(&row).Error; err != nil {
		fail(c, 401, 401, "刷新令牌无效或已过期")
		return
	}
	access, _ := s.signToken(row.UserID, 24*time.Hour)
	success(c, gin.H{"token": access, "expires_in": 86400})
}
func (s *Server) me(c *gin.Context) {
	var u model.User
	if s.services.DB.First(&u, "id = ?", userID(c)).Error != nil {
		fail(c, 404, 404, "用户不存在")
		return
	}
	var totalMinutes int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", u.ID).Select("COALESCE(SUM(duration_minutes),0)").Scan(&totalMinutes)
	var completed, solved int64
	s.services.DB.Model(&model.LearningProgress{}).Where("user_id = ? AND progress >= 100", u.ID).Count(&completed)
	s.services.DB.Model(&model.Submission{}).Where("user_id = ? AND status = ?", u.ID, "accepted").Distinct("problem_id").Count(&solved)
	// 连续打卡:统计最近一次学习中断至今的连续天数
	streak := 0
	var lastStudy time.Time
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", u.ID).Order("study_date desc").Limit(1).Pluck("study_date", &lastStudy)
	if !lastStudy.IsZero() {
		streak = countStreak(s.services.DB, u.ID, lastStudy)
	}
	success(c, gin.H{"user_id": u.ID, "username": u.Username, "email": u.Email, "avatar": u.Avatar, "bio": u.Bio, "level": u.Level, "level_title": u.LevelTitle, "join_date": u.CreatedAt.Format("2006-01-02"), "learning_days": u.LearningDays, "stats": gin.H{"total_hours": float64(totalMinutes) / 60, "completed_courses": completed, "solved_problems": solved, "current_streak": streak}})
}

// countStreak 从最后一次学习日期往前数连续天数。
func countStreak(db *gorm.DB, uid string, from time.Time) int {
	days := 0
	cur := from
	for {
		var cnt int64
		db.Model(&model.DailyStudyTime{}).Where("user_id = ? AND study_date = ?", uid, cur).Count(&cnt)
		if cnt == 0 {
			break
		}
		days++
		cur = cur.AddDate(0, 0, -1)
	}
	return days
}
func (s *Server) updateMe(c *gin.Context) {
	var in struct{ Username, Avatar, Bio string }
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "请求格式错误")
		return
	}
	updates := map[string]any{}
	if in.Username != "" {
		updates["username"] = in.Username
	}
	if in.Avatar != "" {
		updates["avatar"] = in.Avatar
	}
	updates["bio"] = in.Bio
	if err := s.services.DB.Model(&model.User{}).Where("id = ?", userID(c)).Updates(updates).Error; err != nil {
		fail(c, 500, 500, "更新失败")
		return
	}
	s.me(c)
}
func (s *Server) uploadAvatar(c *gin.Context) {
	h, err := c.FormFile("avatar")
	if err != nil {
		fail(c, 400, 400, "请选择头像文件")
		return
	}
	if h.Size > 5<<20 {
		fail(c, 400, 400, "头像不能超过 5MB")
		return
	}
	key, err := s.store.Save(c, "avatars/"+userID(c), h)
	if err != nil {
		fail(c, 500, 1007, "上传失败")
		return
	}
	s.services.DB.Model(&model.User{}).Where("id = ?", userID(c)).Update("avatar", key)
	success(c, gin.H{"avatar": key})
}
func (s *Server) streak(c *gin.Context) {
	uid := userID(c)
	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	// 今日学习时长
	var todayMinutes int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ? AND study_date = ?", uid, today).
		Select("COALESCE(SUM(duration_minutes),0)").Scan(&todayMinutes)

	// 本周打卡（周一到周日）
	weekStart := today.AddDate(0, 0, -(int(now.Weekday())+6)%7)
	weekDays := make([]bool, 7)
	weekTotal := 0
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		var cnt int64
		s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ? AND study_date = ?", uid, day).Count(&cnt)
		weekDays[i] = cnt > 0
		if cnt > 0 {
			weekTotal++
		}
	}

	// 本月打卡（30 天）
	monthDays := make([]bool, 30)
	monthTotal := 0
	for i := 0; i < 30; i++ {
		day := today.AddDate(0, 0, -i)
		var cnt int64
		s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ? AND study_date = ?", uid, day).Count(&cnt)
		monthDays[29-i] = cnt > 0
		if cnt > 0 {
			monthTotal++
		}
	}

	// 真实连续天数
	streakDays := s.computeStreak(uid)

	success(c, gin.H{
		"date":            now.Format("2006-01-02"),
		"today_minutes":   int(todayMinutes),
		"week_days":       weekDays,
		"week_total_days": weekTotal,
		"month_days":      monthDays,
		"month_total_days": monthTotal,
		"streak_days":     streakDays,
	})
}
func (s *Server) progress(c *gin.Context) {
	uid := userID(c)
	// 各方向课程进度:按当前用户的学习进度计算
	var rows []struct {
		Category  string
		Total     int
		Completed int
	}
	s.services.DB.Model(&model.Course{}).
		Select("courses.category, count(*) total, sum(case when lp.progress >= 100 then 1 else 0 end) completed").
		Joins("LEFT JOIN learning_progresses lp ON lp.course_id = courses.id AND lp.user_id = ?", uid).
		Group("courses.category").Scan(&rows)
	data := gin.H{}
	for _, r := range rows {
		pct := 0
		if r.Total > 0 {
			pct = r.Completed * 100 / r.Total
		}
		data[r.Category] = gin.H{"progress": pct, "completed_chapters": r.Completed, "total_chapters": r.Total, "completed_modules": r.Completed, "total_modules": r.Total}
	}
	// 算法题:按当前用户已提交且通过的题目统计
	var total int64
	var solved int64
	s.services.DB.Model(&model.Problem{}).Count(&total)
	s.services.DB.Model(&model.Submission{}).Where("user_id = ? AND status = ?", uid, "accepted").Distinct("problem_id").Count(&solved)
	// 按难度拆分已解决题目
	var easySolved, mediumSolved, hardSolved int64
	s.services.DB.Model(&model.Problem{}).
		Joins("JOIN submissions s ON s.problem_id = problems.id AND s.user_id = ? AND s.status = 'accepted'", uid).
		Where("problems.difficulty = ?", "简单").Distinct("problems.id").Count(&easySolved)
	s.services.DB.Model(&model.Problem{}).
		Joins("JOIN submissions s ON s.problem_id = problems.id AND s.user_id = ? AND s.status = 'accepted'", uid).
		Where("problems.difficulty = ?", "中等").Distinct("problems.id").Count(&mediumSolved)
	s.services.DB.Model(&model.Problem{}).
		Joins("JOIN submissions s ON s.problem_id = problems.id AND s.user_id = ? AND s.status = 'accepted'", uid).
		Where("problems.difficulty = ?", "困难").Distinct("problems.id").Count(&hardSolved)
	data["algorithm"] = gin.H{"progress": int(solved * 100 / max64(total, 1)), "solved_problems": int(solved), "total_problems": int(total), "by_difficulty": gin.H{"easy": gin.H{"solved": int(easySolved), "total": 0}, "medium": gin.H{"solved": int(mediumSolved), "total": 0}, "hard": gin.H{"solved": int(hardSolved), "total": 0}}}
	success(c, data)
}
func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func (s *Server) activities(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit < 1 || limit > 50 {
		limit = 6
	}
	var rows []model.UserActivity
	s.services.DB.Where("user_id = ?", userID(c)).Order("created_at desc").Limit(limit).Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		items = append(items, gin.H{"type": r.Type, "text": r.Text, "time": humanTime(r.CreatedAt), "created_at": r.CreatedAt})
	}
	success(c, gin.H{"items": items})
}
func humanTime(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return strconv.Itoa(int(d.Minutes())) + " 分钟前"
	}
	if d < 24*time.Hour {
		return strconv.Itoa(int(d.Hours())) + " 小时前"
	}
	return strconv.Itoa(int(d.Hours()/24)) + " 天前"
}
func (s *Server) achievements(c *gin.Context) {
	var defs []model.Achievement
	s.services.DB.Find(&defs)
	var unlocked []model.UserAchievement
	s.services.DB.Where("user_id = ?", userID(c)).Find(&unlocked)
	set := map[string]bool{}
	for _, u := range unlocked {
		set[u.AchievementID] = true
	}
	items := make([]gin.H, 0, len(defs))
	for _, a := range defs {
		items = append(items, gin.H{"id": a.ID, "name": a.Name, "desc": a.Description, "icon": a.Icon, "unlocked": set[a.ID]})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) favorites(c *gin.Context) {
	var rows []model.Favorite
	s.services.DB.Where("user_id = ?", userID(c)).Order("created_at desc").Find(&rows)
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.CourseID)
	}
	var courses []model.Course
	if len(ids) > 0 {
		s.services.DB.Where("id IN ?", ids).Find(&courses)
	}
	items := make([]gin.H, 0, len(courses))
	for _, v := range courses {
		items = append(items, gin.H{"id": v.ID, "title": v.Title, "category": v.CategoryLabel})
	}
	success(c, gin.H{"items": items})
}
func (s *Server) notes(c *gin.Context) {
	var rows []model.Note
	s.services.DB.Where("user_id = ?", userID(c)).Order("created_at desc").Find(&rows)
	items := make([]gin.H, 0, len(rows))
	for _, n := range rows {
		var course model.Course
		s.services.DB.First(&course, n.CourseID)
		items = append(items, gin.H{"id": n.ID, "title": n.Title, "course": course.Title, "content": n.Content, "date": n.CreatedAt.Format("2006-01-02")})
	}
	success(c, gin.H{"items": items})
}
func isNotFound(err error) bool { return err == gorm.ErrRecordNotFound }
