package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"codeforge/backend/internal/model"
	"codeforge/backend/internal/sandbox"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// llmGenerateQuestions 用 LLM 根据方向/难度/题数生成笔试题目。LLM 不可用时回退到内置题库。
func (s *Server) llmGenerateQuestions(direction, difficulty string, count int) []model.InterviewQuestion {
	fallback := []model.InterviewQuestion{
		{ID: "q1", Type: "code", Category: "算法", Difficulty: difficulty, Title: "两数之和", Description: "给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值的那两个整数，并返回它们的数组下标。", Constraints: []string{"2 <= nums.length <= 10^4", "-10^9 <= nums[i] <= 10^9", "只会存在一个有效答案"}, TimeLimit: 20, Score: 100, Example: &model.Example{Input: "nums = [2,7,11,15], target = 9", Output: "[0,1]"}},
		{ID: "q2", Type: "text", Category: "系统设计", Difficulty: difficulty, Title: "设计高可用短链接系统", Description: "设计一个 URL 短链接服务，说明 API 设计、数据模型、缓存策略、扩展方案和故障处理。", Constraints: []string{"支持高并发", "考虑数据一致性"}, TimeLimit: 30, Score: 100},
		{ID: "q3", Type: "text", Category: "基础知识", Difficulty: difficulty, Title: "解释事务隔离级别", Description: "请说明数据库常见的事务隔离级别，以及脏读、不可重复读、幻读的区别。", Constraints: []string{"结合实际数据库"}, TimeLimit: 15, Score: 100},
		{ID: "q4", Type: "code", Category: "数据结构", Difficulty: difficulty, Title: "反转链表", Description: "给你单链表的头节点 head，请你反转链表，并返回反转后的链表。", Constraints: []string{"O(1) 额外空间", "链表节点数 [0, 5000]"}, TimeLimit: 15, Score: 100, Example: &model.Example{Input: "head = [1,2,3,4,5]", Output: "[5,4,3,2,1]"}},
	}
	if !s.llm.Enabled() {
		questions := make([]model.InterviewQuestion, 0, count)
		for i := 0; i < count; i++ {
			q := fallback[i%len(fallback)]
			q.ID = fmt.Sprintf("q%d", i+1)
			questions = append(questions, q)
		}
		return questions
	}
	systemPrompt := "你是一个技术笔试出题官。根据给定的方向、难度和题数生成笔试题目。必须返回 JSON 数组，每个元素包含字段：id（如 q1）、type（code 或 text）、category、difficulty、title、description、constraints（数组）、time_limit（分钟）、score（100）、example（可选，含 input/output）。不要返回其他内容。"
	userPrompt := fmt.Sprintf("方向：%s\n难度：%s\n题目数量：%d\n\n请生成 %d 道 %s 方向的 %s 难度笔试题。代码题和文字题混合。", direction, difficulty, count, count, direction, difficulty)
	answer, err := s.llm.Chat(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		questions := make([]model.InterviewQuestion, 0, count)
		for i := 0; i < count; i++ {
			q := fallback[i%len(fallback)]
			q.ID = fmt.Sprintf("q%d", i+1)
			questions = append(questions, q)
		}
		return questions
	}
	answer = strings.TrimSpace(answer)
	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimSuffix(answer, "```")
	answer = strings.TrimSpace(answer)
	var questions []model.InterviewQuestion
	if json.Unmarshal([]byte(answer), &questions) == nil && len(questions) > 0 {
		for i := range questions {
			if questions[i].ID == "" {
				questions[i].ID = fmt.Sprintf("q%d", i+1)
			}
			if questions[i].Score == 0 {
				questions[i].Score = 100
			}
			if questions[i].TimeLimit == 0 {
				questions[i].TimeLimit = 20
			}
		}
		return questions
	}
	questions = make([]model.InterviewQuestion, 0, count)
	for i := 0; i < count; i++ {
		q := fallback[i%len(fallback)]
		q.ID = fmt.Sprintf("q%d", i+1)
		questions = append(questions, q)
	}
	return questions
}

// llmScoreTextAnswer 用 LLM 对文字题评分，返回 0-100 分和反馈。LLM 不可用时回退到字数评分。
func (s *Server) llmScoreTextAnswer(question, answer string) (int, string) {
	if !s.llm.Enabled() {
		score := 40
		if len(strings.TrimSpace(answer)) > 80 {
			score = 75
		}
		if len(strings.TrimSpace(answer)) > 250 {
			score = 90
		}
		comment := "回答偏短，请补充原理、示例与权衡。"
		if score >= 75 {
			comment = "回答结构完整，建议补充边界条件。"
		}
		return score, comment
	}
	systemPrompt := "你是一个技术笔试评分官。请对用户的回答评分（0-100）并给出改进建议。必须返回 JSON：{\"score\": 数字, \"feedback\": \"评语\"}。不要返回其他内容。"
	userPrompt := fmt.Sprintf("题目：%s\n\n用户回答：%s\n\n请评分并给出改进建议。", question, answer)
	result, err := s.llm.Chat(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return 50, "评分服务暂时不可用，已按默认分评分。"
	}
	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)
	var parsed struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	}
	if json.Unmarshal([]byte(result), &parsed) == nil && parsed.Score >= 0 && parsed.Score <= 100 {
		return parsed.Score, parsed.Feedback
	}
	return 50, "评分解析失败，已按默认分评分。"
}

