package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaultAISettings(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("AI_MODE", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("DEEPSEEK_BASE_URL", "")
	t.Setenv("DEEPSEEK_MODEL", "")
	t.Setenv("DEEPSEEK_TIMEOUT_SECONDS", "")
	t.Setenv("AI_FALLBACK_TO_MOCK", "")

	cfg := Load()

	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.AIMode != "mock" {
		t.Fatalf("AIMode = %q, want %q", cfg.AIMode, "mock")
	}
	if cfg.DeepSeekAPIKey != "" {
		t.Fatalf("DeepSeekAPIKey = %q, want empty", cfg.DeepSeekAPIKey)
	}
	if cfg.DeepSeekBaseURL != "https://api.deepseek.com" {
		t.Fatalf("DeepSeekBaseURL = %q, want %q", cfg.DeepSeekBaseURL, "https://api.deepseek.com")
	}
	if cfg.DeepSeekModel != "deepseek-v4" {
		t.Fatalf("DeepSeekModel = %q, want %q", cfg.DeepSeekModel, "deepseek-v4")
	}
	if cfg.DeepSeekTimeout != 30*time.Second {
		t.Fatalf("DeepSeekTimeout = %v, want %v", cfg.DeepSeekTimeout, 30*time.Second)
	}
	if cfg.AIFallbackToMock {
		t.Fatal("AIFallbackToMock = true, want false")
	}
}

func TestLoadUsesAISettingsFromEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("AI_MODE", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	t.Setenv("DEEPSEEK_BASE_URL", "https://example.test")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4")
	t.Setenv("DEEPSEEK_TIMEOUT_SECONDS", "5")
	t.Setenv("AI_FALLBACK_TO_MOCK", "true")

	cfg := Load()

	if cfg.Addr != ":9090" {
		t.Fatalf("Addr = %q, want %q", cfg.Addr, ":9090")
	}
	if cfg.AIMode != "deepseek" {
		t.Fatalf("AIMode = %q, want %q", cfg.AIMode, "deepseek")
	}
	if cfg.DeepSeekAPIKey != "test-api-key" {
		t.Fatalf("DeepSeekAPIKey = %q, want %q", cfg.DeepSeekAPIKey, "test-api-key")
	}
	if cfg.DeepSeekBaseURL != "https://example.test" {
		t.Fatalf("DeepSeekBaseURL = %q, want %q", cfg.DeepSeekBaseURL, "https://example.test")
	}
	if cfg.DeepSeekModel != "deepseek-v4" {
		t.Fatalf("DeepSeekModel = %q, want %q", cfg.DeepSeekModel, "deepseek-v4")
	}
	if cfg.DeepSeekTimeout != 5*time.Second {
		t.Fatalf("DeepSeekTimeout = %v, want %v", cfg.DeepSeekTimeout, 5*time.Second)
	}
	if !cfg.AIFallbackToMock {
		t.Fatal("AIFallbackToMock = false, want true")
	}
}

func TestLoadUsesDefaultDeepSeekTimeoutForInvalidEnv(t *testing.T) {
	t.Setenv("DEEPSEEK_TIMEOUT_SECONDS", "invalid")

	cfg := Load()

	if cfg.DeepSeekTimeout != 30*time.Second {
		t.Fatalf("DeepSeekTimeout = %v, want %v", cfg.DeepSeekTimeout, 30*time.Second)
	}
}

func TestLoadUsesFalseForInvalidFallbackEnv(t *testing.T) {
	t.Setenv("AI_FALLBACK_TO_MOCK", "not-bool")

	cfg := Load()

	if cfg.AIFallbackToMock {
		t.Fatal("AIFallbackToMock = true, want false")
	}
}
