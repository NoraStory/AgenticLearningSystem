package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codeforge/backend/internal/config"
)

type Client struct {
	apiKey, baseURL, model, wireAPI string
	http                            *http.Client
}
type apiRequest struct {
	Model           string  `json:"model"`
	Input           any     `json:"input,omitempty"`
	Messages        any     `json:"messages,omitempty"`
	Stream          bool    `json:"stream,omitempty"`
	MaxOutputTokens int     `json:"max_output_tokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}
type responsesResponse struct {
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func New(cfg config.Config) *Client {
	return &Client{apiKey: cfg.ArkAPIKey, baseURL: strings.TrimRight(cfg.ArkBaseURL, "/"), model: cfg.ArkModel, wireAPI: cfg.LLMWireAPI, http: &http.Client{Timeout: 90 * time.Second}}
}
func (c *Client) Enabled() bool { return c.apiKey != "" && c.model != "" }
func (c *Client) Chat(ctx context.Context, system, user string) (string, error) {
	return c.ChatWithImages(ctx, system, user, nil)
}
func (c *Client) ChatWithImages(ctx context.Context, system, user string, images []string) (string, error) {
	var answer strings.Builder
	err := c.StreamChatWithImages(ctx, system, user, images, func(delta string) { answer.WriteString(delta) })
	return answer.String(), err
}

func (c *Client) StreamChatWithImages(ctx context.Context, system, user string, images []string, onToken func(string)) error {
	if !c.Enabled() {
		onToken(fallback(user))
		return nil
	}
	if strings.EqualFold(c.wireAPI, "responses") {
		return c.streamResponses(ctx, system, user, images, onToken)
	}
	return c.streamChatCompletions(ctx, system, user, images, onToken)
}

func (c *Client) streamResponses(ctx context.Context, system, user string, images []string, onToken func(string)) error {
	var userContent any = user
	if len(images) > 0 {
		content := []map[string]any{{"type": "input_text", "text": user}}
		for _, img := range images {
			content = append(content, map[string]any{"type": "input_image", "image_url": img})
		}
		userContent = content
	}
	input := []map[string]any{{"role": "system", "content": system}, {"role": "user", "content": userContent}}
	maxTokens := 4096
	if len(images) > 0 {
		maxTokens = 2048
	}
	body, _ := json.Marshal(apiRequest{Model: c.model, Input: input, Stream: true, MaxOutputTokens: maxTokens})
	res, err := c.request(ctx, c.baseURL+"/responses", body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 4<<20)
	received := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var event struct {
			Type     string             `json:"type"`
			Delta    string             `json:"delta"`
			Response *responsesResponse `json:"response"`
		}
		if json.Unmarshal([]byte(payload), &event) != nil {
			continue
		}
		if event.Delta != "" {
			received = true
			onToken(event.Delta)
		} else if !received && event.Response != nil {
			for _, item := range event.Response.Output {
				for _, part := range item.Content {
					if part.Text != "" {
						received = true
						onToken(part.Text)
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !received {
		return fmt.Errorf("responses stream returned empty output")
	}
	return nil
}

func (c *Client) streamChatCompletions(ctx context.Context, system, user string, images []string, onToken func(string)) error {
	var userContent any = user
	if len(images) > 0 {
		content := []map[string]any{{"type": "text", "text": user}}
		for _, img := range images {
			content = append(content, map[string]any{"type": "image_url", "image_url": map[string]string{"url": img}})
		}
		userContent = content
	}
	messages := []map[string]any{{"role": "system", "content": system}, {"role": "user", "content": userContent}}
	body, _ := json.Marshal(apiRequest{Model: c.model, Messages: messages, Stream: true, MaxOutputTokens: 4096, Temperature: .7})
	res, err := c.request(ctx, c.baseURL+"/chat/completions", body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 4<<20)
	received := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(payload), &event) == nil && len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			received = true
			onToken(event.Choices[0].Delta.Content)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !received {
		return fmt.Errorf("chat stream returned empty output")
	}
	return nil
}

func (c *Client) request(ctx context.Context, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
		res.Body.Close()
		return nil, fmt.Errorf("llm api %s: %s", res.Status, string(raw))
	}
	return res, nil
}
func fallback(q string) string {
	return "当前模型服务不可用，无法生成 AI 回答。\n\n你的问题：" + q + "\n\n建议检查 .env 中的 ARK_API_KEY 和 ARK_MODEL 配置，确保模型服务正常运行后重试。"
}
