package consts

import "errors"

// Redis errors
var ErrRedisConnectionFailed = errors.New("failed to connect to redis")

// Order service errors
var (
	ErrOrderIDGenerationFailed = errors.New("生成订单ID失败")
	ErrOrderSaveFailed         = errors.New("保存订单失败")
	ErrOrderQueryFailed        = errors.New("查找订单失败")
	ErrOrderStatusUpdateFailed = errors.New("更新订单状态失败")
	ErrUserPlanUpdateFailed    = errors.New("更新用户套餐失败")

	ErrPaymentMethodNotEnabled     = errors.New("支付方式未启用")
	ErrPaymentOrderCreationFailed  = errors.New("创建支付订单失败")
	ErrPaymentWebhookMissingParams = errors.New("支付回调参数异常")
	ErrPaymentCallbackVerification = errors.New("支付回调验证失败")
)

// Token service errors
var (
	ErrDailyTokenLimitReached     = errors.New("已达到每日 Token 限制")
	ErrMonthlyTokenLimitReached   = errors.New("已达到每月 Token 限制")
	ErrDailyRequestLimitReached   = errors.New("已达到每日请求限制")
	ErrMonthlyRequestLimitReached = errors.New("已达到每月请求限制")
	ErrDailyProjectLimitReached   = errors.New("已达到每日项目限制")
	ErrMonthlyProjectLimitReached = errors.New("已达到每月项目限制")
)

// User service errors
var (
	ErrUserNotFound           = errors.New("用户不存在")
	ErrEmailAlreadyUsed       = errors.New("邮箱已被使用")
	ErrUserUpdateFailed       = errors.New("用户更新失败")
	ErrAPIKeyGenerationFailed = errors.New("API Key 生成失败")
	ErrAPIKeyUpdateFailed     = errors.New("API Key 更新失败")
	ErrUserStatusUpdateFailed = errors.New("用户状态更新失败")
)

// Setup service errors
var (
	ErrPeriodRegexError      = errors.New("周期正则表达式错误")
	ErrPlanPeriodFormatError = errors.New("套餐周期格式错误")
	ErrInvalidPeriodNumber   = errors.New("无效的周期数字")
	ErrUnsupportedPeriodUnit = errors.New("不支持的周期单位")
	ErrUsageRegexError       = errors.New("usage regex error")
	ErrPlanUsageFormatError  = errors.New("plan usage format error")
	ErrInvalidUsageFormat    = errors.New("invalid usage format")
	ErrInvalidUsageNumber    = errors.New("invalid usage number")
	ErrUnsupportedUsageType  = errors.New("unsupported usage type")

	ErrPlanQueryError       = errors.New("Query Plan error")
	ErrPlanNotEnabled       = errors.New("Plan is not enabled")
	ErrParsePlanLimitError  = errors.New("ParsePlanLimit error")
	ErrConfigCreationFailed = errors.New("failed to create config")
)

// Payment errors - 用户友好的支付错误（面向用户显示）
var (
	// 用户友好的错误信息 - 只告诉用户哪个环节出错
	ErrPaymentMethodNotSupported = errors.New("不支持的支付方式")
	ErrPaymentConfigurationError = errors.New("支付配置错误")
	ErrPaymentCreationError      = errors.New("创建支付失败")
	ErrPaymentQueryError         = errors.New("查询支付状态失败")
	ErrPaymentCancelError        = errors.New("取消支付失败")
	ErrPaymentRefundError        = errors.New("退款处理失败")
	ErrPaymentWebhookError       = errors.New("支付回调处理失败")
	ErrPaymentNetworkError       = errors.New("支付网络请求失败")

	// 详细的内部错误常量（用于日志记录和开发调试）
	ErrPaymentProviderNotConfigured       = errors.New("payment provider not configured")
	ErrPaymentClientCreationFailed        = errors.New("failed to create payment client")
	ErrPaymentCreationFailed              = errors.New("failed to create payment")
	ErrPaymentQueryFailed                 = errors.New("failed to query payment")
	ErrPaymentCloseFailed                 = errors.New("failed to close payment")
	ErrPaymentCloseError                  = errors.New("payment close error")
	ErrPaymentRefundFailed                = errors.New("failed to process refund")
	ErrPaymentWebhookNotImplemented       = errors.New("webhook verification not implemented")
	ErrPaymentConfigIncomplete            = errors.New("payment configuration incomplete")
	ErrPaymentSignGenerationFailed        = errors.New("failed to generate payment sign")
	ErrPaymentError                       = errors.New("payment error")
	ErrOrderCannotBeRefunded              = errors.New("order cannot be refunded")
	ErrRefundNotSuccessful                = errors.New("refund was not successful")
	ErrRequestDataMarshalFailed           = errors.New("failed to marshal request data")
	ErrRequestCreationFailed              = errors.New("failed to create request")
	ErrRequestSendFailed                  = errors.New("failed to send request")
	ErrResponseReadFailed                 = errors.New("failed to read response")
	ErrResponseParseFailed                = errors.New("failed to parse response")
	ErrAPIRequestFailed                   = errors.New("API request failed")
	ErrWebhookBodyReadFailed              = errors.New("failed to read webhook body")
	ErrInvalidHTTPMethod                  = errors.New("invalid HTTP method, expected POST")
	ErrWebhookSignatureVerificationFailed = errors.New("webhook signature verification failed")
	ErrWebhookEventParseFailed            = errors.New("failed to parse webhook event")
	ErrWebhookEventMissingID              = errors.New("webhook event missing required field: id")
	ErrWebhookEventMissingEventType       = errors.New("webhook event missing required field: eventType")
	ErrMissingWebhookSignatureHeader      = errors.New("missing webhook signature header")
	ErrWebhookSecretNotConfigured         = errors.New("webhook secret not configured")
	ErrSignatureMismatch                  = errors.New("signature mismatch")
)

