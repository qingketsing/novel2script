package ai

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewDeepSeekProviderRequiresAPIKey(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})

	if !errors.Is(err, ErrDeepSeekAPIKeyRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekAPIKeyRequired)
	}
}

func TestNewDeepSeekProviderRequiresBaseURL(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey: "test-api-key",
		Model:  "deepseek-v4",
	})

	if !errors.Is(err, ErrDeepSeekBaseURLRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekBaseURLRequired)
	}
}

func TestNewDeepSeekProviderRequiresModel(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
	})

	if !errors.Is(err, ErrDeepSeekModelRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekModelRequired)
	}
}

func TestNewDeepSeekProviderReturnsProvider(t *testing.T) {
	provider, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})

	if err != nil {
		t.Fatalf("NewDeepSeekProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider is nil, want DeepSeek provider")
	}
}

func TestNewDeepSeekProviderPassesTimeoutToClient(t *testing.T) {
	provider, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewDeepSeekProvider() error = %v", err)
	}

	deepSeekProvider, ok := provider.(DeepSeekProvider)
	if !ok {
		t.Fatalf("provider = %T, want DeepSeekProvider", provider)
	}
	client, ok := deepSeekProvider.yamlGenerator.(*DeepSeekClient)
	if !ok {
		t.Fatalf("yamlGenerator = %T, want *DeepSeekClient", deepSeekProvider.yamlGenerator)
	}
	httpClient, ok := client.httpClient.(*http.Client)
	if !ok {
		t.Fatalf("httpClient = %T, want *http.Client", client.httpClient)
	}
	if httpClient.Timeout != 5*time.Second {
		t.Fatalf("timeout = %v, want %v", httpClient.Timeout, 5*time.Second)
	}
}

func TestDeepSeekProviderGenerateScreenplayReturnsRawYAML(t *testing.T) {
	generator := &recordingYAMLGenerator{yamlText: validScreenplayYAML}
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: generator,
	}

	output, err := provider.GenerateScreenplay(context.Background(), GenerateInput{
		Novel: samplePromptNovel(),
	})
	if err != nil {
		t.Fatalf("GenerateScreenplay returned error: %v", err)
	}

	if output.RawYAML != validScreenplayYAML {
		t.Fatalf("unexpected raw yaml:\n%s", output.RawYAML)
	}
	if generator.prompt == "" {
		t.Fatal("expected provider to build prompt before calling client")
	}
	if !strings.Contains(generator.prompt, "第一章 雨夜来信") {
		t.Fatalf("expected prompt to include chapter title:\n%s", generator.prompt)
	}
	if !strings.Contains(generator.prompt, "林舟在雨夜收到一封没有署名的信。") {
		t.Fatalf("expected prompt to include chapter content:\n%s", generator.prompt)
	}
}

func TestDeepSeekProviderGenerateScreenplayRejectsInvalidYAML(t *testing.T) {
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: &recordingYAMLGenerator{yamlText: "schema_version: \"1.0\""},
	}

	_, err := provider.GenerateScreenplay(context.Background(), GenerateInput{
		Novel: samplePromptNovel(),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr YAMLValidationError
	if !AsYAMLValidationError(err, &validationErr) {
		t.Fatalf("expected YAMLValidationError, got %T: %v", err, err)
	}
}

func TestDeepSeekProviderGenerateScreenplayReturnsClientError(t *testing.T) {
	wantErr := errors.New("client failed")
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: &recordingYAMLGenerator{err: wantErr},
	}

	_, err := provider.GenerateScreenplay(context.Background(), GenerateInput{
		Novel: samplePromptNovel(),
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func validDeepSeekConfig() DeepSeekConfig {
	return DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	}
}

type recordingYAMLGenerator struct {
	prompt   string
	yamlText string
	err      error
}

func (g *recordingYAMLGenerator) GenerateYAML(_ context.Context, prompt string) (string, error) {
	g.prompt = prompt
	if g.err != nil {
		return "", g.err
	}
	return g.yamlText, nil
}
