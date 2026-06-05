package app

import (
	"errors"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/ai"
	"github.com/qingketsing/novel2script/backend/internal/config"
)

func TestNewProviderFromConfigUsesMockProvider(t *testing.T) {
	provider, err := NewProviderFromConfig(config.Config{AIMode: "mock"})

	if err != nil {
		t.Fatalf("NewProviderFromConfig() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider is nil, want mock provider")
	}
}

func TestNewProviderFromConfigUsesDeepSeekProvider(t *testing.T) {
	provider, err := NewProviderFromConfig(config.Config{
		AIMode:          "deepseek",
		DeepSeekAPIKey:  "test-api-key",
		DeepSeekBaseURL: "https://api.deepseek.com",
		DeepSeekModel:   "deepseek-v4",
	})

	if err != nil {
		t.Fatalf("NewProviderFromConfig() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider is nil, want DeepSeek provider")
	}
}

func TestNewProviderFromConfigRequiresDeepSeekAPIKey(t *testing.T) {
	_, err := NewProviderFromConfig(config.Config{
		AIMode:          "deepseek",
		DeepSeekBaseURL: "https://api.deepseek.com",
		DeepSeekModel:   "deepseek-v4",
	})

	if !errors.Is(err, ai.ErrDeepSeekAPIKeyRequired) {
		t.Fatalf("error = %v, want %v", err, ai.ErrDeepSeekAPIKeyRequired)
	}
}

func TestNewProviderFromConfigRejectsUnknownMode(t *testing.T) {
	_, err := NewProviderFromConfig(config.Config{AIMode: "other"})

	if !errors.Is(err, ErrUnsupportedAIMode) {
		t.Fatalf("error = %v, want %v", err, ErrUnsupportedAIMode)
	}
}
