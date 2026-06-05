package app

import (
	"errors"
	"testing"

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

func TestNewProviderFromConfigRejectsDeepSeekUntilImplemented(t *testing.T) {
	_, err := NewProviderFromConfig(config.Config{AIMode: "deepseek"})

	if !errors.Is(err, ErrDeepSeekProviderNotImplemented) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekProviderNotImplemented)
	}
}

func TestNewProviderFromConfigRejectsUnknownMode(t *testing.T) {
	_, err := NewProviderFromConfig(config.Config{AIMode: "other"})

	if !errors.Is(err, ErrUnsupportedAIMode) {
		t.Fatalf("error = %v, want %v", err, ErrUnsupportedAIMode)
	}
}
