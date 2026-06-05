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

// Converter is the boundary between the HTTP API layer and the domain pipeline.
// Implementations may parse novel input, generate mock screenplay YAML, or later
// delegate to a real AI provider without changing the HTTP contract.
type Converter interface {
	Convert(ctx context.Context, req ConvertRequest) (ConvertResponse, error)
}
