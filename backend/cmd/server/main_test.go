package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/app"
)

func TestNewHandlerWiresMockDomainConverter(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "雨夜来信",
		"content": "# 第一章 雨夜来信\n林舟在雨夜收到一封没有署名的信。\n\n# 第二章 旧书店\n林舟来到旧书店，寻找姐姐留下的线索。\n\n# 第三章 街灯\n街灯忽明忽暗，线索指向城市另一端。",
		"input_type": "md"
	}`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body app.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ChapterCount != 3 {
		t.Fatalf("expected 3 chapters, got %d", body.ChapterCount)
	}
	if body.Mode != "mock" {
		t.Fatalf("expected mock mode, got %q", body.Mode)
	}
	required := []string{
		`schema_version: "1.0"`,
		`source_chapter_count: 3`,
		`characters:`,
		`source_chapters:`,
		`screenplay:`,
	}
	for _, want := range required {
		if !strings.Contains(body.ScreenplayYAML, want) {
			t.Fatalf("expected screenplay YAML to contain %q\n%s", want, body.ScreenplayYAML)
		}
	}
}
