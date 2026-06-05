package app

import "fmt"

const (
	ErrorCodeInvalidJSON   = "INVALID_JSON"
	ErrorCodeInvalidInput  = "INVALID_INPUT"
	ErrorCodeInternalError = "INTERNAL_ERROR"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewError(code, message string) error {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
