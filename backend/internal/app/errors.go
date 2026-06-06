package app

import "fmt"

const (
	ErrorCodeInvalidJSON             = "INVALID_JSON"
	ErrorCodeInvalidInput            = "INVALID_INPUT"
	ErrorCodeInternalError           = "INTERNAL_ERROR"
	ErrorCodeAIProviderNotConfigured = "AI_PROVIDER_NOT_CONFIGURED"
	ErrorCodeAIGenerationFailed      = "AI_GENERATION_FAILED"
	ErrorCodeAIInvalidYAML           = "AI_INVALID_YAML"
	ErrorCodeAITimeout               = "AI_TIMEOUT"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewError 创建应用层错误，供 HTTP 层转换为统一 JSON 错误响应。
func NewError(code, message string) error {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Error 返回带错误码的可读错误文本，便于日志和测试断言。
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
