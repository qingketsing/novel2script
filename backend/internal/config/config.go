package config

import "os"

type Config struct {
	Addr            string
	AIMode          string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string
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

	return Config{
		Addr:            addr,
		AIMode:          aiMode,
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekBaseURL: deepSeekBaseURL,
		DeepSeekModel:   deepSeekModel,
	}
}