// generateExamV2 用 LLM 出题生成笔试。
func (s *Server) generateExamV2(c *gin.Context) {
	var in struct {
		Direction     string `json:"direction"`
		Difficulty    string `json:"difficulty"`
		QuestionCount int    `json:"question_count"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "生成参数无效")
		return
	}
	if in.QuestionCount < 1 {
		in.QuestionCount = 4
	}
	if in.QuestionCount > 10 {
		in.QuestionCount = 10
	}
	if in.Direction == "" {
		in.Direction = "全栈开发"
	}
	if in.Difficulty == "" {
		in.Difficulty = "中等"
	}
	questions := s.llmGenerateQuestions(in.Direction, in.Difficulty, in.QuestionCount)
	exam := model.InterviewExam{ID: uuid.NewString(), UserID: userID(c), Direction: in.Direction, Difficulty: in.Difficulty, Questions: questions}
	s.services.DB.Create(&exam)
	success(c, gin.H{"exam_id": exam.ID, "questions": questionPayload(questions)})
}

// runExamQuestionV2 支持测试用例验证的代码运行。
func (s *Server) runExamQuestionV2(c *gin.Context) {
	var in codeInput
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "代码参数无效")
		return
	}
	if in.ProblemID > 0 {
		var prob model.Problem
		if s.services.DB.First(&prob, in.ProblemID).Error == nil && len(prob.TestCases) > 0 {
			cases := make([]sandbox.TestCase, 0, len(prob.TestCases))
			for _, tc := range prob.TestCases {
				cases = append(cases, sandbox.TestCase{Input: tc.Input, Expected: tc.Expected})
			}
			cr, err := sandbox.ValidateWithCases(in.Language, in.Code, cases)
			if err != nil {
				fail(c, 500, 1003, "沙箱执行失败")
				return
			}
			caseResults := make([]gin.H, 0, len(cr.Results))
			for _, r := range cr.Results {
				caseResults = append(caseResults, gin.H{"input": r.Input, "expected": r.Expected, "actual": r.Actual, "passed": r.Passed, "error": r.Error, "time_ms": r.TimeMS})
			}
			success(c, gin.H{"status": cr.Status, "case_results": caseResults, "passed_cases": cr.PassedCount, "total_cases": cr.TotalCount, "execution_time_ms": cr.TimeMS, "stdout": cr.Stdout, "stderr": cr.Stderr})
			return
		}
	}
	r, err := sandbox.Run(in.Language, in.Code)
	if err != nil {
		fail(c, 500, 1003, "沙箱执行失败")
		return
	}
	success(c, gin.H{"status": r.Status, "test_results": []gin.H{}, "execution_time_ms": r.ExecutionTimeMS, "stdout": r.Stdout, "stderr": r.Stderr})
}

// submitExamV2 用 AI 评分 + 代码测试验证的笔试提交，返回逐题反馈。
func (s *Server) submitExamV2(c *gin.Context) {
	var in struct {
		Answers []struct {
			QuestionID string `json:"question_id"`
			Answer     string `json:"answer"`
			Language   string `json:"language"`
		} `json:"answers"`
	}
	if c.ShouldBindJSON(&in) != nil {
		fail(c, 400, 400, "答案格式错误")
		return
	}
	var e model.InterviewExam
	if s.services.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID(c)).First(&e).Error != nil {
		fail(c, 404, 404, "笔试不存在")
		return
	}
	questionMap := make(map[string]model.InterviewQuestion)
	for _, q := range e.Questions {
		questionMap[q.ID] = q
	}
	feedback := make([]gin.H, 0, len(in.Answers))
	total := 0
	for _, a := range in.Answers {
		q, ok := questionMap[a.QuestionID]
		score := 0
		comment := ""
		if !ok {
			score = 0
			comment = "题目不存在"
		} else if q.Type == "code" && q.Example != nil && a.Language != "" {
			cases := []sandbox.TestCase{{Input: q.Example.Input, Expected: q.Example.Output}}
			cr, err := sandbox.ValidateWithCases(a.Language, a.Answer, cases)
			if err == nil && cr.TotalCount > 0 {
				score = cr.PassedCount * q.Score / cr.TotalCount
				if score >= 100 {
					comment = "全部测试用例通过，答案正确！"
				} else {
					comment = fmt.Sprintf("通过 %d/%d 个测试用例。", cr.PassedCount, cr.TotalCount)
				}
			} else {
				r, runErr := sandbox.Run(a.Language, a.Answer)
				if runErr == nil && r.Status == "success" {
					score = 60
					comment = "代码运行成功但输出与期望不符。"
				} else if runErr == nil {
					score = 40
					comment = "代码运行出错：" + r.Stderr
				} else {
					score = 20
					comment = "沙箱执行失败。"
				}
			}
		} else {
			score, comment = s.llmScoreTextAnswer(q.Description, a.Answer)
		}
		total += score
		feedback = append(feedback, gin.H{
			"question_id": a.QuestionID,
			"title":       q.Title,
			"type":        q.Type,
			"score":       score,
			"max_score":   q.Score,
			"feedback":    comment,
		})
	}
	final := 0
	if len(in.Answers) > 0 {
		final = total / len(in.Answers)
	}
	s.services.DB.Model(&e).Update("score", final)
	s.trackEvent(userID(c), "interview_submit", "完成了一次笔试", gin.H{"direction": e.Direction, "difficulty": e.Difficulty, "score": final})
	success(c, gin.H{"score": final, "feedback": feedback})
}