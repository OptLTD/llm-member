package relay

import (
	"os"
	"strings"
)

// LLMConfig 单一上游 LLM 配置（OpenAI-compatible）
type LLMConfig struct {
	Provider     string // 展示/日志用标签，如 openai、deepseek、openrouter
	APIKey       string
	BaseURL      string
	DefaultModel string
	Models       []string // 对外暴露的模型列表；空则仅含 DefaultModel
}

// ConfigFromEnv 读取 LLM_* 环境变量；未配置 API Key 时返回 nil
func ConfigFromEnv() *LLMConfig {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil
	}
	cfg := &LLMConfig{
		Provider:     envOr("LLM_PROVIDER", "openai"),
		APIKey:       apiKey,
		BaseURL:      envOr("LLM_BASE_URL", "https://api.openai.com/v1"),
		DefaultModel: os.Getenv("LLM_DEFAULT_MODEL"),
	}
	if raw := os.Getenv("LLM_MODELS"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				cfg.Models = append(cfg.Models, part)
			}
		}
	}
	if len(cfg.Models) == 0 && cfg.DefaultModel != "" {
		cfg.Models = []string{cfg.DefaultModel}
	}
	return cfg
}

// HasLLM 是否已配置 LLM 上游
func HasLLM() bool {
	return ConfigFromEnv() != nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
