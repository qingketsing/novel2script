package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr             string
	AIMode           string
	DeepSeekAPIKey   string
	DeepSeekBaseURL  string
	DeepSeekModel    string
	DeepSeekTimeout  time.Duration
	AIFallbackToMock bool
}

// Load 从环境变量读取服务配置，并提供本地开发默认值。
func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	aiMode := os.Getenv("AI_MODE")
	if aiMode == "" {
		aiMode = "mock"
	}

	deepSeekBaseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if deepSeekBaseURL == "" {
		deepSeekBaseURL = "https://api.deepseek.com"
	}

	deepSeekModel := os.Getenv("DEEPSEEK_MODEL")
	if deepSeekModel == "" {
		deepSeekModel = "deepseek-v4"
	}

	deepSeekTimeout := deepSeekTimeoutFromEnv()
	aiFallbackToMock := boolFromEnv("AI_FALLBACK_TO_MOCK")

	return Config{
		Addr:             addr,
		AIMode:           aiMode,
		DeepSeekAPIKey:   os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBaseURL:  deepSeekBaseURL,
		DeepSeekModel:    deepSeekModel,
		DeepSeekTimeout:  deepSeekTimeout,
		AIFallbackToMock: aiFallbackToMock,
	}
}

func deepSeekTimeoutFromEnv() time.Duration {
	const defaultTimeout = 30 * time.Second

	value := os.Getenv("DEEPSEEK_TIMEOUT_SECONDS")
	if value == "" {
		return defaultTimeout
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(seconds) * time.Second
}

func boolFromEnv(name string) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	return err == nil && value
}
