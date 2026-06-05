package app

import (
	"context"

	"github.com/qingketsing/novel2script/backend/internal/exporter"
	"github.com/qingketsing/novel2script/backend/internal/generator"
	"github.com/qingketsing/novel2script/backend/internal/parser"
	"github.com/qingketsing/novel2script/backend/internal/validation"
)

const (
	ErrCodeEmptyText            = validation.CodeEmptyText
	ErrCodeInsufficientChapters = validation.CodeInsufficientChapters
)

type MockDomainConverter struct{}

func NewMockDomainConverter() Converter {
	return MockDomainConverter{}
}

func (MockDomainConverter) Convert(_ context.Context, req ConvertRequest) (ConvertResponse, error) {
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

	screenplay := generator.GenerateMockScreenplay(novel)

	return ConvertResponse{
		ScreenplayYAML: exporter.ExportYAML(screenplay),
		ChapterCount:   len(novel.Chapters),
		Mode:           screenplay.Mode,
	}, nil
}
