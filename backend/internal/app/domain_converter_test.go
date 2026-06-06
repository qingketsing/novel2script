package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/domain"
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

const sampleConvertNovel = `# 第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

# 第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

# 第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
