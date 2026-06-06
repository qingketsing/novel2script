package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
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

func TestDeepSeekProviderGenerateScreenplayReturnsRawYAML(t *testing.T) {
	generator := &recordingYAMLGenerator{responses: []yamlGeneratorResponse{{yamlText: validScreenplayYAML}}}
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
	if len(generator.prompts) != 1 {
		t.Fatalf("expected one client call, got %d", len(generator.prompts))
	}
	if generator.prompts[0] == "" {
		t.Fatal("expected provider to build prompt before calling client")
	}
	if !strings.Contains(generator.prompts[0], "第一章 雨夜来信") {
		t.Fatalf("expected prompt to include chapter title:\n%s", generator.prompts[0])
	}
	if !strings.Contains(generator.prompts[0], "林舟在雨夜收到一封没有署名的信。") {
		t.Fatalf("expected prompt to include chapter content:\n%s", generator.prompts[0])
	}
}

func TestDeepSeekProviderGenerateScreenplayRepairsInvalidYAMLOnce(t *testing.T) {
	const invalidYAML = "schema_version: \"1.0\""
	generator := &recordingYAMLGenerator{responses: []yamlGeneratorResponse{
		{yamlText: invalidYAML},
		{yamlText: validScreenplayYAML},
	}}
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
		t.Fatalf("expected repaired YAML, got:\n%s", output.RawYAML)
	}
	if len(generator.prompts) != 2 {
		t.Fatalf("expected initial call and one repair call, got %d", len(generator.prompts))
	}

	repairPrompt := generator.prompts[1]
	required := []string{
		invalidYAML,
		"metadata.title",
		"metadata.title 不能为空",
		"只输出 YAML",
		"不要输出 Markdown 代码块",
	}
	for _, want := range required {
		if !strings.Contains(repairPrompt, want) {
			t.Fatalf("expected repair prompt to contain %q:\n%s", want, repairPrompt)
		}
	}
}

func TestDeepSeekProviderGenerateScreenplayRejectsInvalidYAML(t *testing.T) {
	provider := DeepSeekProvider{
		cfg: validDeepSeekConfig(),
		yamlGenerator: &recordingYAMLGenerator{responses: []yamlGeneratorResponse{
			{yamlText: "schema_version: \"1.0\""},
			{yamlText: "schema_version: \"1.0\""},
		}},
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

func TestDeepSeekProviderGenerateScreenplayReturnsRepairClientError(t *testing.T) {
	wantErr := errors.New("repair client failed")
	provider := DeepSeekProvider{
		cfg: validDeepSeekConfig(),
		yamlGenerator: &recordingYAMLGenerator{responses: []yamlGeneratorResponse{
			{yamlText: "schema_version: \"1.0\""},
			{err: wantErr},
		}},
	}

	_, err := provider.GenerateScreenplay(context.Background(), GenerateInput{
		Novel: samplePromptNovel(),
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestDeepSeekProviderGenerateScreenplayReturnsClientError(t *testing.T) {
	wantErr := errors.New("client failed")
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: &recordingYAMLGenerator{responses: []yamlGeneratorResponse{{err: wantErr}}},
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
	prompts   []string
	responses []yamlGeneratorResponse
}

type yamlGeneratorResponse struct {
	yamlText string
	err      error
}

func (g *recordingYAMLGenerator) GenerateYAML(_ context.Context, prompt string) (string, error) {
	g.prompts = append(g.prompts, prompt)
	if len(g.responses) == 0 {
		return "", errors.New("missing yaml generator response")
	}

	response := g.responses[0]
	g.responses = g.responses[1:]
	if response.err != nil {
		return "", response.err
	}
	return response.yamlText, nil
}
