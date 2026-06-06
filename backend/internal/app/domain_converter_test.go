package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/domain"
	"github.com/qingketsing/novel2script/backend/internal/observability"
)

func TestMockDomainConverterConvertsSuccessfulInput(t *testing.T) {
	converter := NewMockDomainConverter()

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:     "雨夜来信",
		Content:   sampleConvertNovel,
		InputType: "text",
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if resp.ChapterCount != 3 {
		t.Fatalf("expected 3 chapters, got %d", resp.ChapterCount)
	}
	if resp.Mode != "mock" {
		t.Fatalf("expected mock mode, got %q", resp.Mode)
	}
	if !strings.Contains(resp.ScreenplayYAML, `source_chapter_count: 3`) {
		t.Fatalf("expected chapter count in yaml:\n%s", resp.ScreenplayYAML)
	}
}

func TestMockDomainConverterReturnsStableValidationError(t *testing.T) {
	converter := NewMockDomainConverter()

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "短篇",
		Content: "第一章\n只有一章",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != ErrCodeInsufficientChapters {
		t.Fatalf("expected %s, got %s", ErrCodeInsufficientChapters, appErr.Code)
	}
}

func TestDomainConverterUsesProviderOutput(t *testing.T) {
	provider := recordingProvider{
		screenplay: domain.Screenplay{
			SchemaVersion: "1.0",
			Title:         "Provider 输出",
			SourceType:    "novel",
			Language:      "zh-CN",
			Provider:      "test-provider",
			Mode:          "test",
			Characters: []domain.Character{
				{ID: "char_001", Name: "林舟", Role: "protagonist"},
			},
			SourceChapters: []domain.Chapter{
				{ID: "chapter_001", Title: "第一章", Order: 1, Summary: "一"},
				{ID: "chapter_002", Title: "第二章", Order: 2, Summary: "二"},
				{ID: "chapter_003", Title: "第三章", Order: 3, Summary: "三"},
			},
			Acts: []domain.Act{
				{
					ID:    "act_001",
					Title: "开端",
					Order: 1,
					Scenes: []domain.Scene{
						{
							ID:               "scene_001",
							SourceChapterIDs: []string{"chapter_001"},
							Heading:          domain.Heading{Location: "测试场景", Time: "日", Interior: true},
							Summary:          "测试摘要",
							Characters:       []string{"char_001"},
							Beats:            []domain.Beat{{Type: "action", Text: "测试动作"}},
						},
					},
				},
			},
		},
	}
	converter := NewDomainConverter(&provider)

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if !provider.called {
		t.Fatal("expected provider to be called")
	}
	if provider.input.Novel.Title != "雨夜来信" {
		t.Fatalf("unexpected provider input title: %q", provider.input.Novel.Title)
	}
	if resp.Mode != "test" {
		t.Fatalf("expected provider mode test, got %q", resp.Mode)
	}
	if !strings.Contains(resp.ScreenplayYAML, `provider: "test-provider"`) {
		t.Fatalf("expected YAML to include provider output:\n%s", resp.ScreenplayYAML)
	}
}

func TestDomainConverterReturnsProviderRawYAML(t *testing.T) {
	const rawYAML = "schema_version: \"1.0\"\nmetadata:\n  generated_by:\n    mode: \"api\"\n"
	provider := recordingProvider{rawYAML: rawYAML}
	converter := NewDomainConverter(&provider)

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if !provider.called {
		t.Fatal("expected provider to be called")
	}
	if resp.ScreenplayYAML != rawYAML {
		t.Fatalf("expected raw YAML to be returned unchanged:\n%s", resp.ScreenplayYAML)
	}
	if resp.ChapterCount != 3 {
		t.Fatalf("expected 3 chapters, got %d", resp.ChapterCount)
	}
	if resp.Mode != "api" {
		t.Fatalf("expected api mode for raw YAML output, got %q", resp.Mode)
	}
}

