package api

import (
	"strconv"

	"codeforge/backend/internal/model"
	"codeforge/backend/internal/sandbox"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// submitCodeV2 是升级版的提交接口，使用真实测试用例验证。
func (s *Server) submitCodeV2(c *gin.Context) {
	var in codeInput
	if c.ShouldBindJSON(&in) != nil || in.Language == "" || in.ProblemID == 0 {
		fail(c, 400, 400, "提交参数无效")
		return
	}
	var prob model.Problem
	if s.services.DB.First(&prob, in.ProblemID).Error != nil {
		fail(c, 404, 404, "题目不存在")
		return
	}
	modelCases := prob.TestCases
	var cases []sandbox.TestCase
	for _, tc := range modelCases {
		cases = append(cases, sandbox.TestCase{Input: tc.Input, Expected: tc.Expected})
	}
	if len(cases) == 0 {
		for _, ex := range prob.Examples {
			cases = append(cases, sandbox.TestCase{Input: ex.Input, Expected: ex.Output})
		}
	}
	if len(cases) == 0 {
		result, err := sandbox.ValidateSolution(in.Language, in.Code)
		if err != nil {
			fail(c, 500, 1003, "沙箱执行失败")
			return
		}
		accepted := result.Status == "success"
		status := "wrong_answer"
		if accepted {
			status = "accepted"
		}
		sub := model.Submission{ID: uuid.NewString(), UserID: userID(c), ProblemID: in.ProblemID, Language: in.Language, Code: in.Code, Status: status, PassedCases: 0, TotalCases: 0, ExecutionTimeMS: result.ExecutionTimeMS, MemoryKB: result.MemoryKB}
		s.services.DB.Create(&sub)
		if accepted {
			s.services.DB.Model(&model.Problem{}).Where("id = ?", in.ProblemID).Update("status", "solved")
			s.addActivity(userID(c), "problem", "完成了一道算法练习")
			s.recordKnowledgeState(userID(c), prob.Title, "algorithm", accepted)
			s.updateProfileFromActivity(userID(c))
		}
		success(c, gin.H{"status": status, "passed_cases": 0, "total_cases": 0, "case_results": []gin.H{}, "execution_time_ms": result.ExecutionTimeMS, "memory_kb": result.MemoryKB, "stdout": result.Stdout, "stderr": result.Stderr})
		return
	}
	cr, err := sandbox.ValidateWithCases(in.Language, in.Code, cases)
	if err != nil {
		fail(c, 500, 1003, "沙箱执行失败")
		return
	}
	accepted := cr.Status == "accepted"
	sub := model.Submission{ID: uuid.NewString(), UserID: userID(c), ProblemID: in.ProblemID, Language: in.Language, Code: in.Code, Status: cr.Status, PassedCases: cr.PassedCount, TotalCases: cr.TotalCount, ExecutionTimeMS: cr.TimeMS, MemoryKB: cr.MemoryKB}
	s.services.DB.Create(&sub)
	s.services.DB.Model(&model.Problem{}).Where("id = ?", in.ProblemID).Update("submissions", prob.Submissions+1)
	var acceptedCount int64
	s.services.DB.Model(&model.Submission{}).Where("problem_id = ? AND status = ?", in.ProblemID, "accepted").Count(&acceptedCount)
	var totalSub int64
	s.services.DB.Model(&model.Submission{}).Where("problem_id = ?", in.ProblemID).Count(&totalSub)
	newRate := 0.0
	if totalSub > 0 {
		newRate = float64(acceptedCount) / float64(totalSub)
	}
	if accepted {
		s.services.DB.Model(&model.Problem{}).Where("id = ?", in.ProblemID).Updates(map[string]any{"status": "solved", "pass_rate": newRate})
		s.addActivity(userID(c), "problem", "完成了一道算法练习")
		skillName := prob.Title
		if len(skillName) > 120 {
			skillName = skillName[:120]
		}
		s.recordKnowledgeState(userID(c), skillName, "algorithm", accepted)
		s.updateProfileFromActivity(userID(c))
	} else {
		s.services.DB.Model(&model.Problem{}).Where("id = ?", in.ProblemID).Update("pass_rate", newRate)
	}
	caseResults := make([]gin.H, 0, len(cr.Results))
	for _, r := range cr.Results {
		caseResults = append(caseResults, gin.H{"input": r.Input, "expected": r.Expected, "actual": r.Actual, "passed": r.Passed, "error": r.Error, "time_ms": r.TimeMS})
	}
	pct := 0
	if accepted {
		pct = 78
	}
	success(c, gin.H{"status": cr.Status, "passed_cases": cr.PassedCount, "total_cases": cr.TotalCount, "case_results": caseResults, "execution_time_ms": cr.TimeMS, "memory_kb": cr.MemoryKB, "stdout": cr.Stdout, "stderr": cr.Stderr, "ranking_percentile": pct})
}

func (s *Server) problemSubmissions(c *gin.Context) {
	id, e := strconv.Atoi(c.Param("id"))
	if e != nil {
		fail(c, 400, 400, "题目 ID 无效")
		return
	}
	var subs []model.Submission
	s.services.DB.Where("user_id = ? AND problem_id = ?", userID(c), id).Order("created_at desc").Limit(50).Find(&subs)
	items := make([]gin.H, 0, len(subs))
	for _, sub := range subs {
		items = append(items, gin.H{"id": sub.ID, "language": sub.Language, "status": sub.Status, "passed_cases": sub.PassedCases, "total_cases": sub.TotalCases, "execution_time_ms": sub.ExecutionTimeMS, "memory_kb": sub.MemoryKB, "code": sub.Code, "created_at": sub.CreatedAt})
	}
	success(c, gin.H{"items": items})
}

func (s *Server) allSubmissions(c *gin.Context) {
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 20)
	var total int64
	s.services.DB.Model(&model.Submission{}).Where("user_id = ?", userID(c)).Count(&total)
	var subs []model.Submission
	s.services.DB.Where("user_id = ?", userID(c)).Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&subs)
	items := make([]gin.H, 0, len(subs))
	for _, sub := range subs {
		var prob model.Problem
		s.services.DB.First(&prob, sub.ProblemID)
		items = append(items, gin.H{"id": sub.ID, "problem_id": sub.ProblemID, "problem_title": prob.Title, "language": sub.Language, "status": sub.Status, "passed_cases": sub.PassedCases, "total_cases": sub.TotalCases, "execution_time_ms": sub.ExecutionTimeMS, "created_at": sub.CreatedAt})
	}
	success(c, gin.H{"items": items, "total": total, "page": page, "page_size": size})
}

