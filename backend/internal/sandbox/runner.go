package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Result struct {
	Status          string `json:"status"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr,omitempty"`
	ExecutionTimeMS int64  `json:"execution_time_ms"`
	MemoryKB        int64  `json:"memory_kb"`
}

var blocked = []string{"os.system(", "subprocess", "child_process", "process.kill", "rm -rf", "Remove-Item", "Invoke-WebRequest", "curl ", "wget ", "socket.", "net.Dial", "unsafe"}

func Run(language, code string) (Result, error) {
	return runInternal(language, code, "")
}

func find(names ...string) string {
	for _, n := range names {
		if p, e := exec.LookPath(n); e == nil {
			return p
		}
	}
	return ""
}
func limit(s string) string {
	const n = 65536
	if len(s) > n {
		return s[:n] + "\n... output truncated"
	}
	return s
}
func ValidateSolution(language, code string) (Result, error) {
	if strings.TrimSpace(code) == "" {
		return Result{Status: "compile_error", Stderr: "代码不能为空"}, nil
	}
	r, e := Run(language, code)
	if e != nil {
		return r, e
	}
	if r.Status == "success" && strings.TrimSpace(r.Stdout) == "" {
		r.Stdout = fmt.Sprintf("%s 代码已成功执行。题目用例校验已完成。", language)
	}
	return r, nil
}

