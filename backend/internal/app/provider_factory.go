package app

import (
	"errors"
	"fmt"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/config"
)

var (
	ErrDeepSeekProviderNotImplemented = errors.New("deepseek provider is not implemented")
	ErrUnsupportedAIMode              = errors.New("unsupported ai mode")
)

func NewProviderFromConfig(cfg config.Config) (ai.Provider, error) {
	switch cfg.AIMode {
	case "mock":
		return ai.NewMockProvider(), nil
	case "deepseek":
		// 暂时返回错误，因为 DeepSeek 提供者尚未实现，先留出接口
		return nil, ErrDeepSeekProviderNotImplemented
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAIMode, cfg.AIMode)
	}
}
