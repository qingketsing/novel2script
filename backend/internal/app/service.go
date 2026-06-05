package app

import "context"

type ConvertRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	InputType string `json:"input_type"`
}

type ConvertResponse struct {
	ScreenplayYAML string `json:"screenplay_yaml"`
	ChapterCount   int    `json:"chapter_count"`
	Mode           string `json:"mode"`
}

// Converter 定义 HTTP API 层和领域管线之间的边界。
// 后续实现可以解析小说、生成 mock 剧本 YAML，或在不改变 HTTP 契约的前提下接入真实 AI provider。
type Converter interface {
	Convert(ctx context.Context, req ConvertRequest) (ConvertResponse, error)
}