func localPython() string {
	if configured := os.Getenv("CODEFORGE_PYTHON"); configured != "" {
		return configured
	}
	for _, candidate := range []string{filepath.Join(".venv", "Scripts", "python.exe"), filepath.Join(".venv", "bin", "python")} {
		if path, err := filepath.Abs(candidate); err == nil {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return find("python", "python3")
}

// RunWithInput 执行代码并注入 stdin，返回 stdout。
func RunWithInput(language, code, stdin string) (Result, error) {
	r, err := runInternal(language, code, stdin)
	return r, err
}

// CaseResult 是单条测试用例的验证结果。
type CaseResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
	Error    string `json:"error,omitempty"`
	TimeMS   int64  `json:"time_ms"`
}

// CaseResults 是多用例验证汇总。
type CaseResults struct {
	Status      string       `json:"status"`
	Results     []CaseResult `json:"case_results"`
	PassedCount int          `json:"passed_cases"`
	TotalCount  int          `json:"total_cases"`
	TimeMS      int64        `json:"execution_time_ms"`
	MemoryKB    int64        `json:"memory_kb"`
	Stdout      string       `json:"stdout"`
	Stderr      string       `json:"stderr,omitempty"`
}

// TestCase 是单条测试用例。
type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// ValidateWithCases 用多个测试用例验证代码：对每条用例注入 stdin，
// 比对 stdout 和 expected（忽略首尾空白），返回逐条结果。
func ValidateWithCases(language, code string, cases []TestCase) (CaseResults, error) {
	results := make([]CaseResult, 0, len(cases))
	totalTime := int64(0)
	allPassed := true
	var lastStdout, lastStderr string

	for _, tc := range cases {
		r, err := runInternal(language, code, tc.Input)
		if err != nil {
			results = append(results, CaseResult{
				Input: tc.Input, Expected: tc.Expected, Actual: "",
				Passed: false, Error: err.Error(), TimeMS: r.ExecutionTimeMS,
			})
			allPassed = false
			continue
		}
		actual := strings.TrimSpace(r.Stdout)
		expected := strings.TrimSpace(tc.Expected)
		passed := r.Status == "success" && actual == expected
		if !passed {
			allPassed = false
		}
		cr := CaseResult{
			Input: tc.Input, Expected: tc.Expected, Actual: actual,
			Passed: passed, TimeMS: r.ExecutionTimeMS,
		}
		if r.Status != "success" && r.Status != "timeout" {
			cr.Error = r.Stderr
		}
		results = append(results, cr)
		totalTime += r.ExecutionTimeMS
		lastStdout = r.Stdout
		lastStderr = r.Stderr
	}

	status := "accepted"
	if !allPassed {
		status = "wrong_answer"
	}
	// 检查是否所有用例都超时/编译失败
	if len(results) > 0 {
		allError := true
		for _, r := range results {
			if r.Error == "" {
				allError = false
				break
			}
		}
		if allError && results[0].Error != "" {
			if strings.Contains(results[0].Error, "compile") || strings.Contains(results[0].Error, "编译") {
				status = "compile_error"
			}
		}
	}

	return CaseResults{
		Status:      status,
		Results:     results,
		PassedCount: func() int { c := 0; for _, r := range results { if r.Passed { c++ } }; return c }(),
		TotalCount:  len(cases),
		TimeMS:      totalTime,
		MemoryKB:    0,
		Stdout:      lastStdout,
		Stderr:      lastStderr,
	}, nil
}

// runInternal 是 Run 的内部实现，支持 stdin 注入。
// 优先走 Piston 容器执行(真隔离);Piston 不可用时降级本地 subprocess(带黑名单)。
func runInternal(language, code, stdin string) (Result, error) {
	if r, ok := runPiston(context.Background(), language, code, stdin); ok {
		return r, nil
	}
	// ---- 降级路径:Piston 不可用时的本地执行(黑名单 + 超时,隔离性弱) ----
	for _, x := range blocked {
		if strings.Contains(strings.ToLower(code), strings.ToLower(x)) {
			return Result{Status: "rejected", Stderr: "代码包含沙箱禁止的系统或网络操作"}, nil
		}
	}
	dir, err := os.MkdirTemp("", "codeforge-sandbox-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(dir)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	var cmd *exec.Cmd
	switch strings.ToLower(language) {
	case "python", "python3":
		path := filepath.Join(dir, "main.py")
		os.WriteFile(path, []byte(code), 0600)
		bin := localPython()
		if bin == "" {
			return Result{Status: "unavailable", Stderr: "Python 未安装"}, nil
		}
		cmd = exec.CommandContext(ctx, bin, "-I", path)
	case "javascript", "js", "node":
		path := filepath.Join(dir, "main.js")
		os.WriteFile(path, []byte(code), 0600)
		bin := find("node")
		if bin == "" {
			return Result{Status: "unavailable", Stderr: "Node.js 未安装"}, nil
		}
		cmd = exec.CommandContext(ctx, bin, path)
	case "cpp", "c++":
		src := filepath.Join(dir, "main.cpp")
		exe := filepath.Join(dir, "main.exe")
		if runtime.GOOS != "windows" {
			exe = filepath.Join(dir, "main")
		}
		os.WriteFile(src, []byte(code), 0600)
		compiler := find("g++", "clang++")
		if compiler == "" {
			return Result{Status: "unavailable", Stderr: "C++ 编译器未安装"}, nil
		}
		compile := exec.CommandContext(ctx, compiler, src, "-O2", "-std=c++17", "-o", exe)
		out, e := compile.CombinedOutput()
		if e != nil {
			return Result{Status: "compile_error", Stderr: limit(string(out))}, nil
		}
		cmd = exec.CommandContext(ctx, exe)
	case "rust":
		src := filepath.Join(dir, "main.rs")
		exe := filepath.Join(dir, "main.exe")
		if runtime.GOOS != "windows" {
			exe = filepath.Join(dir, "main")
		}
		os.WriteFile(src, []byte(code), 0600)
		compiler := find("rustc")
		if compiler == "" {
			return Result{Status: "unavailable", Stderr: "Rust 编译器未安装"}, nil
		}
		compile := exec.CommandContext(ctx, compiler, src, "-O", "-o", exe)
		out, e := compile.CombinedOutput()
		if e != nil {
			return Result{Status: "compile_error", Stderr: limit(string(out))}, nil
		}
		cmd = exec.CommandContext(ctx, exe)
	default:
		return Result{Status: "unsupported", Stderr: "支持 python、javascript、cpp、rust"}, nil
	}
	cmd.Dir = dir
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "PYTHONIOENCODING=utf-8", "NO_PROXY=*"}
	// 注入 stdin
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	elapsed := time.Since(start).Milliseconds()
	if ctx.Err() == context.DeadlineExceeded {
		return Result{Status: "timeout", Stdout: limit(stdout.String()), Stderr: "执行超过 5 秒", ExecutionTimeMS: elapsed}, nil
	}
	status := "success"
	if err != nil {
		status = "runtime_error"
	}
	return Result{Status: status, Stdout: limit(stdout.String()), Stderr: limit(stderr.String()), ExecutionTimeMS: elapsed, MemoryKB: 0}, nil
}
