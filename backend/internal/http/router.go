package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/qingketsing/novel2script/backend/internal/app"
	"github.com/qingketsing/novel2script/backend/internal/observability"
)

const maxUploadBytes = 2 * 1024 * 1024
const localFrontendOrigin = "http://localhost:5173"

type errorResponse struct {
	Error app.AppError `json:"error"`
}

// NewRouter 注册后端 MVP 所需的 HTTP 路由，并把转换能力注入到 API 层。
func NewRouter(converter app.Converter) http.Handler {
	return NewRouterWithLogger(converter, slog.Default())
}

func NewRouterWithLogger(converter app.Converter, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/convert", handleConvert(logger, converter))
	mux.HandleFunc("POST /api/convert/upload", handleConvertUpload(logger, converter))
	return withRequestLogging(logger, withCORS(mux))
}

// withCORS 允许本地前端开发服务器调用后端 API，并处理浏览器预检请求。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == localFrontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", localFrontendOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleHealth 提供轻量健康检查，供本地启动和部署探活使用。
func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = observability.NewRequestID()
		}
		ctx := observability.WithRequestID(observability.WithLogger(r.Context(), logger), requestID)
		r = r.WithContext(ctx)

		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)

		logger.InfoContext(ctx, "http request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"content_length", r.ContentLength,
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleConvert 负责请求解析、入参校验、调用转换器，并统一输出 JSON 响应。
func handleConvert(logger *slog.Logger, converter app.Converter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req app.ConvertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAppError(w, http.StatusBadRequest, app.AppError{
				Code:    app.ErrorCodeInvalidJSON,
				Message: "请求体必须是合法 JSON。",
			})
			return
		}
		if appErr, ok := validateConvertRequest(req); !ok {
			writeAppError(w, http.StatusBadRequest, appErr)
			return
		}

		resp, err := converter.Convert(r.Context(), req)
		if err != nil {
			writeConvertError(w, err)
			return
		}

		logger.InfoContext(r.Context(), "convert request completed",
			"request_id", observability.RequestID(r.Context()),
			"input_type", req.InputType,
			"content_length", len(req.Content),
			"title_present", strings.TrimSpace(req.Title) != "",
			"chapter_count", resp.ChapterCount,
			"mode", resp.Mode,
			"yaml_length", len(resp.ScreenplayYAML),
		)
		writeJSON(w, http.StatusOK, resp)
	}
}

// handleConvertUpload 读取 multipart 上传文件，并复用转换器完成小说到剧本 YAML 的生成。
func handleConvertUpload(logger *slog.Logger, converter app.Converter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			writeAppError(w, http.StatusBadRequest, app.AppError{
				Code:    app.ErrorCodeInvalidInput,
				Message: "上传文件过大或表单格式不正确。",
			})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeAppError(w, http.StatusBadRequest, app.AppError{
				Code:    app.ErrorCodeInvalidInput,
				Message: "请上传 .txt 或 .md 小说文件。",
			})
			return
		}
		defer file.Close()

		inputType, ok := inputTypeFromFilename(header.Filename)
		if !ok {
			writeAppError(w, http.StatusBadRequest, app.AppError{
				Code:    app.ErrorCodeInvalidInput,
				Message: "当前仅支持 .txt 或 .md 文件。",
			})
			return
		}

		content, appErr, ok := readUploadedNovel(file)
		if !ok {
			writeAppError(w, http.StatusBadRequest, appErr)
			return
		}

		req := app.ConvertRequest{
			Title:     strings.TrimSpace(r.FormValue("title")),
			Content:   content,
			InputType: inputType,
		}
		resp, err := converter.Convert(r.Context(), req)
		if err != nil {
			writeConvertError(w, err)
			return
		}

		logger.InfoContext(r.Context(), "convert upload request completed",
			"request_id", observability.RequestID(r.Context()),
			"input_type", req.InputType,
			"content_length", len(req.Content),
			"title_present", strings.TrimSpace(req.Title) != "",
			"chapter_count", resp.ChapterCount,
			"mode", resp.Mode,
			"yaml_length", len(resp.ScreenplayYAML),
		)
		writeJSON(w, http.StatusOK, resp)
	}
}

// validateConvertRequest 只做 HTTP 请求级校验，章节数量等业务校验留给领域管线处理。
func validateConvertRequest(req app.ConvertRequest) (app.AppError, bool) {
	if strings.TrimSpace(req.Content) == "" {
		return app.AppError{
			Code:    app.ErrorCodeInvalidInput,
			Message: "小说正文不能为空，请上传文件或粘贴正文。",
		}, false
	}

	inputType := strings.TrimSpace(strings.ToLower(req.InputType))
	if inputType == "" {
		inputType = "text"
	}
	switch inputType {
	case "text", "txt", "md":
		return app.AppError{}, true
	default:
		return app.AppError{
			Code:    app.ErrorCodeInvalidInput,
			Message: "当前仅支持 text、txt 或 md 输入类型。",
		}, false
	}
}

func inputTypeFromFilename(filename string) (string, bool) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".txt":
		return "txt", true
	case ".md":
		return "md", true
	default:
		return "", false
	}
}

func readUploadedNovel(file io.Reader) (string, app.AppError, bool) {
	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		return "", app.AppError{
			Code:    app.ErrorCodeInvalidInput,
			Message: "读取上传文件失败。",
		}, false
	}
	if len(data) > maxUploadBytes {
		return "", app.AppError{
			Code:    app.ErrorCodeInvalidInput,
			Message: "上传文件不能超过 2MB。",
		}, false
	}
	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", app.AppError{
			Code:    app.ErrorCodeInvalidInput,
			Message: "上传文件内容不能为空。",
		}, false
	}
	return content, app.AppError{}, true
}

func writeConvertError(w http.ResponseWriter, err error) {
	var appErr *app.AppError
	if errors.As(err, &appErr) {
		writeAppError(w, http.StatusBadRequest, *appErr)
		return
	}
	writeAppError(w, http.StatusInternalServerError, app.AppError{
		Code:    app.ErrorCodeInternalError,
		Message: "服务暂时不可用，请稍后重试。",
	})
}

// writeAppError 将应用错误包装为统一的 {"error": ...} JSON 结构。
func writeAppError(w http.ResponseWriter, status int, appErr app.AppError) {
	writeJSON(w, status, errorResponse{Error: appErr})
}

// writeJSON 统一设置 JSON 响应头和状态码，避免各 handler 重复处理。
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