func TestDomainConverterLogsPipelineStages(t *testing.T) {
	const rawYAML = "schema_version: \"1.0\"\nmetadata:\n  generated_by:\n    mode: \"api\"\n"
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	ctx := observability.WithRequestID(observability.WithLogger(context.Background(), logger), "req_test")
	provider := recordingProvider{rawYAML: rawYAML}
	converter := NewDomainConverter(&provider)

	resp, err := converter.Convert(ctx, ConvertRequest{
		Title:     "雨夜来信",
		Content:   sampleConvertNovel,
		InputType: "text",
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	logs := decodeAppJSONLogs(t, logBuffer.String())
	started := findAppLogMessage(t, logs, "convert pipeline started")
	if started["request_id"] != "req_test" {
		t.Fatalf("request_id = %v, want req_test", started["request_id"])
	}
	if started["content_length"] != float64(len(sampleConvertNovel)) {
		t.Fatalf("content_length = %v, want %d", started["content_length"], len(sampleConvertNovel))
	}
	parsed := findAppLogMessage(t, logs, "novel parsed")
	if parsed["chapter_count"] != float64(3) {
		t.Fatalf("chapter_count = %v, want 3", parsed["chapter_count"])
	}
	completed := findAppLogMessage(t, logs, "screenplay generation completed")
	if _, ok := completed["duration_ms"]; !ok {
		t.Fatalf("expected duration_ms in completed log: %+v", completed)
	}
	if completed["chapter_count"] != float64(resp.ChapterCount) {
		t.Fatalf("chapter_count = %v, want %d", completed["chapter_count"], resp.ChapterCount)
	}
	if completed["mode"] != resp.Mode {
		t.Fatalf("mode = %v, want %s", completed["mode"], resp.Mode)
	}
	if completed["yaml_length"] != float64(len(rawYAML)) {
		t.Fatalf("yaml_length = %v, want %d", completed["yaml_length"], len(rawYAML))
	}
}

func TestDomainConverterMapsAIInvalidYAMLError(t *testing.T) {
	provider := recordingProvider{err: ai.YAMLValidationError{
		Path:    "metadata.title",
		Message: "metadata.title 不能为空",
	}}
	converter := NewDomainConverter(&provider)

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})

	assertAppError(t, err, ErrorCodeAIInvalidYAML, "AI 返回的 YAML 未通过结构校验，请重试。")
}

func TestDomainConverterMapsAIProviderConfigError(t *testing.T) {
	provider := recordingProvider{err: ai.ErrDeepSeekAPIKeyRequired}
	converter := NewDomainConverter(&provider)

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})

	assertAppError(t, err, ErrorCodeAIProviderNotConfigured, "AI provider 配置不完整，请检查 DeepSeek API key、Base URL 和模型配置。")
}

func TestDomainConverterMapsAIGenerationError(t *testing.T) {
	provider := recordingProvider{err: errors.New("deepseek api returned status 500")}
	converter := NewDomainConverter(&provider)

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})

	assertAppError(t, err, ErrorCodeAIGenerationFailed, "AI 生成失败，请稍后重试。")
}

func TestDomainConverterMapsAITimeoutError(t *testing.T) {
	provider := recordingProvider{err: context.DeadlineExceeded}
	converter := NewDomainConverter(&provider)

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})

	assertAppError(t, err, ErrorCodeAITimeout, "AI 生成超时，请稍后重试或调大 DeepSeek 超时时间。")
}

func TestDomainConverterMapsNetTimeoutError(t *testing.T) {
	provider := recordingProvider{err: timeoutError{}}
	converter := NewDomainConverter(&provider)

	_, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})

	assertAppError(t, err, ErrorCodeAITimeout, "AI 生成超时，请稍后重试或调大 DeepSeek 超时时间。")
}

func TestFallbackConverterUsesMockWhenPrimaryFails(t *testing.T) {
	primary := &recordingConverter{err: NewError(ErrorCodeAIGenerationFailed, "AI 生成失败，请稍后重试。")}
	fallback := NewMockDomainConverter()
	converter := NewFallbackConverter(primary, fallback)

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if !primary.called {
		t.Fatal("expected primary converter to be called")
	}
	if resp.Mode != "mock" {
		t.Fatalf("expected fallback mock mode, got %q", resp.Mode)
	}
	if resp.ChapterCount != 3 {
		t.Fatalf("expected 3 chapters, got %d", resp.ChapterCount)
	}
	if !strings.Contains(resp.ScreenplayYAML, `mode: "mock"`) {
		t.Fatalf("expected mock YAML output:\n%s", resp.ScreenplayYAML)
	}
}

