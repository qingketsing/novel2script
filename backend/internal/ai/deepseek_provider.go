package ai

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDeepSeekAPIKeyRequired  = errors.New("deepseek api key is required")
	ErrDeepSeekBaseURLRequired = errors.New("deepseek base url is required")
	ErrDeepSeekModelRequired   = errors.New("deepseek model is required")
)

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

type DeepSeekProvider struct {
	cfg           DeepSeekConfig
	yamlGenerator deepSeekYAMLGenerator
}

func NewDeepSeekProvider(cfg DeepSeekConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, ErrDeepSeekAPIKeyRequired
	}
	if cfg.BaseURL == "" {
		return nil, ErrDeepSeekBaseURLRequired
	}
	if cfg.Model == "" {
		return nil, ErrDeepSeekModelRequired
	}

	client, err := NewDeepSeekClient(DeepSeekClientConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		Timeout: cfg.Timeout,
	}, nil)
	if err != nil {
		return nil, err
	}

	return DeepSeekProvider{
		cfg:           cfg,
		yamlGenerator: client,
	}, nil
}

func (p DeepSeekProvider) GenerateScreenplay(ctx context.Context, input GenerateInput) (GenerateOutput, error) {
	prompt := BuildScreenplayPrompt(input.Novel)
	rawYAML, err := p.yamlGenerator.GenerateYAML(ctx, prompt)
	if err != nil {
		return GenerateOutput{}, err
	}
	if err := ValidateScreenplayYAML(rawYAML); err != nil {
		return GenerateOutput{}, err
	}

	return GenerateOutput{RawYAML: rawYAML}, nil
}

type deepSeekYAMLGenerator interface {
	GenerateYAML(ctx context.Context, prompt string) (string, error)
}
