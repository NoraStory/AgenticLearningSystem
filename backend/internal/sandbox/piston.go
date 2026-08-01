package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ---- Piston 代码执行引擎客户端(MIT 许可,自建实例,容器化隔离) ----
// 设计: 优先调 Piston 的 /execute 接口在隔离容器里执行代码;
// Piston 不可用时降级到本地 subprocess 实现(见 runner.go),保证功能可用。

// Piston 语言别名 → 引擎运行时名(与 runner.go 支持的语言对齐)。
var pistonLangAlias = map[string]string{
	"python": "python", "python3": "python",
	"javascript": "javascript", "js": "javascript", "node": "javascript",
	"cpp": "c++", "c++": "c++",
	"rust": "rust",
}

// pistonRuntime 是 /runtimes 返回的单个运行时。
type pistonRuntime struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Aliases  []string `json:"aliases"`
}

// pistonRequest 对应 /execute 请求体。
type pistonRequest struct {
	Language string   `json:"language"`
	Version  string   `json:"version"`
	Files    []pistonFile `json:"files"`
	Stdin    string   `json:"stdin,omitempty"`
}

type pistonFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// pistonResponse 对应 /execute 响应体。
type pistonResponse struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Run      struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Code   int    `json:"code"`
		Signal *string `json:"signal"`
	} `json:"run"`
}

// pistonClient 是 Piston HTTP API 客户端。
type pistonClient struct {
	url    string
	http   *http.Client
	mu     sync.Mutex
	// 语言 → 最新版本,首次调用时从 /runtimes 拉取缓存。
	runtimes map[string]string
	// 健康状态缓存,避免每次执行都探测。
	healthyState *bool
	healthyAt    time.Time
}

var pistonSingleton *pistonClient
var pistonOnce sync.Once

// pistonClientInstance 返回全局 Piston 客户端(惰性创建,复用 HTTP 连接)。
// 每次调用重新读环境变量,便于测试注入。
func pistonClientInstance() *pistonClient {
	pistonOnce.Do(func() {
		url := os.Getenv("PISTON_URL")
		if url == "" {
			url = "http://localhost:2000"
		}
		pistonSingleton = &pistonClient{
			url:  strings.TrimRight(url, "/"),
			http: &http.Client{Timeout: 15 * time.Second},
		}
	})
	return pistonSingleton
}

// healthy 探测 Piston 服务是否可用(10 秒缓存)。
// 需要 /api/v2/runtimes 返回非空运行时列表才算可用——空列表说明
// 运行时未安装,execute 会失败,此时应降级本地沙箱。
func (p *pistonClient) healthy() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.healthyState != nil && time.Since(p.healthyAt) < 10*time.Second {
		return *p.healthyState
	}
	ok := false
	if p.url != "" {
		res, err := p.http.Get(p.url + "/api/v2/runtimes")
		if err == nil {
			defer res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
				ok = len(raw) > 2 && !strings.Contains(string(raw), `[]`) // 非空数组才算可用
			}
		}
	}
	p.healthyState = &ok
	p.healthyAt = time.Now()
	return ok
}

// loadRuntimes 拉取可用运行时并缓存版本映射。
func (p *pistonClient) loadRuntimes() map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.runtimes != nil {
		return p.runtimes
	}
	res, err := p.http.Get(p.url + "/api/v2/runtimes")
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil
	}
	var runtimes []pistonRuntime
	if json.NewDecoder(res.Body).Decode(&runtimes) != nil {
		return nil
	}
	m := map[string]string{}
	for _, r := range runtimes {
		m[r.Language] = r.Version
	}
	p.runtimes = m
	return m
}

// execute 在 Piston 上执行代码,返回与 runner.go Result 一致的结构。
func (p *pistonClient) execute(ctx context.Context, language, code, stdin string) (Result, error) {
	engine := pistonLangAlias[strings.ToLower(language)]
	if engine == "" {
		return Result{Status: "unsupported", Stderr: "Piston 不支持该语言"}, nil
	}
	version := ""
	if runtimes := p.loadRuntimes(); runtimes != nil {
		version = runtimes[engine]
	}
	body, _ := json.Marshal(pistonRequest{
		Language: engine,
		Version:  version,
		Files:    []pistonFile{{Name: "main", Content: code}},
		Stdin:    stdin,
	})
	res, err := p.http.Post(p.url+"/api/v2/execute", "application/json", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 {
		return Result{Status: "runtime_error", Stderr: fmt.Sprintf("Piston HTTP %d: %s", res.StatusCode, limit(string(raw)))}, nil
	}
	var out pistonResponse
	if json.Unmarshal(raw, &out) != nil {
		return Result{Status: "runtime_error", Stderr: "Piston 响应解析失败"}, nil
	}
	r := Result{
		Status:   "success",
		Stdout:   limit(out.Run.Stdout),
		Stderr:   limit(out.Run.Stderr),
		MemoryKB: 0, // Piston 不返回内存统计
	}
	if out.Run.Code != 0 {
		r.Status = "runtime_error"
	}
	return r, nil
}

// runPiston 尝试在 Piston 上执行;服务不可用时返回 (zero, false) 由调用方降级。
func runPiston(ctx context.Context, language, code, stdin string) (Result, bool) {
	p := pistonClientInstance()
	if !p.healthy() {
		return Result{}, false
	}
	r, err := p.execute(ctx, language, code, stdin)
	if err != nil {
		return Result{}, false // 网络失败视为不可用,降级本地
	}
	return r, true
}
