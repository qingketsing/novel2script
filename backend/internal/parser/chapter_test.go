package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePlainTextChapters(t *testing.T) {
	content := readFixture(t, "backend/testdata/plain_three_chapters.txt")

	novel, err := ParseNovel("雨夜来信", content)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	if novel.Title != "雨夜来信" {
		t.Fatalf("unexpected title: %q", novel.Title)
	}
	if len(novel.Chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(novel.Chapters))
	}
	if novel.Chapters[0].ID != "chapter_001" || novel.Chapters[0].Order != 1 {
		t.Fatalf("unexpected first chapter identity: %+v", novel.Chapters[0])
	}
	if novel.Chapters[1].Title != "第二章 旧书店" {
		t.Fatalf("unexpected second chapter title: %q", novel.Chapters[1].Title)
	}
}

func TestParseMarkdownChapters(t *testing.T) {
	content := readFixture(t, "backend/testdata/markdown_three_chapters.md")

	novel, err := ParseNovel("Markdown 小说", content)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	if len(novel.Chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(novel.Chapters))
	}
	if novel.Chapters[0].Title != "第一章 雨夜来信" {
		t.Fatalf("unexpected first chapter title: %q", novel.Chapters[0].Title)
	}
	if novel.Chapters[2].ID != "chapter_003" {
		t.Fatalf("unexpected third chapter id: %q", novel.Chapters[2].ID)
	}
}

func readFixture(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile(filepath.Join("..", "..", "testdata", filepath.Base(path)))
	}
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(data)
}
