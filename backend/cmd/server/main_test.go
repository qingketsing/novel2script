package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/app"
	"github.com/qingketsing/novel2script/backend/internal/config"
)

func TestNewHandlerWiresMockDomainConverter(t *testing.T) {
	server := httptest.NewServer(mustNewHandler(t, config.Config{AIMode: "mock"}))
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
	server := httptest.NewServer(mustNewHandler(t, config.Config{AIMode: "mock"}))
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

func TestNewHandlerUploadSmokeCoversMockThreeAndFiveChapterNovels(t *testing.T) {
	server := httptest.NewServer(mustNewHandler(t, config.Config{AIMode: "mock"}))
	defer server.Close()

	threeChapterNovel, err := os.ReadFile("../../../docs/examples/novel-example.md")
	if err != nil {
		t.Fatalf("read demo novel: %v", err)
	}

	tests := []struct {
		name         string
		filename     string
		content      string
		wantChapters int
		requiredYAML []string
	}{
		{
			name:         "three chapter demo novel",
			filename:     "novel-example.md",
			content:      string(threeChapterNovel),
			wantChapters: 3,
			requiredYAML: []string{
				`schema_version: "1.0"`,
				`source_chapter_count: 3`,
				`characters:`,
				`source_chapters:`,
				`screenplay:`,
				`scene_001`,
				`chapter_003`,
				`beats:`,
			},
		},
		{
			name:         "five chapter demo novel",
			filename:     "five-chapter-demo.md",
			content:      fiveChapterMarkdownNovel,
			wantChapters: 5,
			requiredYAML: []string{
				`schema_version: "1.0"`,
				`source_chapter_count: 5`,
				`characters:`,
				`source_chapters:`,
				`screenplay:`,
				`scene_005`,
				`chapter_005`,
				`beats:`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postServerUpload(server.URL+"/api/convert/upload", tt.filename, tt.content, "Smoke Demo")
			if err != nil {
				t.Fatalf("POST /api/convert/upload failed: %v", err)
			}
			defer resp.Body.Close()

			assertUploadSmokeResponse(t, resp, tt.wantChapters, tt.requiredYAML)
		})
	}
}

func TestNewHandlerRejectsMissingDeepSeekAPIKey(t *testing.T) {
	_, err := newHandler(config.Config{AIMode: "deepseek"})

	if !errors.Is(err, ai.ErrDeepSeekAPIKeyRequired) {
		t.Fatalf("error = %v, want %v", err, ai.ErrDeepSeekAPIKeyRequired)
	}
}

func TestNewHandlerFallsBackToMockWhenDeepSeekConfigMissing(t *testing.T) {
	server := httptest.NewServer(mustNewHandler(t, config.Config{
		AIMode:           "deepseek",
		AIFallbackToMock: true,
	}))
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
	if body.Mode != "mock" {
		t.Fatalf("expected fallback mock mode, got %q", body.Mode)
	}
}

func TestNewHandlerRejectsUnknownModeWhenFallbackEnabled(t *testing.T) {
	_, err := newHandler(config.Config{
		AIMode:           "unknown",
		AIFallbackToMock: true,
	})

	if !errors.Is(err, app.ErrUnsupportedAIMode) {
		t.Fatalf("error = %v, want %v", err, app.ErrUnsupportedAIMode)
	}
}

func TestNewHandlerWiresConfiguredDeepSeekProvider(t *testing.T) {
	handler, err := newHandler(config.Config{
		AIMode:          "deepseek",
		DeepSeekAPIKey:  "test-api-key",
		DeepSeekBaseURL: "https://api.deepseek.com",
		DeepSeekModel:   "deepseek-v4",
	})

	if err != nil {
		t.Fatalf("newHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("handler is nil, want configured handler")
	}
}

func mustNewHandler(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()

	handler, err := newHandler(cfg)
	if err != nil {
		t.Fatalf("newHandler() error = %v", err)
	}
	return handler
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

func postServerUpload(url, filename, content, title string) (*http.Response, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("title", title); err != nil {
		return nil, err
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return http.DefaultClient.Do(req)
}

func assertUploadSmokeResponse(t *testing.T, resp *http.Response, wantChapters int, requiredYAML []string) {
	t.Helper()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body app.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ChapterCount != wantChapters {
		t.Fatalf("expected %d chapters, got %d", wantChapters, body.ChapterCount)
	}
	if body.Mode != "mock" {
		t.Fatalf("expected mock mode, got %q", body.Mode)
	}
	for _, want := range requiredYAML {
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

const fiveChapterMarkdownNovel = `# Chapter 1
The first signal appears beside the old station.

# Chapter 2
The team follows the signal into the archive room.

# Chapter 3
A missing ledger reveals who changed the route.

# Chapter 4
The warning reaches the control desk before midnight.

# Chapter 5
The final light turns back on and the train slows safely.`
