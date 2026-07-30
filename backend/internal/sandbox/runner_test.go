package sandbox

import (
	"strings"
	"testing"
)

func TestRunPython(t *testing.T) {
	result, err := Run("python", "print(1 + 1)")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "success" {
		t.Fatalf("status=%s stderr=%s", result.Status, result.Stderr)
	}
	if strings.TrimSpace(result.Stdout) != "2" {
		t.Fatalf("stdout=%q", result.Stdout)
	}
}

func TestRejectDangerousCode(t *testing.T) {
	result, err := Run("python", "import os\nos.system('echo unsafe')")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "rejected" {
		t.Fatalf("expected rejected, got %s", result.Status)
	}
}
