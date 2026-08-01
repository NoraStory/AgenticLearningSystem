package api

import "testing"

func TestApplyChangesRewrite(t *testing.T) {
	lines := []string{"姓名: 张三", "工作经历: 开发了平台"}
	changes := []resumeChange{
		{Path: "sections.0.items.1", Action: "rewrite", NewValue: "工作经历: 开发了高并发平台,吞吐提升 40%"},
	}
	out, dropped := applyChanges(lines, changes)
	if len(out) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(out), out)
	}
	if out[1] != "工作经历: 开发了高并发平台,吞吐提升 40%" {
		t.Fatalf("rewrite not applied: %q", out[1])
	}
	if dropped != 0 {
		t.Fatalf("want 0 dropped, got %d", dropped)
	}
}

func TestApplyChangesLockedFieldDropped(t *testing.T) {
	// 锁定字段:姓名/邮箱/电话/日期 所在行禁止修改
	lines := []string{"姓名: 张三", "电话: 13800000000", "工作经历: 开发了平台"}
	changes := []resumeChange{
		{Path: "sections.0.items.0", Action: "rewrite", NewValue: "姓名: 李四"},
		{Path: "sections.0.items.1", Action: "rewrite", NewValue: "电话: 99999999999"},
		{Path: "sections.0.items.2", Action: "rewrite", NewValue: "工作经历: 主导了核心项目"},
	}
	out, dropped := applyChanges(lines, changes)
	if dropped != 2 {
		t.Fatalf("want 2 dropped (name+phone), got %d", dropped)
	}
	if out[0] != "姓名: 张三" || out[1] != "电话: 13800000000" {
		t.Fatalf("locked fields were modified: %v", out)
	}
	if out[2] != "工作经历: 主导了核心项目" {
		t.Fatalf("valid rewrite not applied: %q", out[2])
	}
}

func TestApplyChangesDeleteAndInsert(t *testing.T) {
	lines := []string{"a", "b", "c"}
	changes := []resumeChange{
		{Path: "sections.0.items.1", Action: "delete"},
		{Path: "sections.0.items.2", Action: "insert", NewValue: "d"},
	}
	out, _ := applyChanges(lines, changes)
	want := []string{"a", "c", "d"}
	if len(out) != len(want) {
		t.Fatalf("want %v, got %v", want, out)
	}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("want %v, got %v", want, out)
		}
	}
}

func TestApplyChangesInvalidPath(t *testing.T) {
	lines := []string{"a", "b"}
	changes := []resumeChange{
		{Path: "sections.9.items.99", Action: "rewrite", NewValue: "x"}, // 越界
		{Path: "badpath", Action: "rewrite", NewValue: "y"},             // 非法格式
		{Path: "sections.0.items.0", Action: "unknown", NewValue: "z"},  // 非法 action
	}
	out, dropped := applyChanges(lines, changes)
	if dropped != 3 {
		t.Fatalf("want 3 dropped, got %d", dropped)
	}
	if len(out) != 2 || out[0] != "a" {
		t.Fatalf("lines should be unchanged: %v", out)
	}
}

func TestParsePath(t *testing.T) {
	cases := []struct {
		path string
		idx  int
		ok   bool
	}{
		{"sections.0.items.0", 0, true},
		{"sections.2.items.5", 5, true},
		{"sections.0.items.", 0, false},
		{"sections.0", 0, false},
		{"items.0", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		idx, ok := parsePath(c.path)
		if ok != c.ok || (ok && idx != c.idx) {
			t.Fatalf("parsePath(%q) = %d,%v want %d,%v", c.path, idx, ok, c.idx, c.ok)
		}
	}
}

func TestContainsLocked(t *testing.T) {
	locked := []string{"姓名: 张三", "邮箱: a@b.com", "电话: 123", "2022.03 至今"}
	for _, l := range locked {
		if !containsLocked(l) {
			t.Fatalf("expected %q locked", l)
		}
	}
	free := []string{"工作经历: 开发了平台", "技能: Go 语言"}
	for _, f := range free {
		if containsLocked(f) {
			t.Fatalf("expected %q not locked", f)
		}
	}
}
