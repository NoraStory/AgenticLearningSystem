package api

import (
	"sort"
	"strconv"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// weakThreshold 掌握度低于该值视为弱项方向。
const weakThreshold = 0.45

// categoryLabelCN 后端分类 key → 中文展示名（前端也有同义映射）。
var categoryLabelCN = map[string]string{
	"python": "Python", "cpp": "C++", "database": "数据库", "algorithm": "算法", "agent": "AI Agent",
}

// weakAreaRecommendations 弱项驱动的课程推荐。
// 逻辑：按当前用户 knowledge_states 聚合各方向平均掌握度，弱项方向（< 45%）
// 取该方向未完成课程补强；无画像数据（游客/新用户）时回退热门榜，结构与
// /courses/recommended 保持一致（{id,title,category}），首页永不空。
func (s *Server) weakAreaRecommendations(c *gin.Context) {
	uid := userID(c)

	// 1. 聚合各方向平均掌握度
	var states []model.KnowledgeState
	s.services.DB.Where("user_id = ?", uid).Limit(200).Find(&states)
	catMastery := map[string]float64{}
	catCount := map[string]int{}
	for _, st := range states {
		catMastery[st.Category] += st.Mastery
		catCount[st.Category]++
	}
	type catAvg struct {
		key     string
		mastery float64
	}
	var weak []catAvg
	for k, sum := range catMastery {
		if catCount[k] == 0 {
			continue
		}
		avg := sum / float64(catCount[k])
		if avg < weakThreshold {
			weak = append(weak, catAvg{k, avg})
		}
	}
	sort.Slice(weak, func(i, j int) bool { return weak[i].mastery < weak[j].mastery })

	// 2. 无画像数据 → 热门榜兜底（保持原结构）
	if len(weak) == 0 {
		s.recommendedCourses(c)
		return
	}

	// 3. 弱项方向取课程，排除已完成的
	items := make([]gin.H, 0, 4)
	seen := map[uint]bool{}
	for _, w := range weak {
		var courses []model.Course
		s.services.DB.Where("category = ?", w.key).
			Where("id NOT IN (SELECT course_id FROM learning_progresses WHERE user_id = ? AND progress >= 100)", uid).
			Order("views desc").Limit(2).Find(&courses)
		for _, cr := range courses {
			if seen[cr.ID] {
				continue
			}
			seen[cr.ID] = true
			items = append(items, gin.H{
				"id":          cr.ID,
				"title":       cr.Title,
				"category":    cr.CategoryLabel,
				"category_key": cr.Category,
				"reason":      weakReason(w.key, w.mastery),
			})
			if len(items) >= 4 {
				break
			}
		}
		if len(items) >= 4 {
			break
		}
	}

	// 4. 不够 4 门补热门
	if len(items) < 4 {
		var hot []model.Course
		s.services.DB.Where("id NOT IN ?", mapKeys(seen)).Order("views desc").Limit(4 - len(items)).Find(&hot)
		for _, cr := range hot {
			items = append(items, gin.H{"id": cr.ID, "title": cr.Title, "category": cr.CategoryLabel, "category_key": cr.Category})
		}
	}
	success(c, gin.H{"items": items})
}

func weakReason(category string, mastery float64) string {
	label := category
	if v, ok := categoryLabelCN[category]; ok {
		label = v
	}
	return label + "方向掌握度 " + strconv.Itoa(int(mastery*100)) + "%，建议优先补强"
}

func mapKeys(m map[uint]bool) []uint {
	out := make([]uint, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
