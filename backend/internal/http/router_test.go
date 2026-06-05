package httpapi

import (
	"context"
	"encoding/json"
	"errors"
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
