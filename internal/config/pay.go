package config

type Wechat struct {
	AppID      string // 应用ID
	Token      string // 密钥/Token
	MchID      string // 商户ID
	SerialNo   string // 商户证书序列号
	APIv3Key   string // APIv3密钥
	PrivateKey string // 私钥内容
	NotifyURL  string // 回调地址
}

func GetWechatConfig() *Wechat {
	if getEnv("WECHAT_TOKEN", "") == "" {
		return nil
	}
	return &Wechat{
		AppID:      getEnv("WECHAT_APP_ID", ""),
		Token:      getEnv("WECHAT_TOKEN", ""),
		MchID:      getEnv("WECHAT_MCH_ID", ""),
		SerialNo:   getEnv("WECHAT_SERIAL_NO", ""),
		APIv3Key:   getEnv("WECHAT_APIV3_KEY", ""),
		PrivateKey: getEnv("WECHAT_PRIVATE_KEY", ""),
		NotifyURL:  getEnv("WECHAT_NOTIFY_URL", "http://localhost:8080/api/payment/wechat/notify"),
	}
}

type Creem struct {
	ApiKey string

	WhSecret string

	PlanBasicId string
	PlanExtraId string
	PlanUltraId string
	PlanSuperId string
}

func GetCreemConfig() *Creem {
	if getEnv("CREEM_API_KEY", "") == "" {
		return nil
	}
	return &Creem{
		ApiKey:   getEnv("CREEM_API_KEY", ""),
		WhSecret: getEnv("CREEM_WH_SECRET", ""),

		PlanBasicId: getEnv("PLAN_BASIC_ID", ""),
		PlanExtraId: getEnv("PLAN_EXTRA_ID", ""),
		PlanUltraId: getEnv("PLAN_ULTRA_ID", ""),
		PlanSuperId: getEnv("PLAN_SUPER_ID", ""),
	}
}

type Stripe struct {
	PublicKey string // 公钥
	SecretKey string // 私钥
	WhSecret  string // Webhook签名密钥

	PlanBasicId string
	PlanExtraId string
	PlanUltraId string
	PlanSuperId string
}

func GetStripeConfig() *Stripe {
	if getEnv("STRIPE_SECRET_KEY", "") == "" {
		return nil
	}
	return &Stripe{
		PublicKey: getEnv("STRIPE_PUBLIC_KEY", ""),
		SecretKey: getEnv("STRIPE_SECRET_KEY", ""),
		WhSecret:  getEnv("STRIPE_WH_SECRET", ""),

		PlanBasicId: getEnv("PLAN_BASIC_ID", ""),
		PlanExtraId: getEnv("PLAN_EXTRA_ID", ""),
		PlanUltraId: getEnv("PLAN_ULTRA_ID", ""),
		PlanSuperId: getEnv("PLAN_SUPER_ID", ""),
	}
}

type Alipay struct {
	AppID string // 应用ID
	Token string // 密钥/Token
}

func GetAlipayConfig() *Alipay {
	if getEnv("ALIPAY_APP_ID", "") == "" {
		return nil
	}
	return &Alipay{
		AppID: getEnv("ALIPAY_APP_ID", ""),
		Token: getEnv("ALIPAY_TOKEN", ""),
	}
}

type PayPal struct {
	AppID string // 应用ID
	Token string // 密钥/Token
}

func GetPayPalConfig() *PayPal {
	if getEnv("PAYPAL_APP_ID", "") == "" {
		return nil
	}
	return &PayPal{
		AppID: getEnv("PAYPAL_APP_ID", ""),
		Token: getEnv("PAYPAL_TOKEN", ""),
	}
}

// HasPayment 检查是否配置了指定的支付配置
func HasPayment(name string) bool {
	switch name {
	case "alipay":
		return GetAlipayConfig() != nil
	case "wechat":
		return GetWechatConfig() != nil
	case "stripe":
		return GetStripeConfig() != nil
	case "creem":
		return GetCreemConfig() != nil
	default:
		return false
	}
}
