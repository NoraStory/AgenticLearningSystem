package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ---- docx 模板工具:以 Python 子进程(backend/.venv)调用 scripts/ 下的脚本 ----

// pythonExe 返回 venv 中的 python 可执行文件路径;找不到时返回 "python" 兜底。
func (s *Server) pythonExe() string {
	root := s.backendRoot()
	candidates := []string{
		filepath.Join(root, ".venv", "Scripts", "python.exe"),
		filepath.Join(root, ".venv", "bin", "python3"),
		filepath.Join(root, ".venv", "bin", "python"),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return "python"
}

// backendRoot 从当前工作目录上溯查找 backend 目录(含 go.mod 的目录)。
func (s *Server) backendRoot() string {
	if d, err := os.Getwd(); err == nil {
		for cur := d; cur != "" && cur != filepath.Dir(cur); cur = filepath.Dir(cur) {
			if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
				return cur
			}
		}
	}
	return "."
}

// runScript 以 stdin 传入 JSON、运行指定脚本,返回 stdout 字节。
func (s *Server) runScript(ctx context.Context, script string, args []string, input any) ([]byte, error) {
	exe := s.pythonExe()
	scriptPath := filepath.Join(s.backendRoot(), "scripts", script)
	fullArgs := append([]string{scriptPath}, args...)
	var in bytes.Buffer
	if input != nil {
		raw, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		in.Write(raw)
	}
	cmd := exec.CommandContext(ctx, exe, fullArgs...)
	cmd.Stdin = &in
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("python %s: %w: %s", script, err, strings.TrimSpace(errBuf.String()))
	}
	return out.Bytes(), nil
}

// parseDocxTemplate 解析 docx 结构,返回段落与表格。
func (s *Server) parseDocxTemplate(ctx context.Context, path string) (map[string]any, error) {
	out, err := s.runScript(ctx, "tpl_parse.py", []string{path}, nil)
	if err != nil {
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("tpl_parse json: %w", err)
	}
	return parsed, nil
}

// registerDocxTemplate 按章节结构注入占位符,输出注册模板到 outPath。
func (s *Server) registerDocxTemplate(ctx context.Context, srcPath, outPath string, sections []map[string]any) error {
	_, err := s.runScript(ctx, "tpl_register.py", []string{srcPath, outPath}, map[string]any{"sections": sections})
	return err
}

// renderDocxTemplate 用 docxtpl 渲染注册模板,返回渲染后 docx 字节。
func (s *Server) renderDocxTemplate(ctx context.Context, templatePath string, sections []map[string]any) ([]byte, error) {
	return s.runScript(ctx, "tpl_render.py", []string{templatePath}, map[string]any{"sections": sections})
}

// sofficePath 返回 LibreOffice 可执行文件路径(Linux/macOS 为 soffice,Windows 常见安装路径)。
// 保留作为本地 LibreOffice 的兜底;首选 Gotenberg(docker 服务)转换。
func sofficePath() string {
	if p := os.Getenv("SOFFICE_PATH"); p != "" {
		return p
	}
	for _, p := range []string{
		`C:\Program Files\LibreOffice\program\soffice.exe`,
		`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if _, err := exec.LookPath("soffice"); err == nil {
		return "soffice"
	}
	return ""
}

// gotenbergClient 是 docx→PDF 转换客户端(封装 Gotenberg HTTP API,MIT 许可)。
type gotenbergClient struct {
	url    string
	http   *http.Client
	soffice string // 本地 LibreOffice 兜底路径(为空表示无)
}

func newGotenbergClient(url string) *gotenbergClient {
	return &gotenbergClient{
		url:     strings.TrimRight(url, "/"),
		http:    &http.Client{Timeout: 60 * time.Second},
		soffice: sofficePath(),
	}
}

// convertDocxToPDF 优先走 Gotenberg HTTP API,失败时回退本地 soffice,两者都不可用则报错。
func (g *gotenbergClient) convertDocxToPDF(ctx context.Context, docxBytes []byte) ([]byte, error) {
	if g.url != "" {
		data, err := g.convertViaGotenberg(ctx, docxBytes)
		if err == nil {
			return data, nil
		}
	}
	if g.soffice != "" {
		return docxToPDFViaSoffice(ctx, g.soffice, docxBytes)
	}
	return nil, fmt.Errorf("no pdf converter available (gotenberg unreachable and libreoffice not installed)")
}

// convertViaGotenberg 用 Gotenberg /forms/libreoffice/convert 接口转换。
func (g *gotenbergClient) convertViaGotenberg(ctx context.Context, docxBytes []byte) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", "resume.docx")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(docxBytes); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.url+"/forms/libreoffice/convert", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := g.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
		return nil, fmt.Errorf("gotenberg: http %d: %s", res.StatusCode, string(raw))
	}
	return io.ReadAll(res.Body)
}

// docxToPDFViaSoffice 用本地 LibreOffice headless 把 docx 转成 PDF。
func docxToPDFViaSoffice(ctx context.Context, soffice string, docxBytes []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "codeforge-docx-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	src := filepath.Join(tmpDir, "resume.docx")
	if err := os.WriteFile(src, docxBytes, 0o644); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, soffice, "--headless", "--convert-to", "pdf", "--outdir", tmpDir, src)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("soffice: %w: %s", err, strings.TrimSpace(errBuf.String()))
	}
	out := filepath.Join(tmpDir, "resume.pdf")
	data, err := os.ReadFile(out)
	if err != nil {
		return nil, fmt.Errorf("soffice pdf output: %w", err)
	}
	return data, nil
}

// gotenbergHealthy 探测 Gotenberg 服务是否可用(用于 health 接口)。
func (s *Server) gotenbergHealthy() bool {
	url := s.cfg.GotenbergURL
	if url == "" {
		return false
	}
	res, err := http.Get(strings.TrimRight(url, "/") + "/health")
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK
}

// docxConverter 返回 Server 持有的 docx→PDF 转换器。
func (s *Server) docxConverter() *gotenbergClient {
	return newGotenbergClient(s.cfg.GotenbergURL)
}
