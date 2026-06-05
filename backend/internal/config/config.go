package config

import "os"

type Config struct {
	Addr string
}

// Load 从环境变量读取服务配置，并提供本地开发默认值。
func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{Addr: addr}
}