// Auth service errors
var (
	ErrInvalidCredentials               = errors.New("用户名或密码错误")
	ErrUserAccountDisabled              = errors.New("用户账户已被禁用")
	ErrUserAccountNotVerified           = errors.New("未验证用户账户")
	ErrTokenStorageFailed               = errors.New("存储token失败")
	ErrUsernameAlreadyExists            = errors.New("用户名已存在")
	ErrEmailAlreadyExists               = errors.New("邮箱已存在")
	ErrPasswordEncryptionFailed         = errors.New("密码加密失败")
	ErrUserCreationFailed               = errors.New("创建用户失败")
	ErrInvalidVerificationCode          = errors.New("无效的验证码")
	ErrVerificationCodeExpired          = errors.New("验证码已过期")
	ErrEmailVerificationFailed          = errors.New("邮箱验证失败")
	ErrVerificationCodeDeletionFailed   = errors.New("删除验证码失败")
	ErrVerificationCodeGenerationFailed = errors.New("验证码生成失败")
	ErrInvalidAPIKey                    = errors.New("无效的 API Key")
	ErrInvalidToken                     = errors.New("无效的 token")
	ErrEmailNotRegistered               = errors.New("该邮箱未注册")
	ErrResetCodeGenerationFailed        = errors.New("重置码生成失败")
	ErrInvalidResetCode                 = errors.New("无效的重置码")
	ErrResetCodeExpired                 = errors.New("重置码已过期")
	ErrPasswordUpdateFailed             = errors.New("密码更新失败")
	ErrTokenSerializationFailed         = errors.New("序列化token信息失败")
	ErrTokenExpired                     = errors.New("token已过期")
	ErrRedisTokenStorageFailed          = errors.New("Redis存储token失败")
	ErrTempTokenStorageFailed           = errors.New("存储临时token失败")
)

// Relay service errors
var (
	ErrUnsupportedModel      = errors.New("unsupported model")
	ErrProviderNotConfigured = errors.New("provider not configured")
	ErrAPIError              = errors.New("API error")
)

// Mail service errors
var (
	ErrEmailTemplateLoadFailed = errors.New("failed to load email template")
	ErrSMTPAuthError           = errors.New("smpt auth error")
	ErrTemplateFileNotFound    = errors.New("template file not found")
	ErrEmailSendFailed         = errors.New("邮件发送失败")
	ErrEmailTemplateNotFound   = errors.New("邮件模板未找到")
	ErrEmailConfigInvalid      = errors.New("邮件配置无效")
)

// Setup service errors (additional)
var (
	ErrDatabaseMigrationFailed  = errors.New("数据库迁移失败")
	ErrCreateDefaultAdminFailed = errors.New("创建默认管理员失败")
	ErrInitDefaultConfigsFailed = errors.New("初始化默认配置失败")
	ErrCreateAdminUserFailed    = errors.New("创建管理员用户失败")
	ErrPlanNotFound             = errors.New("套餐不存在")
	ErrParsePlanLimitFailed     = errors.New("解析套餐限制失败")
	ErrTokenGenerationFailed    = errors.New("生成token失败")
)
