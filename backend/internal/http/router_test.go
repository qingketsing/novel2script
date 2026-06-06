package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/app"
)

func TestHealthEndpoint(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got %q", body["status"])
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("expected json content type, got %q", got)
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/missing")
	if err != nil {
		t.Fatalf("GET /missing failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestConvertEndpointRejectsWrongMethod(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/convert")
	if err != nil {
		t.Fatalf("GET /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestCORSPreflightAllowsLocalFrontend(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodOptions, server.URL+"/api/convert", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

func TestCORSHeadersAreSetOnAPIResponse(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{
		response: app.ConvertResponse{
			ScreenplayYAML: "schema_version: \"1.0\"\n",
			ChapterCount:   3,
			Mode:           "mock",
		},
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/convert", strings.NewReader(`{
		"title": "示例小说",
		"content": "第一章\n内容\n第二章\n内容\n第三章\n内容",
		"input_type": "text"
	}`))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

func TestConvertEndpointWritesRequestLifecycleLogs(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	server := httptest.NewServer(NewRouterWithLogger(stubConverter{
		response: app.ConvertResponse{
			ScreenplayYAML: "schema_version: \"1.0\"\n",
			ChapterCount:   3,
			Mode:           "mock",
		},
	}, logger))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "示例小说",
		"content": "第一章\n内容\n第二章\n内容\n第三章\n内容",
		"input_type": "text"
	}`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	logs := decodeJSONLogs(t, logBuffer.String())
	completed := findLogMessage(t, logs, "convert request completed")
	if completed["request_id"] == "" {
		t.Fatalf("expected request_id in convert log: %+v", completed)
	}
	if completed["chapter_count"] != float64(3) {
		t.Fatalf("chapter_count = %v, want 3", completed["chapter_count"])
	}
	if completed["mode"] != "mock" {
		t.Fatalf("mode = %v, want mock", completed["mode"])
	}
	if completed["yaml_length"] != float64(len("schema_version: \"1.0\"\n")) {
		t.Fatalf("yaml_length = %v, want %d", completed["yaml_length"], len("schema_version: \"1.0\"\n"))
	}

	httpLog := findLogMessage(t, logs, "http request completed")
	if httpLog["request_id"] != completed["request_id"] {
		t.Fatalf("http request_id = %v, want %v", httpLog["request_id"], completed["request_id"])
	}
	if httpLog["method"] != http.MethodPost {
		t.Fatalf("method = %v, want POST", httpLog["method"])
	}
	if httpLog["path"] != "/api/convert" {
		t.Fatalf("path = %v, want /api/convert", httpLog["path"])
	}
	if httpLog["status"] != float64(http.StatusOK) {
		t.Fatalf("status = %v, want 200", httpLog["status"])
	}
	if _, ok := httpLog["duration_ms"]; !ok {
		t.Fatalf("expected duration_ms in http log: %+v", httpLog)
	}
}

func TestConvertEndpointReturnsConverterResponse(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{
		response: app.ConvertResponse{
			ScreenplayYAML: "schema_version: \"1.0\"\n",
			ChapterCount:   3,
			Mode:           "mock",
		},
	}))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "示例小说",
		"content": "第一章\n内容\n第二章\n内容\n第三章\n内容",
		"input_type": "text"
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
	if body.ScreenplayYAML == "" || body.ChapterCount != 3 || body.Mode != "mock" {
		t.Fatalf("unexpected response: %+v", body)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("expected json content type, got %q", got)
	}
}

func TestConvertEndpointReturnsAppError(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{
		err: app.NewError("INSUFFICIENT_CHAPTERS", "至少需要 3 个章节。"),
	}))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "示例小说",
		"content": "第一章\n内容",
		"input_type": "text"
	}`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "INSUFFICIENT_CHAPTERS" {
		t.Fatalf("unexpected error response: %+v", body)
	}
}

func TestConvertEndpointRejectsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{}))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{bad json`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "INVALID_JSON" {
		t.Fatalf("unexpected error response: %+v", body)
	}
}

