package config

import (
	"os"
)

type Config struct {
	Port        string
	GinMode     string
	DBPath      string

	// 大模型配置
	OpenAIAPIKey      string
	OpenAIBaseURL     string
	ClaudeAPIKey      string
	ClaudeBaseURL     string
	QwenAPIKey        string
	QwenBaseURL       string
	DoubaoAPIKey      string
	DoubaoBaseURL     string
	BigModelAPIKey    string
	BigModelBaseURL   string
	GrokAPIKey        string
	GrokBaseURL       string
	GeminiAPIKey      string
	GeminiBaseURL     string
	OpenRouterAPIKey  string
	OpenRouterBaseURL string
	SiliconFlowAPIKey string
	SiliconFlowBaseURL string
	OpenAILikeAPIKey  string
	OpenAILikeBaseURL string

	// 管理员配置
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),
		DBPath:     getEnv("DB_PATH", "./app.db"),

		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		ClaudeAPIKey:      getEnv("CLAUDE_API_KEY", ""),
		ClaudeBaseURL:     getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"),
		QwenAPIKey:        getEnv("QWEN_API_KEY", ""),
		QwenBaseURL:       getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/api/v1"),
		DoubaoAPIKey:      getEnv("DOUBAO_API_KEY", ""),
		DoubaoBaseURL:     getEnv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		BigModelAPIKey:    getEnv("BIGMODEL_API_KEY", ""),
		BigModelBaseURL:   getEnv("BIGMODEL_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		GrokAPIKey:        getEnv("GROK_API_KEY", ""),
		GrokBaseURL:       getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		GeminiBaseURL:     getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
		OpenRouterAPIKey:  getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterBaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		SiliconFlowAPIKey: getEnv("SILICONFLOW_API_KEY", ""),
		SiliconFlowBaseURL: getEnv("SILICONFLOW_BASE_URL", "https://api.siliconflow.cn/v1"),
		OpenAILikeAPIKey:  getEnv("OPENAI_LIKE_API_KEY", ""),
		OpenAILikeBaseURL: getEnv("OPENAI_LIKE_BASE_URL", ""),

		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}