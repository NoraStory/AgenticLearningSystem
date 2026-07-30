package api

import (
	"strings"
	"testing"
	"time"

	"codeforge/backend/internal/model"
)

func TestContextualFollowUpDetection(t *testing.T) {
	for _, message := range []string{"详细解释一下", "这个呢", "继续说", "explain more"} {
		if !isContextualFollowUp(message) {
			t.Errorf("%q should be detected as a follow-up", message)
		}
	}
	if isContextualFollowUp("请搜索 Go 语言最新变化") {
		t.Error("an explicit new search should not be treated as a pure follow-up")
	}
}

func TestConversationPromptUsesHistory(t *testing.T) {
	history := []model.SessionMessage{
		{Role: "user", Agent: "reviewer", Content: "这个提示词是什么？", CreatedAt: time.Now().Add(-time.Minute)},
		{Role: "assistant", Agent: "reviewer", Content: "这是全栈后端工程落地专用的 AI 工程师提示词。", CreatedAt: time.Now()},
	}
	prompt := buildConversationPrompt("详细解释一下", history, nil)
	if !strings.Contains(prompt, "这个提示词是什么") || !strings.Contains(prompt, "全栈后端工程") {
		t.Fatalf("conversation memory missing from prompt: %s", prompt)
	}
	if !strings.Contains(prompt, "指代") || !strings.Contains(prompt, "不要要求用户重复") {
		t.Fatalf("context resolution rule missing from prompt: %s", prompt)
	}
}
