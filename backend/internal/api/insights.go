package api

import (
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// agentInsights 学习洞察：从 UserActivity + daily_study_times + submissions 聚合
// 当前用户的宏观学习行为。游客（Demo 用户）无数据时返回空结构，前端显示空态。
func (s *Server) agentInsights(c *gin.Context) {
	uid := userID(c)
	now := time.Now()

	// ---- 1. 行为分布（近 30 天，按事件类型计数）----
	type actRow struct {
		Type string
		Cnt  int
	}
	var acts []actRow
	s.services.DB.Model(&model.UserActivity{}).
		Where("user_id = ? AND created_at > ?", uid, now.AddDate(0, 0, -30)).
		Select("type, count(*) cnt").Group("type").Scan(&acts)
	behavior := make([]gin.H, 0, len(acts))
	for _, a := range acts {
		behavior = append(behavior, gin.H{"name": a.Type, "value": a.Cnt})
	}

	// ---- 2. 时段偏好（学习时长按小时分布）----
	// daily_study_times.study_date 是 date 类型（无时刻），用记录创建时间 created_at 近似学习时段
	type slotRow struct {
		Hour int
		Min  int
	}
	var slots []slotRow
	s.services.DB.Raw(`
		SELECT extract(hour from created_at)::int AS hour, SUM(duration_minutes)::int AS min
		FROM daily_study_times WHERE user_id = ? AND created_at > ?
		GROUP BY 1 ORDER BY 1`, uid, now.AddDate(0, 0, -30)).Scan(&slots)
	hours := make([]gin.H, 24)
	for i := range hours {
		hours[i] = gin.H{"hour": i, "minutes": 0}
	}
	for _, sl := range slots {
		if sl.Hour >= 0 && sl.Hour < 24 {
			hours[sl.Hour] = gin.H{"hour": sl.Hour, "minutes": sl.Min}
		}
	}

	// ---- 3. 兴趣分布（各 category 浏览/做题/学习时长占比）----
	var interests []struct {
		Category string
		Cnt      int
	}
	s.services.DB.Raw(`
		SELECT metadata->>'category' AS category, count(*) AS cnt
		FROM user_activities
		WHERE user_id = ? AND type = 'course_view' AND metadata->>'category' IS NOT NULL
		GROUP BY 1 ORDER BY cnt DESC`, uid).Scan(&interests)
	if interests == nil {
		interests = []struct {
			Category string
			Cnt      int
		}{}
	}

	// ---- 4. 近 14 天趋势（每日事件数 + 学习时长）----
	type dayAct struct {
		Day time.Time
		Cnt int
	}
	var dayActs []dayAct
	s.services.DB.Raw(`
		SELECT date_trunc('day', created_at)::date AS day, count(*) AS cnt
		FROM user_activities WHERE user_id = ? AND created_at > ?
		GROUP BY 1`, uid, now.AddDate(0, 0, -14)).Scan(&dayActs)
	dayMap := map[string]int{}
	for _, d := range dayActs {
		dayMap[d.Day.Format("2006-01-02")] = d.Cnt
	}
	var dayTime []struct {
		Day time.Time
		Min int
	}
	s.services.DB.Raw(`
		SELECT study_date AS day, SUM(duration_minutes)::int AS min
		FROM daily_study_times WHERE user_id = ? AND study_date > ?
		GROUP BY 1`, uid, now.AddDate(0, 0, -14)).Scan(&dayTime)
	timeMap := map[string]int{}
	for _, d := range dayTime {
		timeMap[d.Day.Format("2006-01-02")] = d.Min
	}
	trend := make([]gin.H, 0, 14)
	for i := 13; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		key := day.Format("2006-01-02")
		trend = append(trend, gin.H{"date": key, "events": dayMap[key], "minutes": timeMap[key]})
	}

	// ---- 5. 一致性：总学习时长/最活跃星期几/近 7 天活跃天数 ----
	var totalMin int64
	s.services.DB.Model(&model.DailyStudyTime{}).Where("user_id = ?", uid).
		Select("COALESCE(SUM(duration_minutes),0)").Scan(&totalMin)
	weekAct := map[string]int{}
	for _, d := range dayTime {
		w := int(d.Day.Weekday())
		if w == 0 {
			w = 7
		}
		weekAct[weekdayCN[w]] += d.Min
	}
	activeDays := 0
	for i := 6; i >= 0; i-- {
		if timeMap[now.AddDate(0, 0, -i).Format("2006-01-02")] > 0 {
			activeDays++
		}
	}
	bestDay := ""
	bestMin := 0
	for k, v := range weekAct {
		if v > bestMin {
			bestDay, bestMin = k, v
		}
	}
	consistency := gin.H{
		"total_minutes":     int(totalMin),
		"avg_daily_minutes": totalMin / 30,
		"active_days_7":     activeDays,
		"best_weekday":      bestDay,
	}
	if bestDay == "" {
		consistency["best_weekday"] = "暂无数据"
	}

	success(c, gin.H{
		"behavior_distribution": behavior,
		"time_slot_preference":  hours,
		"interest_distribution": interests,
		"activity_trend":        trend,
		"consistency":           consistency,
	})
}

var weekdayCN = map[int]string{
	1: "周一", 2: "周二", 3: "周三", 4: "周四", 5: "周五", 6: "周六", 7: "周日",
}
