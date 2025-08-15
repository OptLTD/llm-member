package support

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// 辅助函数：从map中获取字符串值
func GetString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// 辅助函数：从map中获取字符串切片值
func GetStringSlice(m map[string]any, key string) []string {
	if val, ok := m[key].([]any); ok {
		var result []string
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// 辅助函数：从map中获取浮点数值
func GetFloat64(m map[string]any, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	if val, ok := m[key].(int); ok {
		return float64(val)
	}
	return 0.0
}

// 辅助函数：从map中获取布尔值
func GetBool(m map[string]any, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// IsPhoneNumber 检查是否为手机号格式
func IsPhoneNumber(input string) bool {
	// 简单的手机号格式检查（中国手机号）
	if len(input) != 11 {
		return false
	}
	for _, char := range input {
		if char < '0' || char > '9' {
			return false
		}
	}
	return strings.HasPrefix(input, "1")
}

// isEmail 检查是否为邮箱格式
func IsEmail(input string) bool {
	return strings.Contains(input, "@") && strings.Contains(input, ".")
}

// generateAPIKey 生成随机 API Key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}
