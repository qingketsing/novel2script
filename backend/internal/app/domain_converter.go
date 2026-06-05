package app

import (
	"context"

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
		return ConvertResponse{}, err
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
