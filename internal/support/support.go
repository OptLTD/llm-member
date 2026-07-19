package support

import (
	"crypto/rand"
	"crypto/sha256"
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

// GenerateAPIKey 生成随机 API Key（明文，仅返回一次）
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

// HashAPIKey 对 API Key 做 SHA256，用于入库存储
func HashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}

// IsHashedAPIKey 判断是否已是 hash 形态（非 sk- 前缀）
func IsHashedAPIKey(stored string) bool {
	return stored != "" && !strings.HasPrefix(stored, "sk-")
}
