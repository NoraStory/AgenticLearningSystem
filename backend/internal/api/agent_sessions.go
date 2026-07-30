package api

import (
	"net/http"
	"strings"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// sessionSummary 是「对话历史」列表里每一项的聚合视图。
// 不单独建一张会话表：session_messages 本身就是持久化的唯一真相来源，
// 这里只是按 session_id 聚合，避免出现两份数据需要同步。
type sessionSummary struct {
	SessionID    string    `json:"session_id"`
	Title        string    `json:"title"`
	Agent        string    `json:"agent"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// agentSessions 返回当前用户的全部会话，按最近活跃倒序。
// 标题取该会话里第一条用户消息；agent 取最近一次非 user 的 agent。
func (s *Server) agentSessions(c *gin.Context) {
	uid := userID(c)
	var rows []sessionSummary
	err := s.services.DB.Raw(`
		SELECT
		  session_id,
		  COALESCE(
		    (array_agg(content ORDER BY created_at) FILTER (WHERE role = 'user'))[1],
		    (array_agg(content ORDER BY created_at))[1],
		    ''
		  ) AS title,
		  COALESCE(
		    (array_agg(agent ORDER BY created_at DESC) FILTER (WHERE agent <> 'user' AND agent <> ''))[1],
		    ''
		  ) AS agent,
		  COUNT(*)::int AS message_count,
		  MIN(created_at) AS created_at,
		  MAX(created_at) AS updated_at
		FROM session_messages
		WHERE user_id = ? AND COALESCE(session_id, '') <> ''
		GROUP BY session_id
		ORDER BY MAX(created_at) DESC
		LIMIT 100
	`, uid).Scan(&rows).Error
	if err != nil {
		fail(c, http.StatusInternalServerError, 500, "会话列表查询失败")
		return
	}
	for i := range rows {
		rows[i].Title = compactMemoryText(rows[i].Title, 60)
	}
	success(c, gin.H{"items": rows})
}

// deleteAgentSession 删除某个会话下的全部消息。
func (s *Server) deleteAgentSession(c *gin.Context) {
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		fail(c, http.StatusBadRequest, 400, "会话 ID 无效")
		return
	}
	result := s.services.DB.Where("user_id = ? AND session_id = ?", userID(c), sid).Delete(&model.SessionMessage{})
	success(c, gin.H{"deleted": true, "session_id": sid, "deleted_count": result.RowsAffected})
}
