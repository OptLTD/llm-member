package config

// LLMProvider 提供商配置
type LLMProvider struct {
	Name string // 提供商名称

	Default string // 默认模型
	BaseURL string // 基础 URL
	APIKey  string // API 密钥
}

// GetProviders 返回已配置的提供商列表
func GetProviders() []LLMProvider {
	var providers []LLMProvider

	// OpenAI
	if apiKey := getEnv("OPENAI_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "openai",
			BaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			APIKey:  apiKey,
		})
	}

	// Claude
	if apiKey := getEnv("CLAUDE_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "claude",
			BaseURL: getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"),
			APIKey:  apiKey,
		})
	}

	// 通义千问
	if apiKey := getEnv("QWEN_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "qwen",
			BaseURL: getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/api/v1"),
			APIKey:  apiKey,
		})
	}

	// 豆包
	if apiKey := getEnv("DOUBAO_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "doubao",
			BaseURL: getEnv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
			APIKey:  apiKey,
		})
	}

	// 智谱清言
	if apiKey := getEnv("BIGMODEL_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "bigmodel",
			BaseURL: getEnv("BIGMODEL_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
			APIKey:  apiKey,
		})
	}

	// Grok
	if apiKey := getEnv("GROK_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "grok",
			BaseURL: getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),
			APIKey:  apiKey,
		})
	}

	// Gemini
	if apiKey := getEnv("GEMINI_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "gemini",
			BaseURL: getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
			APIKey:  apiKey,
		})
	}

	// OpenRouter
	if apiKey := getEnv("OPENROUTER_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "openrouter",
			BaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			APIKey:  apiKey,
		})
	}

	// SiliconFlow
	if apiKey := getEnv("SILICONFLOW_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "siliconflow",
			BaseURL: getEnv("SILICONFLOW_BASE_URL", "https://api.siliconflow.cn/v1"),
			APIKey:  apiKey,
		})
	}

	// OpenAI-Like
	if apiKey := getEnv("OPENAI_LIKE_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name:    "openai-like",
			BaseURL: getEnv("OPENAI_LIKE_BASE_URL", ""),
			APIKey:  apiKey,
		})
	}

	return providers
}

// GetProvider 根据名称获取特定的提供商配置
func GetProvider(name string) *LLMProvider {
	providers := GetProviders()
	for _, provider := range providers {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}

// HasProvider 检查是否配置了指定的提供商
func HasProvider(name string) bool {
	return GetProvider(name) != nil
}
