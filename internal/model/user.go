package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// UserModel 用户模型
type UserModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Email    string `json:"email" gorm:"type:varchar(256);index;not null"`
	Username string `json:"username" gorm:"type:varchar(128);index;not null"`
	Password string `json:"password" gorm:"type:varchar(64);not null"` // 不在 JSON 中返回密码
	APIKey   string `json:"api_key" gorm:"type:varchar(68);index;not null"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
	Verified bool   `json:"verified" gorm:"default:false"`
	UserRole Role   `json:"user_role" gorm:"default:user"`

	// 当前套餐、过期时间
	ExpireAt *time.Time `json:"expire_at" gorm:"default:null"`     // 过期时间
	UserPlan PayPlan    `json:"user_plan" gorm:"type:varchar(20)"` // 用户套餐
	ApiUsage *ApiUsage  `json:"api_usage" gorm:"type:text;serializer:json"`
	ApiLimit *ApiLimit  `json:"api_limit" gorm:"type:text;serializer:json"`

	// 使用统计
	// TotalTokens   int64 `json:"total_tokens" gorm:"default:0"`
	// TotalRequests int64 `json:"total_requests" gorm:"default:0"`

	// // 限制设置
	// DailyLimit   int64 `json:"daily_limit" gorm:"default:1000"`    // 每日请求限制
	// MonthlyLimit int64 `json:"monthly_limit" gorm:"default:10000"` // 每月请求限制

	gorm.Model
}

func (m UserModel) TableName() string {
	return "llm_user"
}

// CreateUserReq 创建用户请求
type CreateUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"required,min=6"`
}

// SignUpRequest 用户端注册请求
type SignUpRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"required,min=6"`
}

// SignUpResponse 注册响应
type SignUpResponse struct {
	User    *UserModel `json:"user"`
	APIKey  string     `json:"api_key"`
	Message string     `json:"message"`
}

// SignInRequest 登录请求
type SignInRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SignInResponse 登录响应
type SignInResponse struct {
	Token   string     `json:"token"`
	User    *UserModel `json:"user"`
	Message string     `json:"message"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	UserId uint64 `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`

	IsActive *bool    `json:"is_active,omitempty"`
	UserPlan *PayPlan `json:"user_plan,omitempty"`

	DailyLimit   *int64 `json:"daily_limit,omitempty"`
	MonthlyLimit *int64 `json:"monthly_limit,omitempty"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	Email    string `json:"email"`
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	Verified bool   `json:"verified"`
	UserRole Role   `json:"user_role"`

	UserPlan PayPlan    `json:"user_plan"`
	ExpireAt *time.Time `json:"expire_at"`
	ApiUsage *ApiUsage  `json:"api_usage"`
	ApiLimit *ApiLimit  `json:"api_limit"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToUserResponse 将UserModel转换为UserResponse
func (u *UserModel) ToUserResponse() *UserResponse {
	return &UserResponse{
		Email:     u.Email,
		APIKey:    u.APIKey,
		Username:  u.Username,
		IsActive:  u.IsActive,
		Verified:  u.Verified,
		UserRole:  u.UserRole,
		UserPlan:  u.UserPlan,
		ExpireAt:  u.ExpireAt,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// StatsResponse 统计响应结构
type StatsResponse struct {
	// API统计
	TotalRequests    int64            `json:"totalRequests"`
	TotalTokens      int64            `json:"totalTokens"`
	InputTokens      int64            `json:"inputTokens"`
	OutputTokens     int64            `json:"outputTokens"`
	SuccessRate      float64          `json:"successRate"`
	AvgDuration      float64          `json:"avg_duration"`
	ModelUsage       map[string]int64 `json:"modelUsage"`
	ProviderUsage    map[string]int64 `json:"providerUsage"`
	RequestsToday    int64            `json:"requests_today"`
	RequestsThisWeek int64            `json:"requests_this_week"`

	// 会员统计
	TotalMembers          int64 `json:"totalMembers"`
	PaidMembers           int64 `json:"paidMembers"`
	MonthlyNewMembers     int64 `json:"monthlyNewMembers"`
	MonthlyNewPaidMembers int64 `json:"monthlyNewPaidMembers"`

	// 订单统计
	TotalOrders              int64   `json:"totalOrders"`
	SuccessfulOrders         int64   `json:"successfulOrders"`
	TotalRevenue             float64 `json:"totalRevenue"`
	SuccessfulRevenue        float64 `json:"successfulRevenue"`
	MonthlyRevenue           float64 `json:"monthlyRevenue"`
	MonthlySuccessfulRevenue float64 `json:"monthlySuccessfulRevenue"`
	WeeklyRevenue            float64 `json:"weeklyRevenue"`
	WeeklySuccessfulRevenue  float64 `json:"weeklySuccessfulRevenue"`
}

type ApiUsage struct {
	TotalTokens   uint64 `json:"total_tokens"`
	TotalRequests uint64 `json:"total_requests"`
	TotalProjects uint64 `json:"total_projects"`

	TodayTokens   uint64 `json:"today_tokens"`
	TodayRequests uint64 `json:"today_requests"`
	TodayProjects uint64 `json:"today_projects"`
}
type ApiLimit struct {
	// 有效期
	ExpireDays int `json:"expire_days,omitempty"`
	// CurrentPlan PayPlan `json:"current_plan,omitempty"`
	LimitMethod string `json:"limit_method,omitempty"`

	DailyTokens   uint64 `json:"daily_tokens,omitempty"`
	MonthlyTokens uint64 `json:"monthly_tokens,omitempty"`

	DailyRequests   uint64 `json:"daily_requests,omitempty"`
	MonthlyRequests uint64 `json:"monthly_requests,omitempty"`

	DailyProjects   uint64 `json:"daily_projects,omitempty"`
	MonthlyProjects uint64 `json:"monthly_projects,omitempty"`
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
