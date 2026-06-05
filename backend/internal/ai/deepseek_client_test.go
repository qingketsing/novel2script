package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDeepSeekClientRequiresConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  DeepSeekClientConfig
		want error
	}{
		{
			name: "api key",
			cfg: DeepSeekClientConfig{
				BaseURL: "https://api.deepseek.com",
				Model:   "deepseek-v4",
			},
			want: ErrDeepSeekAPIKeyRequired,
		},
		{
			name: "base url",
			cfg: DeepSeekClientConfig{
				APIKey: "test-api-key",
				Model:  "deepseek-v4",
			},
			want: ErrDeepSeekBaseURLRequired,
		},
		{
			name: "model",
			cfg: DeepSeekClientConfig{
				APIKey:  "test-api-key",
				BaseURL: "https://api.deepseek.com",
			},
			want: ErrDeepSeekModelRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDeepSeekClient(tt.cfg, http.DefaultClient)

			if err != tt.want {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDeepSeekClientGenerateYAMLPostsChatCompletion(t *testing.T) {
	var gotAuth string
	var gotRequest struct {
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Fatalf("content-type = %q, want application/json", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"schema_version: \"1.0\"\nmetadata:\n  title: demo"}}]}`))
	}))
	defer server.Close()

	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4",
	})

	yamlText, err := client.GenerateYAML(context.Background(), "convert this novel")

	if err != nil {
		t.Fatalf("GenerateYAML() error = %v", err)
	}
	if yamlText != "schema_version: \"1.0\"\nmetadata:\n  title: demo" {
		t.Fatalf("yaml = %q", yamlText)
	}
	if gotAuth != "Bearer test-api-key" {
		t.Fatalf("authorization = %q, want bearer token", gotAuth)
	}
	if gotRequest.Model != "deepseek-v4" {
		t.Fatalf("model = %q, want deepseek-v4", gotRequest.Model)
	}
	if gotRequest.Temperature != 0.2 {
		t.Fatalf("temperature = %v, want 0.2", gotRequest.Temperature)
	}
	if len(gotRequest.Messages) != 1 {
		t.Fatalf("messages length = %d, want 1", len(gotRequest.Messages))
	}
	if gotRequest.Messages[0].Role != "user" {
		t.Fatalf("message role = %q, want user", gotRequest.Messages[0].Role)
	}
	if gotRequest.Messages[0].Content != "convert this novel" {
		t.Fatalf("message content = %q", gotRequest.Messages[0].Content)
	}
}

func TestDeepSeekClientGenerateYAMLRejectsEmptyPrompt(t *testing.T) {
	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})

	_, err := client.GenerateYAML(context.Background(), "  ")

	if !errors.Is(err, ErrDeepSeekPromptRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekPromptRequired)
	}
}

func TestDeepSeekClientGenerateYAMLReturnsHTTPClientError(t *testing.T) {
	wantErr := errors.New("network unavailable")
	client := mustNewDeepSeekClientWithDoer(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	}, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, wantErr
	}))

	_, err := client.GenerateYAML(context.Background(), "prompt")

	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestDeepSeekClientGenerateYAMLRejectsNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()
	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4",
	})

	_, err := client.GenerateYAML(context.Background(), "prompt")

	if err == nil || !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("error = %v, want status 400 error", err)
	}
}

func TestDeepSeekClientGenerateYAMLRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()
	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4",
	})

	_, err := client.GenerateYAML(context.Background(), "prompt")

	if err == nil {
		t.Fatal("error is nil, want invalid JSON error")
	}
}

func TestDeepSeekClientGenerateYAMLRejectsEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()
	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4",
	})

	_, err := client.GenerateYAML(context.Background(), "prompt")

	if !errors.Is(err, ErrDeepSeekEmptyResponse) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekEmptyResponse)
	}
}

func TestDeepSeekClientGenerateYAMLRejectsEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"  "}}]}`))
	}))
	defer server.Close()
	client := mustNewDeepSeekClient(t, DeepSeekClientConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4",
	})

	_, err := client.GenerateYAML(context.Background(), "prompt")

	if !errors.Is(err, ErrDeepSeekEmptyResponse) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekEmptyResponse)
	}
}

func mustNewDeepSeekClient(t *testing.T, cfg DeepSeekClientConfig) *DeepSeekClient {
	t.Helper()

	client, err := NewDeepSeekClient(cfg, http.DefaultClient)
	if err != nil {
		t.Fatalf("NewDeepSeekClient() error = %v", err)
	}
	return client
}

func mustNewDeepSeekClientWithDoer(t *testing.T, cfg DeepSeekClientConfig, doer HTTPDoer) *DeepSeekClient {
	t.Helper()

	client, err := NewDeepSeekClient(cfg, doer)
	if err != nil {
		t.Fatalf("NewDeepSeekClient() error = %v", err)
	}
	return client
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
