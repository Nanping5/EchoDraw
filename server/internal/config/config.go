package config

import (
	"os"
	"strconv"
)

// Config 运行时配置。所有字段都从环境变量加载, 便于容器化部署。
type Config struct {
	Port       string
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
	CanvasW    int
	CanvasH    int
}

// Load 从环境变量加载配置, 缺失项使用默认值。
func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		LLMBaseURL: getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMAPIKey:  getEnv("LLM_API_KEY", ""),
		LLMModel:   getEnv("LLM_MODEL", "gpt-4o-mini"),
		CanvasW:    getEnvInt("CANVAS_W", 1200),
		CanvasH:    getEnvInt("CANVAS_H", 800),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
