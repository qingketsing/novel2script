package ai

import (
	"context"
	"errors"
)

var (
	ErrDeepSeekAPIKeyRequired         = errors.New("deepseek api key is required")
	ErrDeepSeekBaseURLRequired        = errors.New("deepseek base url is required")
	ErrDeepSeekModelRequired          = errors.New("deepseek model is required")
	ErrDeepSeekProviderNotImplemented = errors.New("deepseek provider is not implemented")
)

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type DeepSeekProvider struct {
	cfg DeepSeekConfig
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
	return DeepSeekProvider{cfg: cfg}, nil
}

func (p DeepSeekProvider) GenerateScreenplay(context.Context, GenerateInput) (GenerateOutput, error) {
	return GenerateOutput{}, ErrDeepSeekProviderNotImplemented
}
