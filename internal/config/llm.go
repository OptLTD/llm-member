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
			Name: "openai", APIKey: apiKey,
			BaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		})
	}

	// Claude
	if apiKey := getEnv("CLAUDE_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "claude", APIKey: apiKey,
			BaseURL: getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"),
		})
	}

	// 通义千问
	if apiKey := getEnv("QWEN_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "qwen", APIKey: apiKey,
			BaseURL: getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/api/v1"),
		})
	}

	// 豆包
	if apiKey := getEnv("DOUBAO_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "doubao", APIKey: apiKey,
			BaseURL: getEnv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		})
	}

	// 智谱清言
	if apiKey := getEnv("BIGMODEL_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "bigmodel", APIKey: apiKey,
			BaseURL: getEnv("BIGMODEL_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		})
	}

	// Grok
	if apiKey := getEnv("GROK_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "grok", APIKey: apiKey,
			BaseURL: getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),
		})
	}

	// Gemini
	if apiKey := getEnv("GEMINI_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "gemini", APIKey: apiKey,
			BaseURL: getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
		})
	}

	// OpenRouter
	if apiKey := getEnv("OPENROUTER_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "openrouter", APIKey: apiKey,
			BaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		})
	}

	// SiliconFlow
	if apiKey := getEnv("SILICONFLOW_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "siliconflow", APIKey: apiKey,
			BaseURL: getEnv("SILICONFLOW_BASE_URL", "https://api.siliconflow.cn/v1"),
		})
	}

	// DeepSeek
	if apiKey := getEnv("DEEPSEEK_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "deepseek", APIKey: apiKey,
			BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		})
	}

	// OpenAI-Like
	if apiKey := getEnv("OPENAI_LIKE_API_KEY", ""); apiKey != "" {
		providers = append(providers, LLMProvider{
			Name: "openai-like", APIKey: apiKey,
			BaseURL: getEnv("OPENAI_LIKE_BASE_URL", ""),
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
