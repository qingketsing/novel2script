package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/app"
)

type errorResponse struct {
	Error app.AppError `json:"error"`
}

func NewRouter(converter app.Converter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/convert", handleConvert(converter))
	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleConvert(converter app.Converter) http.HandlerFunc {
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
			var appErr *app.AppError
			if errors.As(err, &appErr) {
				writeAppError(w, http.StatusBadRequest, *appErr)
				return
			}
			writeAppError(w, http.StatusInternalServerError, app.AppError{
				Code:    app.ErrorCodeInternalError,
				Message: "服务暂时不可用，请稍后重试。",
			})
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

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

func writeAppError(w http.ResponseWriter, status int, appErr app.AppError) {
	writeJSON(w, status, errorResponse{Error: appErr})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
