package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"codeforge/backend/internal/config"
)

func TestToolCatalogHasExecutors(t *testing.T) {
	if len(toolCatalog) != 13 {
		t.Fatalf("tool catalog count = %d, want 13", len(toolCatalog))
	}
	for _, tool := range toolCatalog {
		if ok, _ := toolCapability(tool.ID); !ok {
			t.Errorf("tool %s is not workflow-ready", tool.ID)
		}
	}
}

func TestScheduleToolsByPhaseAndDependency(t *testing.T) {
	tools := scheduleTools([]plannedTool{
		{Meta: toolMeta{ID: "self_heal"}, Phase: "analyze"},
		{Meta: toolMeta{ID: "doc_reader"}, Phase: "retrieve"},
		{Meta: toolMeta{ID: "code_execute"}, Phase: "validate"},
		{Meta: toolMeta{ID: "quiz_gen"}, Phase: "generate"},
	})
	if tools[0].Meta.ID != "doc_reader" || tools[len(tools)-1].Meta.ID != "quiz_gen" {
		t.Fatalf("unexpected phase order: %#v", tools)
	}
	if len(tools[2].DependsOn) != 1 || tools[2].Meta.ID != "self_heal" || tools[2].DependsOn[0] != "code_execute" {
		t.Fatalf("self-heal dependency missing: %#v", tools[2])
	}
}

func TestExtractCodeAndSQL(t *testing.T) {
	language, code, ok := extractCode("运行这段代码：\n```python\nprint(1)\n```")
	if !ok || language != "python" || code != "print(1)" {
		t.Fatalf("unexpected code extraction: %q %q %v", language, code, ok)
	}
	sql := extractSQL("请分析 SELECT * FROM users WHERE name LIKE '%x%'")
	if sql == "" {
		t.Fatal("SQL was not extracted")
	}
	result, _ := (&Server{}).explainSQL(sql)
	if result.Available == false || result.Context == "" {
		t.Fatalf("SQL tool unavailable: %#v", result)
	}
}

func TestTextAttachmentDetection(t *testing.T) {
	if !isTextAttachment("main.py", []byte("print(1)")) {
		t.Fatal("python attachment should be text")
	}
	if isTextAttachment("blob.bin", []byte{0, 1, 2}) {
		t.Fatal("binary attachment should not be text")
	}
}

func TestSearxSearchParsesJSONResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" || r.URL.Query().Get("format") != "json" {
			t.Fatalf("unexpected SearXNG request: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"Go documentation","url":"https://go.dev/doc/","content":"Official Go documentation","engine":"google"}]}`))
	}))
	defer server.Close()

	s := &Server{cfg: config.Config{SearchProvider: "searxng", SearchBaseURL: server.URL, SearchTimeoutSeconds: 2}}
	result, err := s.searchWeb(context.Background(), "Go docs")
	if err != nil {
		t.Fatalf("SearXNG search failed: %v", err)
	}
	if !strings.Contains(result, "Go documentation") || !strings.Contains(result, "https://go.dev/doc/") {
		t.Fatalf("unexpected search result: %s", result)
	}
}
