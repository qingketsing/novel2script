package app

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/exporter"
	"github.com/qingketsing/novel2script/backend/internal/observability"
	"github.com/qingketsing/novel2script/backend/internal/parser"
	"github.com/qingketsing/novel2script/backend/internal/validation"
)

const (
	ErrCodeEmptyText            = validation.CodeEmptyText
	ErrCodeInsufficientChapters = validation.CodeInsufficientChapters
)

type DomainConverter struct {
	provider ai.Provider
}

func NewMockDomainConverter() Converter {
	return NewDomainConverter(ai.NewMockProvider())
}

func NewDomainConverter(provider ai.Provider) Converter {
	return DomainConverter{provider: provider}
}

func (c DomainConverter) Convert(ctx context.Context, req ConvertRequest) (ConvertResponse, error) {
	logger := observability.Logger(ctx)
	requestID := observability.RequestID(ctx)
	logger.InfoContext(ctx, "convert pipeline started",
		"request_id", requestID,
		"input_type", req.InputType,
		"content_length", len(req.Content),
		"title_present", strings.TrimSpace(req.Title) != "",
	)

	if err := validation.ValidateInput(req.Content); err != nil {
		appErr := NewError(err.Code, err.Message)
		logger.WarnContext(ctx, "convert input validation failed",
			"request_id", requestID,
			"error_code", err.Code,
		)
		return ConvertResponse{}, appErr
	}

	parseStart := time.Now()
	novel, err := parser.ParseNovel(req.Title, req.Content)
	if err != nil {
		logger.WarnContext(ctx, "novel parse failed",
			"request_id", requestID,
			"duration_ms", time.Since(parseStart).Milliseconds(),
			"error", err.Error(),
		)
		return ConvertResponse{}, err
	}
	if err := validation.ValidateNovel(novel); err != nil {
		appErr := NewError(err.Code, err.Message)
		logger.WarnContext(ctx, "novel validation failed",
			"request_id", requestID,
			"chapter_count", len(novel.Chapters),
			"error_code", err.Code,
		)
		return ConvertResponse{}, appErr
	}
	logger.InfoContext(ctx, "novel parsed",
		"request_id", requestID,
		"chapter_count", len(novel.Chapters),
		"duration_ms", time.Since(parseStart).Milliseconds(),
	)

	generateStart := time.Now()
	logger.InfoContext(ctx, "screenplay generation started",
		"request_id", requestID,
		"chapter_count", len(novel.Chapters),
	)
	output, err := c.provider.GenerateScreenplay(ctx, ai.GenerateInput{Novel: novel})
	if err != nil {
		appErr := mapAIError(err)
		logger.WarnContext(ctx, "screenplay generation failed",
			"request_id", requestID,
			"duration_ms", time.Since(generateStart).Milliseconds(),
			"error", err.Error(),
		)
		return ConvertResponse{}, appErr
	}
	if output.RawYAML != "" {
		resp := ConvertResponse{
			ScreenplayYAML: output.RawYAML,
			ChapterCount:   len(novel.Chapters),
			Mode:           "api",
		}
		logger.InfoContext(ctx, "screenplay generation completed",
			"request_id", requestID,
			"duration_ms", time.Since(generateStart).Milliseconds(),
			"chapter_count", resp.ChapterCount,
			"mode", resp.Mode,
			"yaml_length", len(resp.ScreenplayYAML),
		)
		return resp, nil
	}
	screenplay := output.Screenplay

	resp := ConvertResponse{
		ScreenplayYAML: exporter.ExportYAML(screenplay),
		ChapterCount:   len(novel.Chapters),
		Mode:           screenplay.Mode,
	}
	logger.InfoContext(ctx, "screenplay generation completed",
		"request_id", requestID,
		"duration_ms", time.Since(generateStart).Milliseconds(),
		"chapter_count", resp.ChapterCount,
		"mode", resp.Mode,
		"yaml_length", len(resp.ScreenplayYAML),
	)
	return resp, nil
}

func mapAIError(err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return err
	}

	var yamlErr ai.YAMLValidationError
	if ai.AsYAMLValidationError(err, &yamlErr) {
		return NewError(ErrorCodeAIInvalidYAML, "AI 返回的 YAML 未通过结构校验，请重试。")
	}

	if errors.Is(err, ai.ErrDeepSeekAPIKeyRequired) ||
		errors.Is(err, ai.ErrDeepSeekBaseURLRequired) ||
		errors.Is(err, ai.ErrDeepSeekModelRequired) {
		return NewError(ErrorCodeAIProviderNotConfigured, "AI provider 配置不完整，请检查 DeepSeek API key、Base URL 和模型配置。")
	}

	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return NewError(ErrorCodeAITimeout, "AI 生成超时，请稍后重试或调大 DeepSeek 超时时间。")
	}

	return NewError(ErrorCodeAIGenerationFailed, "AI 生成失败，请稍后重试。")
}
