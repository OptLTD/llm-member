package config

// PaymentProvider 支付提供商配置
type PaymentProvider struct {
	Name  string // 支付提供商名称
	AppID string // 应用ID
	Token string // 密钥/Token

	WHSEC string // Webhook签名密钥
}

type WechatProvider struct {
	AppID      string // 应用ID
	Token      string // 密钥/Token
	MchID      string // 商户ID
	SerialNo   string // 商户证书序列号
	APIv3Key   string // APIv3密钥
	PrivateKey string // 私钥内容
	NotifyURL  string // 回调地址
}

func GetWechatProvider() *WechatProvider {
	return &WechatProvider{
		AppID:      getEnv("WECHAT_APP_ID", ""),
		Token:      getEnv("WECHAT_TOKEN", ""),
		MchID:      getEnv("WECHAT_MCH_ID", ""),
		SerialNo:   getEnv("WECHAT_SERIAL_NO", ""),
		APIv3Key:   getEnv("WECHAT_APIV3_KEY", ""),
		PrivateKey: getEnv("WECHAT_PRIVATE_KEY", ""),
		NotifyURL:  getEnv("WECHAT_NOTIFY_URL", "http://localhost:8080/api/payment/wechat/notify"),
	}
}

// GetPaymentProviders 返回已配置的支付提供商列表
func GetPaymentProviders() []PaymentProvider {
	var providers []PaymentProvider

	// 支付宝
	if mode := getEnv("APP_MODE", ""); mode != "release" {
		providers = append(providers, PaymentProvider{
			Name: "mock", AppID: mode,
		})
	}

	// 支付宝
	if appID := getEnv("ALIPAY_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name: "alipay", AppID: appID,
			Token: getEnv("ALIPAY_TOKEN", ""),
		})
	}

	// 微信支付
	if appID := getEnv("WECHAT_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name: "wechat", AppID: appID,
			Token: getEnv("WECHAT_TOKEN", ""),
		})
	}

	// 银联支付
	if appID := getEnv("UNION_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name: "union", AppID: appID,
			Token: getEnv("UNION_TOKEN", ""),
		})
	}

	// Stripe
	if appID := getEnv("STRIPE_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name: "stripe", AppID: appID,
			Token: getEnv("STRIPE_TOKEN", ""),
		})
	}

	// Creem
	if appID := getEnv("CREEM_APP_ID", ""); appID != "" {
		providers = append(providers, PaymentProvider{
			Name: "creem", AppID: appID,
			Token: getEnv("CREEM_TOKEN", ""),
			WHSEC: getEnv("CREEM_WHSEC", ""),
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
	return provider != nil
}
