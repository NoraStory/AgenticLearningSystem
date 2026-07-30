package api

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"codeforge/backend/internal/model"
)

const (
	conversationMemoryLimit = 12
	conversationPromptLimit = 12000
)

// loadConversationMemory is the durable working memory for an Agent session.
// SessionMessage is already persisted for every turn, so a separate in-memory
// cache would lose context after a restart or browser refresh.
func (s *Server) loadConversationMemory(uid, sessionID string) []model.SessionMessage {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	var rows []model.SessionMessage
	s.services.DB.Where("user_id = ? AND session_id = ?", uid, sessionID).
		Order("created_at desc").Limit(conversationMemoryLimit).Find(&rows)
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].CreatedAt.Before(rows[j].CreatedAt) })
	return rows
}

func isContextualFollowUp(message string) bool {
	text := strings.TrimSpace(strings.ToLower(message))
	if text == "" {
		return false
	}
	markers := []string{
		"详细解释", "详细说", "展开说", "继续", "接着", "上面", "刚才", "前面", "这个", "那个", "它", "这份", "这段", "该提示词", "what about", "explain more", "continue", "this", "that",
	}
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return utf8.RuneCountInString(text) <= 12
}

func lastConversationAgent(history []model.SessionMessage) string {
	for i := len(history) - 1; i >= 0; i-- {
		if strings.TrimSpace(history[i].Agent) != "" {
			return history[i].Agent
		}
	}
	return ""
}

func buildConversationPrompt(message string, history []model.SessionMessage, toolContexts []string) string {
	var b strings.Builder
	if len(history) > 0 {
		b.WriteString("对话记忆（按时间顺序，仅用于理解当前问题）：\n")
		for _, item := range history {
			role := "助手"
			if item.Role == "user" {
				role = "用户"
			}
			content := compactMemoryText(item.Content, 2400)
			if content == "" {
				continue
			}
			fmt.Fprintf(&b, "%s：%s\n", role, content)
		}
		b.WriteString("\n上下文规则：如果当前消息是追问、补充或含有“这个/它/上面/详细解释”等指代，请结合对话记忆直接确定指代对象，不要要求用户重复已经提供的信息。只有当前消息引入了新的明确任务时，才切换主题或调用新工具。\n\n")
	}
	b.WriteString("当前用户消息：\n")
	b.WriteString(strings.TrimSpace(message))
	if len(toolContexts) > 0 {
		b.WriteString("\n\n工具上下文：\n")
		b.WriteString(strings.Join(toolContexts, "\n"))
	}
	return compactMemoryText(b.String(), conversationPromptLimit)
}

func compactMemoryText(text string, max int) string {
	text = strings.TrimSpace(text)
	if max < 1 || len([]rune(text)) <= max {
		return text
	}
	return string([]rune(text)[:max]) + "…"
}

func contextualToolMessage(current string, history []model.SessionMessage) string {
	if len(history) == 0 || !isContextualFollowUp(current) {
		return current
	}
	start := 0
	if len(history) > 4 {
		start = len(history) - 4
	}
	var b strings.Builder
	b.WriteString("???????\n")
	for _, item := range history[start:] {
		role := "??"
		if item.Role == "user" {
			role = "??"
		}
		fmt.Fprintf(&b, "%s?%s\n", role, compactMemoryText(item.Content, 1600))
	}
	fmt.Fprintf(&b, "?????%s", strings.TrimSpace(current))
	return b.String()
}
