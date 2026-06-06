package app

import (
	"context"
	"errors"
	"net"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/exporter"
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
	if err := validation.ValidateInput(req.Content); err != nil {
		return ConvertResponse{}, NewError(err.Code, err.Message)
	}

	novel, err := parser.ParseNovel(req.Title, req.Content)
	if err != nil {
		return ConvertResponse{}, err
	}
	if err := validation.ValidateNovel(novel); err != nil {
		return ConvertResponse{}, NewError(err.Code, err.Message)
	}

	output, err := c.provider.GenerateScreenplay(ctx, ai.GenerateInput{Novel: novel})
	if err != nil {
		return ConvertResponse{}, mapAIError(err)
	}
	if output.RawYAML != "" {
		return ConvertResponse{
			ScreenplayYAML: output.RawYAML,
			ChapterCount:   len(novel.Chapters),
			Mode:           "api",
		}, nil
	}
	screenplay := output.Screenplay

	return ConvertResponse{
		ScreenplayYAML: exporter.ExportYAML(screenplay),
		ChapterCount:   len(novel.Chapters),
		Mode:           screenplay.Mode,
	}, nil
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
