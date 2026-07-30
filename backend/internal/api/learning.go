package api

import (
	"strconv"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) updateCourseProgress(c *gin.Context) {
	courseID, e := strconv.Atoi(c.Param("id"))
	var in struct {
		Progress      int    `json:"progress"`
		LastSectionID string `json:"last_section_id"`
	}
	if e != nil || c.ShouldBindJSON(&in) != nil || in.Progress < 0 || in.Progress > 100 {
		fail(c, 400, 400, "进度参数无效")
		return
	}
	var row model.LearningProgress
	err := s.services.DB.Where("user_id = ? AND course_id = ?", userID(c), courseID).First(&row).Error
	if err != nil {
		row = model.LearningProgress{ID: uuid.NewString(), UserID: userID(c), CourseID: uint(courseID), Progress: in.Progress, LastSectionID: in.LastSectionID}
		s.services.DB.Create(&row)
	} else {
		s.services.DB.Model(&row).Updates(map[string]any{"progress": in.Progress, "last_section_id": in.LastSectionID})
	}
	s.services.DB.Model(&model.Course{}).Where("id = ?", courseID).Update("progress", in.Progress)
	if in.Progress >= 100 {
		s.services.DB.Model(&model.Course{}).Where("id = ?", courseID).Update("status", "completed")
		s.addActivity(userID(c), "course", "完成了一门课程")
	}
	success(c, gin.H{"progress": in.Progress})
}
func (s *Server) recordTime(c *gin.Context) {
	var in struct {
		DurationMinutes int  `json:"duration_minutes"`
		CourseID        uint `json:"course_id"`
	}
	if c.ShouldBindJSON(&in) != nil || in.DurationMinutes < 1 || in.DurationMinutes > 600 {
		fail(c, 400, 400, "学习时长应为 1-600 分钟")
		return
	}
	day := time.Now().Truncate(24 * time.Hour)
	var row model.DailyStudyTime
	err := s.services.DB.Where("user_id = ? AND course_id = ? AND study_date = ?", userID(c), in.CourseID, day).First(&row).Error
	if err != nil {
		row = model.DailyStudyTime{ID: uuid.NewString(), UserID: userID(c), CourseID: in.CourseID, StudyDate: day, DurationMinutes: in.DurationMinutes}
		s.services.DB.Create(&row)
	} else {
		s.services.DB.Model(&row).Update("duration_minutes", row.DurationMinutes+in.DurationMinutes)
	}
	var total int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", userID(c)).Select("COALESCE(SUM(duration_minutes),0)").Scan(&total)
	s.updateProfileFromActivity(userID(c))
	success(c, gin.H{"total_minutes": total})
}
func (s *Server) learningPaths(c *gin.Context) {
	var rows []model.LearningPath
	s.services.DB.Order("created_at asc").Find(&rows)
	success(c, gin.H{"items": rows})
}
func (s *Server) learningPathStages(c *gin.Context) {
	var rows []model.LearningPathStage
	s.services.DB.Where("path_id = ?", c.Param("id")).Order("sort_order").Find(&rows)
	success(c, gin.H{"items": rows})
}
