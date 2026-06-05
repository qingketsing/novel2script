package app

import "context"

type PlaceholderConverter struct{}

func NewPlaceholderConverter() Converter {
	return PlaceholderConverter{}
}

func (PlaceholderConverter) Convert(_ context.Context, _ ConvertRequest) (ConvertResponse, error) {
	return ConvertResponse{
		ScreenplayYAML: "schema_version: \"1.0\"\nmetadata:\n  generated_by:\n    provider: \"deepseek-v4\"\n    mode: \"mock\"\n",
		ChapterCount:   0,
		Mode:           "mock",
	}, nil
}
