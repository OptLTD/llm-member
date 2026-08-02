package relay

import (
	"database/sql/driver"
	"encoding/json"
)

// ApiUsage 用户用量（JSON 存储于 llm_user_quota.api_usage）
type ApiUsage struct {
	TotalTokens   uint64 `json:"totalTokens"`
	TotalRequests uint64 `json:"totalRequests"`
	TotalProjects uint64 `json:"totalProjects"`

	TodayTokens   uint64 `json:"todayTokens"`
	TodayRequests uint64 `json:"todayRequests"`
	TodayProjects uint64 `json:"todayProjects"`

	// DayStamp 用于日切重置 Today* 字段，格式 YYYY-MM-DD
	DayStamp string `json:"dayStamp,omitempty"`
}

// ApiLimit 用户配额（JSON 存储于 llm_user_quota.api_limit）
type ApiLimit struct {
	// 过期天数
	ExpireDays int `json:"expireDays,omitempty"`
	// 限制方法：tokens, requests, projects
	LimitMethod string `json:"limitMethod,omitempty"`

	DailyTokens   uint64 `json:"dailyTokens,omitempty"`
	MonthlyTokens uint64 `json:"monthlyTokens,omitempty"`

	DailyRequests   uint64 `json:"dailyRequests,omitempty"`
	MonthlyRequests uint64 `json:"monthlyRequests,omitempty"`

	DailyProjects   uint64 `json:"dailyProjects,omitempty"`
	MonthlyProjects uint64 `json:"monthlyProjects,omitempty"`
}

// Value implements driver.Valuer interface for ApiUsage
func (a ApiUsage) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface for ApiUsage
func (a *ApiUsage) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return nil
	}
}

// Value implements driver.Valuer interface for ApiLimit
func (a ApiLimit) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface for ApiLimit
func (a *ApiLimit) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return nil
	}
}
