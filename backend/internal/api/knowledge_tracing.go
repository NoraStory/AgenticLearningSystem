package api

import (
	"time"

	"codeforge/backend/internal/model"
	"github.com/google/uuid"
)

// BKT 参数（经验默认值，后续可根据数据微调）
const (
	bktPrior      = 0.1  // P(L0) 先验掌握概率
	bktTransition = 0.3  // P(T) 每次练习后学会的概率
	bktGuess      = 0.25 // P(G) 猜对概率
	bktSlip       = 0.1  // P(S) 失误概率
)

// bktUpdate 根据做题结果更新掌握概率。
//   做对 → P(L|correct) = P(L)*(1-P(S)) / (P(L)*(1-P(S)) + (1-P(L))*P(G))
//   做错 → P(L|wrong)   = P(L)*P(S) / (P(L)*P(S) + (1-P(L))*(1-P(G)))
//   练习后 → P(L)' = P(L|result) + (1-P(L|result)) * P(T)
func bktUpdate(prior float64, correct bool) float64 {
	var posterior float64
	if correct {
		posterior = prior * (1 - bktSlip) / (prior*(1-bktSlip) + (1-prior)*bktGuess)
	} else {
		posterior = prior * bktSlip / (prior*bktSlip + (1-prior)*(1-bktGuess))
	}
	// 学习转移
	posterior = posterior + (1-posterior)*bktTransition
	if posterior > 0.99 {
		posterior = 0.99
	}
	if posterior < 0.01 {
		posterior = 0.01
	}
	return posterior
}

// recordKnowledgeState 更新某个用户某个知识点的掌握状态。
// skillName 如 "python-decorator"，category 如 "python"。
func (s *Server) recordKnowledgeState(uid, skillName, category string, correct bool) {
	var ks model.KnowledgeState
	err := s.services.DB.Where("user_id = ? AND skill_name = ?", uid, skillName).First(&ks).Error
	if err != nil {
		ks = model.KnowledgeState{
			ID:        uuid.NewString(),
			UserID:    uid,
			SkillName: skillName,
			Category:  category,
			Mastery:   bktPrior,
		}
	}
	ks.Mastery = bktUpdate(ks.Mastery, correct)
	ks.Attempts++
	if correct {
		ks.CorrectCount++
	}
	ks.LastPracticedAt = time.Now()
	if err != nil {
		s.services.DB.Create(&ks)
	} else {
		s.services.DB.Save(&ks)
	}
}

// computeLevel 根据学习时长、完成课程数、解题数、连续天数自动计算等级。
func computeLevel(totalStudyTime int, completedCourses int, problemSolved int, streak int) (int, string) {
	score := totalStudyTime/10 + completedCourses*3 + problemSolved + streak/2
	switch {
	case score < 10:
		return 1, "新手开发者"
	case score < 30:
		return 3, "入门开发者"
	case score < 80:
		return 5, "初级开发者"
	case score < 200:
		return 7, "中级开发者"
	case score < 500:
		return 9, "高级开发者"
	default:
		return 10, "专家开发者"
	}
}

// updateProfileFromActivity 在用户行为发生后更新画像的统计字段。
func (s *Server) updateProfileFromActivity(uid string) {
	var p model.UserProfile
	if s.services.DB.Where("user_id = ?", uid).First(&p).Error != nil {
		return
	}

	// 真实 Streak：从 daily_study_time 表计算
	streak := s.computeStreak(uid)

	// 总学习时长
	var totalMinutes int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", uid).
		Select("COALESCE(SUM(duration_minutes),0)").Scan(&totalMinutes)

	// 完成课程数
	var completedCourses int64
	s.services.DB.Model(&model.Course{}).Where("status = ?", "completed").Count(&completedCourses)

	// 解题数 + 正确率
	var solved int64
	s.services.DB.Model(&model.Submission{}).Where("user_id = ? AND passed = ?", uid, true).Count(&solved)
	var totalSub int64
	s.services.DB.Model(&model.Submission{}).Where("user_id = ?", uid).Count(&totalSub)
	var accuracy float64
	if totalSub > 0 {
		accuracy = float64(solved) / float64(totalSub)
	}

	// Agent 对话次数
	var sessionCount int64
	s.services.DB.Model(&model.SessionMessage{}).Where("user_id = ? AND role = ?", uid, "assistant").Count(&sessionCount)

	_, levelTitle := computeLevel(int(totalMinutes), int(completedCourses), int(solved), streak)

	s.services.DB.Model(&p).Updates(map[string]any{
		"total_study_time":     int(totalMinutes) / 60,
		"streak":                streak,
		"problem_solved_count":  int(solved),
		"problem_accuracy":      accuracy,
		"session_count":         int(sessionCount),
		"level":                 levelTitle,
		"last_active_at":        time.Now(),
	})
}

// computeStreak 从 daily_study_time 表计算真实连续打卡天数。
func (s *Server) computeStreak(uid string) int {
	var dates []time.Time
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", uid).
		Distinct("study_date").Order("study_date desc").Limit(365).Pluck("study_date", &dates)
	if len(dates) == 0 {
		return 0
	}
	streak := 0
	today := time.Now().Truncate(24 * time.Hour)
	for i, d := range dates {
		expected := today.AddDate(0, 0, -i)
		if d.Truncate(24 * time.Hour).Equal(expected) {
			streak++
		} else {
			break
		}
	}
	return streak
}
