package app

import "context"

type PlaceholderConverter struct{}

// NewPlaceholderConverter 返回占位转换器，用于在领域管线完成前稳定 API 契约。
func NewPlaceholderConverter() Converter {
	return PlaceholderConverter{}
}

// Convert 返回固定 mock YAML，避免当前骨架阶段依赖真实解析和 AI 生成。
func (PlaceholderConverter) Convert(_ context.Context, _ ConvertRequest) (ConvertResponse, error) {
	return ConvertResponse{
		ScreenplayYAML: "schema_version: \"1.0\"\nmetadata:\n  generated_by:\n    provider: \"deepseek-v4\"\n    mode: \"mock\"\n",
		ChapterCount:   0,
		Mode:           "mock",
	}, nil
}
