package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

func TestNewHandlerConvertsUploadedMarkdownNovel(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	if err := writer.WriteField("title", "雨夜来信"); err != nil {
		t.Fatalf("write title field: %v", err)
	}
	file, err := writer.CreateFormFile("file", "demo.md")
	if err != nil {
		t.Fatalf("create file field: %v", err)
	}
	if _, err := file.Write([]byte(demoMarkdownNovel)); err != nil {
		t.Fatalf("write file content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/convert/upload", &requestBody)
	if err != nil {
		t.Fatalf("create upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertDemoConvertResponse(t, resp)
}

func assertDemoConvertResponse(t *testing.T, resp *http.Response) {
	t.Helper()

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
		`provider: "deepseek-v4"`,
		`mode: "mock"`,
		`characters:`,
		`source_chapters:`,
		`screenplay:`,
		`beats:`,
	}
	for _, want := range required {
		if !strings.Contains(body.ScreenplayYAML, want) {
			t.Fatalf("expected screenplay YAML to contain %q\n%s", want, body.ScreenplayYAML)
		}
	}
}

const demoMarkdownNovel = `# 第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

# 第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

# 第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
