package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	AppPort string
	AppName string
	AppDesc string
	AppHost string
	GinMode string
	DBPath  string
	LogFile string
}

// Provider 提供商配置
type Provider struct {
	Name string // 提供商名称

	Default string // 默认模型
	BaseURL string // 基础 URL
	APIKey  string // API 密钥
}

// PaymentProvider 支付提供商配置
type PaymentProvider struct {
	Name    string // 支付提供商名称
	AppID   string // 应用ID
	Token   string // 密钥/Token
	Enabled bool   // 是否启用

	// 微信支付v3专用字段
	MchID      string // 商户ID
	SerialNo   string // 商户证书序列号
	APIv3Key   string // APIv3密钥
	PrivateKey string // 私钥内容
	NotifyURL  string // 回调地址
}

// MailConfig 邮件配置
type MailConfig struct {
	FromName string // 发件人名称
	HostAddr string // SMTP服务器地址
	FromAddr string // 发件人邮箱
	Password string // 邮箱密码或授权码
	UserName string // 用户名
}

func Load() *Config {
	cfg := &Config{
		GinMode: getEnv("GIN_MODE", "test"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "Demo"),
		AppDesc: getEnv("APP_DESC", "Desc"),
		AppHost: getEnv("APP_HOST", "localhost"),
		DBPath:  getEnv("DB_PATH", "storage/app.db"),
		LogFile: getEnv("LOG_FILE", "storage/app.log"),
	}
	return cfg
}

// GetLogger 配置日志系统
func GetLogger(cfg *Config) io.Writer {
	// 创建必要的目录（基于日志文件和数据库文件路径）
	logDir := filepath.Dir(cfg.LogFile)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatal("Failed to create log directory:", err)
		}
	}

	// 配置日志文件
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	// 注意：这里不能defer关闭，因为日志文件需要在整个程序运行期间保持打开

	// 设置日志输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// 将multiWriter保存到配置中，供Gin使用
	return multiWriter
}

// GetProviders 返回已配置的提供商列表
func GetProviders() []Provider {
	var providers []Provider

	// OpenAI
	if apiKey := getEnv("OPENAI_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "openai",
			BaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			APIKey:  apiKey,
		})
	}

	// Claude
	if apiKey := getEnv("CLAUDE_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "claude",
			BaseURL: getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"),
			APIKey:  apiKey,
		})
	}

	// 通义千问
	if apiKey := getEnv("QWEN_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "qwen",
			BaseURL: getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/api/v1"),
			APIKey:  apiKey,
		})
	}

	// 豆包
	if apiKey := getEnv("DOUBAO_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "doubao",
			BaseURL: getEnv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
			APIKey:  apiKey,
		})
	}

	// 智谱清言
	if apiKey := getEnv("BIGMODEL_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "bigmodel",
			BaseURL: getEnv("BIGMODEL_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
			APIKey:  apiKey,
		})
	}

	// Grok
	if apiKey := getEnv("GROK_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "grok",
			BaseURL: getEnv("GROK_BASE_URL", "https://api.x.ai/v1"),
			APIKey:  apiKey,
		})
	}

	// Gemini
	if apiKey := getEnv("GEMINI_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "gemini",
			BaseURL: getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
			APIKey:  apiKey,
		})
	}

	// OpenRouter
	if apiKey := getEnv("OPENROUTER_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "openrouter",
			BaseURL: getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
			APIKey:  apiKey,
		})
	}

	// SiliconFlow
	if apiKey := getEnv("SILICONFLOW_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "siliconflow",
			BaseURL: getEnv("SILICONFLOW_BASE_URL", "https://api.siliconflow.cn/v1"),
			APIKey:  apiKey,
		})
	}

	// OpenAI-Like
	if apiKey := getEnv("OPENAI_LIKE_API_KEY", ""); apiKey != "" {
		providers = append(providers, Provider{
			Name:    "openai-like",
			BaseURL: getEnv("OPENAI_LIKE_BASE_URL", ""),
			APIKey:  apiKey,
		})
	}

	return providers
}

// GetProvider 根据名称获取特定的提供商配置
func GetProvider(name string) *Provider {
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

// GetPaymentProviders 返回已配置的支付提供商列表
func GetPaymentProviders() []PaymentProvider {
	var providers []PaymentProvider

	// 支付宝
	if mode := getEnv("GIN_MODE", ""); mode != "release" {
		providers = append(providers, PaymentProvider{
			Name: "mock", AppID: mode, Enabled: true,
		})
	}

	// 支付宝
	if appID := getEnv("ALIPAY_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name:    "alipay",
			AppID:   appID,
			Token:   getEnv("ALIPAY_TOKEN", ""),
			Enabled: getEnv("ALIPAY_ENABLED", "false") == "true",
		})
	}

	// 微信支付
	if appID := getEnv("WECHAT_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name:       "wechat",
			AppID:      appID,
			Token:      getEnv("WECHAT_TOKEN", ""),
			Enabled:    getEnv("WECHAT_ENABLED", "false") == "true",
			MchID:      getEnv("WECHAT_MCH_ID", ""),
			SerialNo:   getEnv("WECHAT_SERIAL_NO", ""),
			APIv3Key:   getEnv("WECHAT_APIV3_KEY", ""),
			PrivateKey: getEnv("WECHAT_PRIVATE_KEY", ""),
			NotifyURL:  getEnv("WECHAT_NOTIFY_URL", "http://localhost:8080/api/payment/wechat/notify"),
		})
	}

	// 银联支付
	if appID := getEnv("UNION_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name:    "union",
			AppID:   appID,
			Token:   getEnv("UNION_TOKEN", ""),
			Enabled: getEnv("UNION_ENABLED", "false") == "true",
		})
	}

	// Stripe
	if appID := getEnv("STRIPE_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name:    "stripe",
			AppID:   appID,
			Token:   getEnv("STRIPE_TOKEN", ""),
			Enabled: getEnv("STRIPE_ENABLED", "false") == "true",
		})
	}

	return providers
}

// GetPaymentProvider 根据名称获取特定的支付提供商配置
func GetPaymentProvider(name string) *PaymentProvider {
	providers := GetPaymentProviders()
	for _, provider := range providers {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}

// HasPaymentProvider 检查是否配置了指定的支付提供商
func HasPaymentProvider(name string) bool {
	provider := GetPaymentProvider(name)
	return provider != nil && provider.Enabled
}

// GetMailConfig 获取邮件配置
func GetMailConfig() *MailConfig {
	return &MailConfig{
		FromName: getEnv("MAIL_FROM_NAME", ""),
		HostAddr: getEnv("MAIL_HOST_ADDR", ""),
		FromAddr: getEnv("MAIL_FROM_ADDR", ""),
		Password: getEnv("MAIL_PASS_WORD", ""),
		UserName: getEnv("MAIL_USER_NAME", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
