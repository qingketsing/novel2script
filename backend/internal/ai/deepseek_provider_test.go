package ai

import (
	"context"
	"errors"
	"testing"
)

func TestNewDeepSeekProviderRequiresAPIKey(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})

	if !errors.Is(err, ErrDeepSeekAPIKeyRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekAPIKeyRequired)
	}
}

func TestNewDeepSeekProviderRequiresBaseURL(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey: "test-api-key",
		Model:  "deepseek-v4",
	})

	if !errors.Is(err, ErrDeepSeekBaseURLRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekBaseURLRequired)
	}
}

func TestNewDeepSeekProviderRequiresModel(t *testing.T) {
	_, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
	})

	if !errors.Is(err, ErrDeepSeekModelRequired) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekModelRequired)
	}
}

func TestNewDeepSeekProviderReturnsProvider(t *testing.T) {
	provider, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})

	if err != nil {
		t.Fatalf("NewDeepSeekProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("provider is nil, want DeepSeek provider")
	}
}

func TestDeepSeekProviderGenerateScreenplayIsNotImplemented(t *testing.T) {
	provider, err := NewDeepSeekProvider(DeepSeekConfig{
		APIKey:  "test-api-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-v4",
	})
	if err != nil {
		t.Fatalf("NewDeepSeekProvider() error = %v", err)
	}

	_, err = provider.GenerateScreenplay(context.Background(), GenerateInput{})

	if !errors.Is(err, ErrDeepSeekProviderNotImplemented) {
		t.Fatalf("error = %v, want %v", err, ErrDeepSeekProviderNotImplemented)
	}
}