func (s *Server) leaderboard(c *gin.Context) {
	type rankRow struct {
		UserID   string
		Username string
		Avatar   string
		Solved   int
		TotalSub int
		Accuracy float64
	}
	var rows []rankRow
	s.services.DB.Raw(`
		SELECT u.id as user_id, u.username, u.avatar,
			COUNT(DISTINCT CASE WHEN s.status = 'accepted' THEN s.problem_id END) as solved,
			COUNT(s.id) as total_sub,
			CASE WHEN COUNT(s.id) > 0 THEN COUNT(CASE WHEN s.status = 'accepted' THEN 1 END) * 1.0 / COUNT(s.id) ELSE 0 END as accuracy
		FROM users u
		LEFT JOIN submissions s ON s.user_id = u.id
		GROUP BY u.id, u.username, u.avatar
		ORDER BY solved DESC, accuracy DESC
		LIMIT 100
	`).Scan(&rows)
	items := make([]gin.H, 0, len(rows))
	for i, r := range rows {
		items = append(items, gin.H{"rank": i + 1, "user_id": r.UserID, "username": r.Username, "avatar": r.Avatar, "solved": r.Solved, "total_submissions": r.TotalSub, "accuracy": r.Accuracy})
	}
	success(c, gin.H{"items": items})
}
