package app

import (
	"context"
	"errors"

	"github.com/qingketsing/novel2script/backend/internal/observability"
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
	logger := observability.Logger(ctx)
	requestID := observability.RequestID(ctx)
	logger.WarnContext(ctx, "convert fallback activated",
		"request_id", requestID,
		"error_code", fallbackErrorCode(err),
	)

	resp, fallbackErr := c.fallback.Convert(ctx, req)
	if fallbackErr != nil {
		return ConvertResponse{}, fallbackErr
	}
	logger.InfoContext(ctx, "convert fallback completed",
		"request_id", requestID,
		"fallback_mode", resp.Mode,
		"chapter_count", resp.ChapterCount,
		"yaml_length", len(resp.ScreenplayYAML),
	)
	return resp, nil
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

func fallbackErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}
