package api

import "testing"

func TestRouteAgent(t *testing.T) {
	cases := map[string]string{"帮我优化简历": "career", "这段代码报错了": "code-review", "解释动态规划题目": "problem-explain", "给我学习计划": "planner"}
	for input, expected := range cases {
		if got := routeAgent(input); got != expected {
			t.Fatalf("%q: got %q want %q", input, got, expected)
		}
	}
}

func TestChunks(t *testing.T) {
	parts := chunks("abcdefgh", 3)
	if len(parts) != 3 || parts[2] != "gh" {
		t.Fatalf("unexpected chunks: %#v", parts)
	}
}
