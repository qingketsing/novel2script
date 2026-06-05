package app

import (
	"errors"
	"fmt"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/config"
)

var (
	ErrUnsupportedAIMode = errors.New("unsupported ai mode")
)

func NewProviderFromConfig(cfg config.Config) (ai.Provider, error) {
	switch cfg.AIMode {
	case "mock":
		return ai.NewMockProvider(), nil
	case "deepseek":
		return ai.NewDeepSeekProvider(ai.DeepSeekConfig{
			APIKey:  cfg.DeepSeekAPIKey,
			BaseURL: cfg.DeepSeekBaseURL,
			Model:   cfg.DeepSeekModel,
		})
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAIMode, cfg.AIMode)
	}
}
