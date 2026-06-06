package app

import (
	"context"
	"errors"
)

type FallbackConverter struct {
	primary  Converter
	fallback Converter
}

func NewFallbackConverter(primary, fallback Converter) Converter {
	return FallbackConverter{
		primary:  primary,
		fallback: fallback,
	}
}

func (c FallbackConverter) Convert(ctx context.Context, req ConvertRequest) (ConvertResponse, error) {
	resp, err := c.primary.Convert(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !isAIFallbackError(err) {
		return ConvertResponse{}, err
	}
	return c.fallback.Convert(ctx, req)
}

func isAIFallbackError(err error) bool {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return false
	}
	switch appErr.Code {
	case ErrorCodeAIProviderNotConfigured,
		ErrorCodeAIGenerationFailed,
		ErrorCodeAIInvalidYAML,
		ErrorCodeAITimeout:
		return true
	default:
		return false
	}
}
