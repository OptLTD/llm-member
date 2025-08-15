package service

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type I18nService struct {
	bundle *i18n.Bundle
}

func NewI18nService() *I18nService {
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 加载语言文件
	files := []string{"zh.toml", "en.toml"}
	for _, file := range files {
		filePath := filepath.Join("locales", file)
		_, err := bundle.LoadMessageFile(filePath)
		if err != nil {
			fmt.Printf("Failed to load language file %s: %v\n", filePath, err)
		}
	}

	return &I18nService{
		bundle: bundle,
	}
}

func (s *I18nService) GetLocalizer(lang string) *i18n.Localizer {
	if lang == "" {
		lang = "zh" // 默认中文
	}

	// 支持的语言映射
	langMap := map[string]string{
		"zh":    "zh",
		"zh-CN": "zh",
		"zh-TW": "zh",
		"en":    "en",
		"en-US": "en",
		"en-GB": "en",
	}

	if mappedLang, ok := langMap[lang]; ok {
		lang = mappedLang
	} else {
		lang = "zh" // 不支持的语言默认为中文
	}

	return i18n.NewLocalizer(s.bundle, lang)
}

func (s *I18nService) Translate(localizer *i18n.Localizer, messageID string, templateData map[string]interface{}) string {
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// 如果翻译失败，返回messageID作为fallback
		return messageID
	}
	return message
}

// 便捷方法：直接通过语言代码获取翻译
func (s *I18nService) T(lang, messageID string, templateData map[string]interface{}) string {
	localizer := s.GetLocalizer(lang)
	return s.Translate(localizer, messageID, templateData)
}

// 获取所有翻译的map，用于前端
func (s *I18nService) GetTranslations(lang string) map[string]string {
	localizer := s.GetLocalizer(lang)

	// 定义需要翻译的所有key
	keys := []string{
		"AppName", "Welcome", "SignIn", "SignUp", "SignOut", "Profile", "Payment", "Admin", "Home",
		"Back", "Submit", "Cancel", "Confirm", "Save", "Delete", "Edit", "Add", "Search", "Loading",
		"Success", "Error", "Warning", "Info",
		"SignInTitle", "SignInPageTitle", "Username", "Email", "Phone", "Password", "ConfirmPassword",
		"RememberMe", "ForgotPassword", "SignInButton", "NoAccount", "SignUpNow", "SignInWith",
		"GoogleLogin", "GitHubLogin", "SignUpTitle", "SignUpPageTitle", "SignUpButton", "AlreadyHaveAccount", "LoginNow",
		"PasswordStrength", "WeakPassword", "MediumPassword", "StrongPassword", "AgreeTerms",
		"TermsOfService", "And", "PrivacyPolicy",
		"ProfileTitle", "ProfilePageTitle", "PersonalInfo", "UsageStats", "APIKeys", "PlanManagement",
		"AccountSettings", "SecuritySettings",
		"PaymentTitle", "PaymentPageTitle", "SelectPlan", "SelectPaymentMethod", "OrderSummary",
		"SelectedPlan", "Amount", "TotalAmount", "PayNow", "CustomAmount", "MinAmount",
		"AdminTitle", "AdminPageTitle", "Dashboard", "ChatTest", "RequestLogs", "UserManagement",
		"IndexTitle", "IndexPageTitle", "Features", "Pricing", "GetStarted", "LearnMore",
		"InvalidCredentials", "NetworkError", "ServerError", "ValidationError", "Unauthorized", "NotFound",
		"SignInSuccess", "RegisterSuccess", "SaveSuccess", "UpdateSuccess", "DeleteSuccess",
	}

	translations := make(map[string]string)
	for _, key := range keys {
		translations[key] = s.Translate(localizer, key, nil)
	}

	return translations
}
