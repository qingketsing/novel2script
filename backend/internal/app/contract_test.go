package app

import (
	"context"
	"errors"
	"testing"
)

func TestAppErrorImplementsError(t *testing.T) {
	err := NewError("INSUFFICIENT_CHAPTERS", "至少需要 3 个章节。")

	if err.Error() != "INSUFFICIENT_CHAPTERS: 至少需要 3 个章节。" {
		t.Fatalf("unexpected error string: %q", err.Error())
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected NewError result to match *AppError")
	}
}

func TestPlaceholderConverterReturnsStableMockResponse(t *testing.T) {
	converter := NewPlaceholderConverter()

	resp, err := converter.Convert(context.Background(), ConvertRequest{
		Title:     "示例小说",
		Content:   "第一章\n内容\n第二章\n内容\n第三章\n内容",
		InputType: "text",
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}

	if resp.ChapterCount != 0 {
		t.Fatalf("expected placeholder chapter count 0, got %d", resp.ChapterCount)
	}
	if resp.Mode != "mock" {
		t.Fatalf("expected mock mode, got %q", resp.Mode)
	}
	if resp.ScreenplayYAML == "" {
		t.Fatal("expected placeholder screenplay yaml")
	}
}