func TestConvertEndpointRejectsEmptyContent(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "示例小说",
		"content": "   ",
		"input_type": "text"
	}`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if converter.called {
		t.Fatal("converter should not be called for invalid input")
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != app.ErrorCodeInvalidInput {
		t.Fatalf("unexpected error response: %+v", body)
	}
}

func TestConvertEndpointRejectsUnsupportedInputType(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := http.Post(server.URL+"/api/convert", "application/json", strings.NewReader(`{
		"title": "示例小说",
		"content": "第一章\n内容\n第二章\n内容\n第三章\n内容",
		"input_type": "pdf"
	}`))
	if err != nil {
		t.Fatalf("POST /api/convert failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if converter.called {
		t.Fatal("converter should not be called for invalid input")
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != app.ErrorCodeInvalidInput {
		t.Fatalf("unexpected error response: %+v", body)
	}
}

func TestConvertUploadEndpointReturnsConverterResponse(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := postUpload(server.URL+"/api/convert/upload", "novel.md", sampleUploadNovel, "雨夜来信")
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !converter.called {
		t.Fatal("expected converter to be called")
	}
	if converter.request.Title != "雨夜来信" {
		t.Fatalf("unexpected title: %q", converter.request.Title)
	}
	if converter.request.InputType != "md" {
		t.Fatalf("unexpected input type: %q", converter.request.InputType)
	}
	if !strings.Contains(converter.request.Content, "第一章") {
		t.Fatalf("expected uploaded content, got %q", converter.request.Content)
	}

	var body app.ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ChapterCount != 3 || body.Mode != "mock" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestConvertUploadEndpointRejectsMissingFile(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("title", "雨夜来信"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/convert/upload", &body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertUploadError(t, resp, app.ErrorCodeInvalidInput)
	if converter.called {
		t.Fatal("converter should not be called for missing file")
	}
}

func TestConvertUploadEndpointRejectsEmptyFile(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := postUpload(server.URL+"/api/convert/upload", "novel.txt", "", "空文件")
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertUploadError(t, resp, app.ErrorCodeInvalidInput)
	if converter.called {
		t.Fatal("converter should not be called for empty file")
	}
}

func TestConvertUploadEndpointRejectsUnsupportedFileType(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := postUpload(server.URL+"/api/convert/upload", "novel.pdf", sampleUploadNovel, "雨夜来信")
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertUploadError(t, resp, app.ErrorCodeInvalidInput)
	if converter.called {
		t.Fatal("converter should not be called for unsupported file type")
	}
}

func TestConvertUploadEndpointRejectsOversizedFile(t *testing.T) {
	converter := &recordingConverter{}
	server := httptest.NewServer(NewRouter(converter))
	defer server.Close()

	resp, err := postUpload(server.URL+"/api/convert/upload", "novel.txt", strings.Repeat("a", maxUploadBytes+1), "大文件")
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertUploadError(t, resp, app.ErrorCodeInvalidInput)
	if converter.called {
		t.Fatal("converter should not be called for oversized file")
	}
}

func TestConvertUploadEndpointReturnsChapterValidationError(t *testing.T) {
	server := httptest.NewServer(NewRouter(stubConverter{
		err: app.NewError("INSUFFICIENT_CHAPTERS", "至少需要 3 个章节。"),
	}))
	defer server.Close()

	resp, err := postUpload(server.URL+"/api/convert/upload", "novel.txt", "第一章\n只有一章", "短篇")
	if err != nil {
		t.Fatalf("POST /api/convert/upload failed: %v", err)
	}
	defer resp.Body.Close()

	assertUploadError(t, resp, "INSUFFICIENT_CHAPTERS")
}

type stubConverter struct {
	response app.ConvertResponse
	err      error
}

func (s stubConverter) Convert(_ context.Context, _ app.ConvertRequest) (app.ConvertResponse, error) {
	if s.err != nil {
		return app.ConvertResponse{}, s.err
	}
	if s.response.ScreenplayYAML == "" {
		return app.ConvertResponse{}, errors.New("unexpected stub call")
	}
	return s.response, nil
}

type recordingConverter struct {
	called  bool
	request app.ConvertRequest
}

func (r *recordingConverter) Convert(_ context.Context, req app.ConvertRequest) (app.ConvertResponse, error) {
	r.called = true
	r.request = req
	return app.ConvertResponse{
		ScreenplayYAML: "schema_version: \"1.0\"\n",
		ChapterCount:   3,
		Mode:           "mock",
	}, nil
}

func postUpload(url, filename, content, title string) (*http.Response, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if title != "" {
		if err := writer.WriteField("title", title); err != nil {
			return nil, err
		}
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

func assertUploadError(t *testing.T, resp *http.Response, wantCode string) {
	t.Helper()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var body errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != wantCode {
		t.Fatalf("expected error code %s, got %+v", wantCode, body)
	}
}

func assertCORSHeaders(t *testing.T, resp *http.Response) {
	t.Helper()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") || !strings.Contains(got, "OPTIONS") {
		t.Fatalf("unexpected allow methods: %q", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Content-Type") {
		t.Fatalf("unexpected allow headers: %q", got)
	}
}

func decodeJSONLogs(t *testing.T, raw string) []map[string]any {
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

func findLogMessage(t *testing.T, logs []map[string]any, message string) map[string]any {
	t.Helper()

	for _, entry := range logs {
		if entry["msg"] == message {
			return entry
		}
	}
	t.Fatalf("missing log message %q in %+v", message, logs)
	return nil
}

const sampleUploadNovel = `# 第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

# 第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

# 第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