func TestFallbackConverterLogsFallbackActivation(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	ctx := observability.WithRequestID(observability.WithLogger(context.Background(), logger), "req_fallback")
	primary := &recordingConverter{err: NewError(ErrorCodeAIGenerationFailed, "AI generation failed")}
	fallback := &recordingConverter{response: ConvertResponse{
		ScreenplayYAML: "schema_version: \"1.0\"\n",
		ChapterCount:   3,
		Mode:           "mock",
	}}
	converter := NewFallbackConverter(primary, fallback)

	resp, err := converter.Convert(ctx, ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	logs := decodeAppJSONLogs(t, logBuffer.String())
	entry := findAppLogMessage(t, logs, "convert fallback activated")
	if entry["request_id"] != "req_fallback" {
		t.Fatalf("request_id = %v, want req_fallback", entry["request_id"])
	}
	if entry["error_code"] != ErrorCodeAIGenerationFailed {
		t.Fatalf("error_code = %v, want %s", entry["error_code"], ErrorCodeAIGenerationFailed)
	}
	completed := findAppLogMessage(t, logs, "convert fallback completed")
	if _, ok := completed["duration_ms"]; !ok {
		t.Fatalf("expected duration_ms in fallback completed log: %+v", completed)
	}
	if completed["fallback_mode"] != resp.Mode {
		t.Fatalf("fallback_mode = %v, want %s", completed["fallback_mode"], resp.Mode)
	}
	if completed["chapter_count"] != float64(resp.ChapterCount) {
		t.Fatalf("chapter_count = %v, want %d", completed["chapter_count"], resp.ChapterCount)
	}
	if completed["yaml_length"] != float64(len(resp.ScreenplayYAML)) {
		t.Fatalf("yaml_length = %v, want %d", completed["yaml_length"], len(resp.ScreenplayYAML))
	}
}

func TestFallbackConverterReturnsPrimarySuccess(t *testing.T) {
	const rawYAML = "schema_version: \"1.0\"\n"
	primary := &recordingConverter{response: ConvertResponse{
		ScreenplayYAML: rawYAML,
		ChapterCount:   3,
		Mode:           "api",
	}}
	fallback := &recordingConverter{}
	converter := NewFallbackConverter(primary, fallback)

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:   "雨夜来信",
		Content: sampleConvertNovel,
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if resp.ScreenplayYAML != rawYAML || resp.Mode != "api" {
		t.Fatalf("unexpected primary response: %+v", resp)
	}
	if fallback.called {
		t.Fatal("fallback should not be called when primary succeeds")
	}
}

func TestFallbackConverterDoesNotHideInputError(t *testing.T) {
	primaryErr := NewError(ErrCodeInsufficientChapters, "至少需要 3 个章节才能生成剧本初稿。")
	primary := &recordingConverter{err: primaryErr}
	fallback := &recordingConverter{}
	converter := NewFallbackConverter(primary, fallback)

	_, err := converter.Convert(context.Background(), ConvertRequest{})

	if !errors.Is(err, primaryErr) {
		t.Fatalf("error = %v, want primary error %v", err, primaryErr)
	}
	if fallback.called {
		t.Fatal("fallback should not be called for input validation errors")
	}
}

type recordingProvider struct {
	called     bool
	input      ai.GenerateInput
	screenplay domain.Screenplay
	rawYAML    string
	err        error
}

func (p *recordingProvider) GenerateScreenplay(_ context.Context, input ai.GenerateInput) (ai.GenerateOutput, error) {
	p.called = true
	p.input = input
	if p.err != nil {
		return ai.GenerateOutput{}, p.err
	}
	return ai.GenerateOutput{Screenplay: p.screenplay, RawYAML: p.rawYAML}, nil
}

func assertAppError(t *testing.T, err error, code, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected code %s, got %s", code, appErr.Code)
	}
	if appErr.Message != message {
		t.Fatalf("expected message %q, got %q", message, appErr.Message)
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func decodeAppJSONLogs(t *testing.T, raw string) []map[string]any {
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

func findAppLogMessage(t *testing.T, logs []map[string]any, message string) map[string]any {
	t.Helper()

	for _, entry := range logs {
		if entry["msg"] == message {
			return entry
		}
	}
	t.Fatalf("missing log message %q in %+v", message, logs)
	return nil
}

type recordingConverter struct {
	called   bool
	response ConvertResponse
	err      error
}

func (c *recordingConverter) Convert(_ context.Context, _ ConvertRequest) (ConvertResponse, error) {
	c.called = true
	if c.err != nil {
		return ConvertResponse{}, c.err
	}
	return c.response, nil
}

const sampleConvertNovel = `# 第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

# 第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

# 第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
