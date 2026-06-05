package app

import (
	"context"
	"errors"
	"strings"
	"testing"
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

const sampleConvertNovel = `# 第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

# 第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

# 第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
