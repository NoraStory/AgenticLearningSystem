package api

import (
	"encoding/json"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxEventsPerBatch 单次上报事件条数上限，防滥用。
const maxEventsPerBatch = 20

// trackEvent 写入一条埋点事件（UserActivity 表）。
// kind 为事件类型（如 page_view / search / problem_start / course_view / bookmark / like / note_create / interview_submit）。
// props 为扁平键值对，序列化进 metadata（jsonb）。text 用于时间线展示。
func (s *Server) trackEvent(uid, kind, text string, props map[string]any) {
	raw := ""
	if len(props) > 0 {
		if b, err := json.Marshal(props); err == nil {
			raw = string(b)
		}
	}
	s.services.DB.Create(&model.UserActivity{ID: uuid.NewString(), UserID: uid, Type: kind, Text: text, Verb: kind, Metadata: raw})
}

// collectEvents 前端埋点统一入口。
// body: {"events":[{"name":"search","props":{"query":"decorator"}}, ...]}（≤20 条）
// 游客自动落 DemoUserID；登录用户带 token 自动关联。
func (s *Server) collectEvents(c *gin.Context) {
	var in struct {
		Events []struct {
			Name  string         `json:"name"`
			Props map[string]any `json:"props"`
		} `json:"events"`
	}
	if c.ShouldBindJSON(&in) != nil || len(in.Events) == 0 {
		fail(c, 400, 400, "事件参数无效")
		return
	}
	if len(in.Events) > maxEventsPerBatch {
		in.Events = in.Events[:maxEventsPerBatch]
	}
	uid := userID(c)
	for _, ev := range in.Events {
		if ev.Name == "" {
			continue
		}
		s.trackEvent(uid, ev.Name, ev.Name, ev.Props)
	}
	success(c, gin.H{"accepted": len(in.Events)})
}
