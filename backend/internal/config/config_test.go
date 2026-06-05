package config

import "testing"

func TestLoadUsesDefaultAISettings(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("AI_MODE", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("DEEPSEEK_BASE_URL", "")
	t.Setenv("DEEPSEEK_MODEL", "")

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
}

func TestLoadUsesAISettingsFromEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("AI_MODE", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	t.Setenv("DEEPSEEK_BASE_URL", "https://example.test")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4")

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
}
