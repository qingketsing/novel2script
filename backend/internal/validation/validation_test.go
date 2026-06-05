package validation

import (
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

func TestValidateNovelRejectsEmptyText(t *testing.T) {
	err := ValidateInput("  \n\t ")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != CodeEmptyText {
		t.Fatalf("expected %s, got %s", CodeEmptyText, err.Code)
	}
	if err.Message == "" {
		t.Fatal("expected user-facing message")
	}
}

func TestValidateNovelRejectsInsufficientChapters(t *testing.T) {
	err := ValidateNovel(domain.Novel{
		Title: "短篇",
		Chapters: []domain.Chapter{
			{ID: "chapter_001", Title: "第一章", Order: 1},
			{ID: "chapter_002", Title: "第二章", Order: 2},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Code != CodeInsufficientChapters {
		t.Fatalf("expected %s, got %s", CodeInsufficientChapters, err.Code)
	}
}
