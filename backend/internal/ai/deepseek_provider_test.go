package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/qingketsing/novel2script/backend/internal/domain"
	"github.com/qingketsing/novel2script/backend/internal/observability"
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

func TestNewDeepSeekProviderDoesNotLetClientTimeoutUndercutDynamicTimeout(t *testing.T) {
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
	if httpClient.Timeout != deepSeekRequestTimeoutMax {
		t.Fatalf("timeout = %v, want %v", httpClient.Timeout, deepSeekRequestTimeoutMax)
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

func TestDeepSeekProviderLogsGenerationStages(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	ctx := observability.WithRequestID(observability.WithLogger(context.Background(), logger), "req_ai")
	generator := &recordingYAMLGenerator{responses: []yamlGeneratorResponse{{yamlText: validScreenplayYAML}}}
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: generator,
	}

	output, err := provider.GenerateScreenplay(ctx, GenerateInput{
		Novel: samplePromptNovel(),
	})
	if err != nil {
		t.Fatalf("GenerateScreenplay returned error: %v", err)
	}

	logs := decodeAIJSONLogs(t, logBuffer.String())
	started := findAILogMessage(t, logs, "deepseek generation started")
	if started["request_id"] != "req_ai" {
		t.Fatalf("request_id = %v, want req_ai", started["request_id"])
	}
	if started["chapter_count"] != float64(3) {
		t.Fatalf("chapter_count = %v, want 3", started["chapter_count"])
	}
	if _, ok := started["timeout_ms"]; !ok {
		t.Fatalf("expected timeout_ms in started log: %+v", started)
	}
	if _, ok := started["prompt_length"]; !ok {
		t.Fatalf("expected prompt_length in started log: %+v", started)
	}

	returned := findAILogMessage(t, logs, "deepseek generation returned")
	if returned["yaml_length"] != float64(len(output.RawYAML)) {
		t.Fatalf("yaml_length = %v, want %d", returned["yaml_length"], len(output.RawYAML))
	}
	findAILogMessage(t, logs, "deepseek yaml validation succeeded")
}

func TestDeepSeekProviderUsesMinimumTimeoutForShortNovel(t *testing.T) {
	provider := DeepSeekProvider{cfg: validDeepSeekConfig()}

	timeout := provider.timeoutForInput(GenerateInput{Novel: domain.Novel{Title: "短"}})

	if timeout != 45*time.Second {
		t.Fatalf("timeout = %v, want %v", timeout, 45*time.Second)
	}
}

func TestDeepSeekProviderTimeoutScalesWithNovelLength(t *testing.T) {
	provider := DeepSeekProvider{cfg: validDeepSeekConfig()}

	shortTimeout := provider.timeoutForInput(GenerateInput{Novel: samplePromptNovel()})
	longTimeout := provider.timeoutForInput(GenerateInput{Novel: longPromptNovel(6, 3200)})

	if longTimeout <= shortTimeout {
		t.Fatalf("long timeout = %v, want greater than short timeout %v", longTimeout, shortTimeout)
	}
	if longTimeout >= deepSeekRequestTimeoutMax {
		t.Fatalf("long timeout = %v, want below max %v", longTimeout, deepSeekRequestTimeoutMax)
	}
}

func TestDeepSeekProviderTimeoutCapsHugeNovel(t *testing.T) {
	provider := DeepSeekProvider{cfg: validDeepSeekConfig()}

	timeout := provider.timeoutForInput(GenerateInput{Novel: longPromptNovel(20, 30000)})

	if timeout != deepSeekRequestTimeoutMax {
		t.Fatalf("timeout = %v, want %v", timeout, deepSeekRequestTimeoutMax)
	}
}

func TestDeepSeekProviderGenerateScreenplayPassesDynamicDeadline(t *testing.T) {
	generator := &recordingYAMLGenerator{responses: []yamlGeneratorResponse{{yamlText: validScreenplayYAML}}}
	provider := DeepSeekProvider{
		cfg:           validDeepSeekConfig(),
		yamlGenerator: generator,
	}
	input := GenerateInput{Novel: longPromptNovel(6, 3200)}

	_, err := provider.GenerateScreenplay(context.Background(), input)
	if err != nil {
		t.Fatalf("GenerateScreenplay returned error: %v", err)
	}
	if len(generator.deadlineDurations) != 1 {
		t.Fatalf("deadline count = %d, want 1", len(generator.deadlineDurations))
	}

	want := provider.timeoutForInput(input)
	got := generator.deadlineDurations[0]
	if got < want-time.Second || got > want {
		t.Fatalf("deadline duration = %v, want near %v", got, want)
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
	prompts           []string
	deadlineDurations []time.Duration
	responses         []yamlGeneratorResponse
}

type yamlGeneratorResponse struct {
	yamlText string
	err      error
}

func (g *recordingYAMLGenerator) GenerateYAML(ctx context.Context, prompt string) (string, error) {
	g.prompts = append(g.prompts, prompt)
	if deadline, ok := ctx.Deadline(); ok {
		g.deadlineDurations = append(g.deadlineDurations, time.Until(deadline))
	}
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

func decodeAIJSONLogs(t *testing.T, raw string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	logs := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		logs = append(logs, entry)
	}
	return logs
}

func findAILogMessage(t *testing.T, logs []map[string]any, message string) map[string]any {
	t.Helper()

	for _, entry := range logs {
		if entry["msg"] == message {
			return entry
		}
	}
	t.Fatalf("missing log message %q in %+v", message, logs)
	return nil
}

func longPromptNovel(chapterCount int, charsPerChapter int) domain.Novel {
	const paragraph = "林川站在雨夜的公交车厢里，反复确认手中的车票和窗外的隧道编号。"
	repeatCount := charsPerChapter/len([]rune(paragraph)) + 1
	content := strings.Repeat(paragraph, repeatCount)
	chapters := make([]domain.Chapter, 0, chapterCount)
	for i := 1; i <= chapterCount; i++ {
		chapters = append(chapters, domain.Chapter{
			ID:      fmt.Sprintf("chapter_%03d", i),
			Title:   fmt.Sprintf("第%d章 长夜测试", i),
			Order:   i,
			Content: content,
			Summary: "林川继续追查无终点车票背后的秘密。",
		})
	}
	return domain.Novel{
		Title:    "没有终点的车票",
		Content:  content,
		Chapters: chapters,
	}
}
