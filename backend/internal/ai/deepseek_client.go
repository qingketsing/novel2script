package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrDeepSeekPromptRequired = errors.New("deepseek prompt is required")
	ErrDeepSeekEmptyResponse  = errors.New("deepseek response content is empty")
)

type DeepSeekClientConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DeepSeekClient struct {
	cfg        DeepSeekClientConfig
	httpClient HTTPDoer
}

func NewDeepSeekClient(cfg DeepSeekClientConfig, httpClient HTTPDoer) (*DeepSeekClient, error) {
	// 在构造阶段拦截缺失配置，避免运行到请求阶段才暴露模糊错误。
	if cfg.APIKey == "" {
		return nil, ErrDeepSeekAPIKeyRequired
	}
	if cfg.BaseURL == "" {
		return nil, ErrDeepSeekBaseURLRequired
	}
	if cfg.Model == "" {
		return nil, ErrDeepSeekModelRequired
	}
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	return &DeepSeekClient{
		cfg:        cfg,
		httpClient: httpClient,
	}, nil
}

func (c *DeepSeekClient) GenerateYAML(ctx context.Context, prompt string) (string, error) {
	// prompt 是真实模型调用的核心输入，空 prompt 直接视为调用方错误。
	if strings.TrimSpace(prompt) == "" {
		return "", ErrDeepSeekPromptRequired
	}

	// DeepSeek 使用 OpenAI-compatible chat completions 协议，这里只发送最小消息结构。
	body, err := json.Marshal(deepSeekChatRequest{
		Model: c.cfg.Model,
		Messages: []deepSeekChatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// client 通过接口注入，单元测试可以使用 fake doer，避免真实访问网络。
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("deepseek api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	// 第一版只取首个 choice 的 message.content，后续 provider 会负责校验其 YAML 结构。
	var decoded deepSeekChatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 {
		return "", ErrDeepSeekEmptyResponse
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return "", ErrDeepSeekEmptyResponse
	}
	return content, nil
}

type deepSeekChatRequest struct {
	Model       string                `json:"model"`
	Messages    []deepSeekChatMessage `json:"messages"`
	Temperature float64               `json:"temperature"`
}

type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
