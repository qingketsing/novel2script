package ai

import (
	"context"

	"github.com/qingketsing/novel2script/backend/internal/domain"
	"github.com/qingketsing/novel2script/backend/internal/generator"
)

type GenerateInput struct {
	Novel domain.Novel
}

type GenerateOutput struct {
	Screenplay domain.Screenplay
	RawYAML    string
}

type Provider interface {
	GenerateScreenplay(ctx context.Context, input GenerateInput) (GenerateOutput, error)
}

type MockProvider struct{}

func NewMockProvider() Provider {
	return MockProvider{}
}

func (MockProvider) GenerateScreenplay(_ context.Context, input GenerateInput) (GenerateOutput, error) {
	return GenerateOutput{
		Screenplay: generator.GenerateMockScreenplay(input.Novel),
	}, nil
}
