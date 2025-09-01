package consts

import (
	"errors"
	"fmt"
)

// PaymentProvider 支付提供商枚举
type PaymentProvider string

const (
	ProviderStripe PaymentProvider = "stripe"
	ProviderAlipay PaymentProvider = "alipay"
	ProviderWechat PaymentProvider = "wechat"
	ProviderPaypal PaymentProvider = "paypal"
	ProviderCreem  PaymentProvider = "creem"
)

// PaymentErrorType 支付错误类型枚举
type PaymentErrorType string

const (
	ErrorTypeConfig   PaymentErrorType = "config"
	ErrorTypeCreation PaymentErrorType = "creation"
	ErrorTypeQuery    PaymentErrorType = "query"
	ErrorTypeRefund   PaymentErrorType = "refund"
	ErrorTypeWebhook  PaymentErrorType = "webhook"
	ErrorTypeNetwork  PaymentErrorType = "network"
	ErrorTypeCancel   PaymentErrorType = "cancel"
)

// 内部错误到用户友好错误的映射
var internalToUserErrorMap = map[error]error{
	// 配置相关错误
	ErrPaymentProviderNotConfigured: ErrPaymentConfigurationError,
	ErrPaymentConfigIncomplete:      ErrPaymentConfigurationError,
	ErrPaymentClientCreationFailed:  ErrPaymentConfigurationError,
	ErrWebhookSecretNotConfigured:   ErrPaymentConfigurationError,

	// 创建支付相关错误
	ErrPaymentCreationFailed:       ErrPaymentCreationError,
	ErrPaymentSignGenerationFailed: ErrPaymentCreationError,
	ErrRequestDataMarshalFailed:    ErrPaymentCreationError,

	// 查询相关错误
	ErrPaymentQueryFailed: ErrPaymentQueryError,
	ErrPaymentError:       ErrPaymentQueryError,

	// 取消/关闭相关错误
	ErrPaymentCloseFailed: ErrPaymentCancelError,
	ErrPaymentCloseError:  ErrPaymentCancelError,

	// 退款相关错误
	ErrPaymentRefundFailed:   ErrPaymentRefundError,
	ErrOrderCannotBeRefunded: ErrPaymentRefundError,
	ErrRefundNotSuccessful:   ErrPaymentRefundError,

	// 网络请求相关错误
	ErrRequestCreationFailed: ErrPaymentNetworkError,
	ErrRequestSendFailed:     ErrPaymentNetworkError,
	ErrResponseReadFailed:    ErrPaymentNetworkError,
	ErrResponseParseFailed:   ErrPaymentNetworkError,
	ErrAPIRequestFailed:      ErrPaymentNetworkError,

	// Webhook 相关错误
	ErrPaymentWebhookNotImplemented:       ErrPaymentWebhookError,
	ErrWebhookBodyReadFailed:              ErrPaymentWebhookError,
	ErrInvalidHTTPMethod:                  ErrPaymentWebhookError,
	ErrWebhookSignatureVerificationFailed: ErrPaymentWebhookError,
	ErrWebhookEventParseFailed:            ErrPaymentWebhookError,
	ErrWebhookEventMissingID:              ErrPaymentWebhookError,
	ErrWebhookEventMissingEventType:       ErrPaymentWebhookError,
	ErrMissingWebhookSignatureHeader:      ErrPaymentWebhookError,
	ErrSignatureMismatch:                  ErrPaymentWebhookError,
}

// CreatePaymentError 创建带有提供商和错误类型信息的支付错误（用于日志记录）
func CreatePaymentError(provider PaymentProvider, errorType PaymentErrorType, message string) error {
	return fmt.Errorf("[%s:%s] %s", provider, errorType, message)
}

// WrapPaymentError 包装现有错误，添加支付上下文信息（用于日志记录）
func WrapPaymentError(provider PaymentProvider, errorType PaymentErrorType, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("[%s:%s] %w", provider, errorType, err)
}

// GetUserFriendlyError 将内部错误转换为用户友好的错误信息
func GetUserFriendlyError(err error) error {
	if err == nil {
		return nil
	}

	// 检查是否有直接映射
	if userErr, exists := internalToUserErrorMap[err]; exists {
		return userErr
	}

	// 检查是否是包装的错误
	var unwrappedErr error
	if errors.As(err, &unwrappedErr) {
		if userErr, exists := internalToUserErrorMap[unwrappedErr]; exists {
			return userErr
		}
	}

	// 默认返回通用的支付错误
	return errors.New("支付处理失败")
}

// LogDetailedError 记录详细错误信息（供管理员查看日志）
func LogDetailedError(provider PaymentProvider, errorType PaymentErrorType, err error, context string) {
	// 这里可以集成具体的日志库，比如 logrus, zap 等
	// 暂时使用简单的格式化输出
	fmt.Printf("[PAYMENT_ERROR] Provider: %s, Type: %s, Context: %s, Error: %v\n",
		provider, errorType, context, err)
}

// IsPaymentError 检查错误是否为支付相关错误
func IsPaymentError(err error) bool {
	if err == nil {
		return false
	}
	// 简单的字符串匹配检查
	errorStr := err.Error()
	return len(errorStr) > 0 && errorStr[0] == '[' &&
		(contains(errorStr, "stripe:") || contains(errorStr, "alipay:") ||
			contains(errorStr, "wechat:") || contains(errorStr, "paypal:") ||
			contains(errorStr, "creem:"))
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 || findSubstring(s, substr) >= 0)
}

// findSubstring 查找子字符串位置
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
