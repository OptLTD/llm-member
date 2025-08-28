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

	Avatar   string `json:"avatar" gorm:"column:avatar;type:varchar(256)"`
	Email    string `json:"email" gorm:"type:varchar(256);index;not null"`
	Username string `json:"username" gorm:"type:varchar(128);index;not null"`
	Password string `json:"password" gorm:"type:varchar(64);not null"`
	APIKey   string `json:"apiKey" gorm:"column:api_key;type:varchar(68);index"`
	IsActive bool   `json:"isActive" gorm:"column:is_active;default:true"`
	Verified bool   `json:"verified" gorm:"column:verified;default:false"`
	UserRole Role   `json:"userRole" gorm:"column:user_role;default:user"`

	// 当前套餐、过期时间
	ExpireAt *time.Time `json:"expireAt" gorm:"column:expire_at;default:null"`     // 过期时间
	UserPlan PayPlan    `json:"userPlan" gorm:"column:user_plan;type:varchar(20)"` // 用户套餐
	ApiUsage *ApiUsage  `json:"apiUsage" gorm:"column:api_usage;serializer:json"`
	ApiLimit *ApiLimit  `json:"apiLimit" gorm:"column:api_limit;serializer:json"`

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
	APIKey  string     `json:"apiKey"`
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
	UserId uint64 `json:"userId,omitempty"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`

	IsActive *bool    `json:"isActive,omitempty"`
	UserPlan *PayPlan `json:"userPlan,omitempty"`

	DailyLimit   *int64 `json:"dailyLimit,omitempty"`
	MonthlyLimit *int64 `json:"monthlyLimit,omitempty"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	APIKey   string `json:"apiKey"`
	Username string `json:"username"`
	IsActive bool   `json:"isActive"`
	Verified bool   `json:"verified"`
	UserRole Role   `json:"userRole"`

	UserPlan PayPlan    `json:"userPlan"`
	ExpireAt *time.Time `json:"expireAt"`
	ApiUsage *ApiUsage  `json:"apiUsage"`
	ApiLimit *ApiLimit  `json:"apiLimit"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToUserResponse 将UserModel转换为UserResponse
func (u *UserModel) ToUserResponse() *UserResponse {
	return &UserResponse{
		Email:     u.Email,
		Avatar:    u.Avatar,
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
	AvgDuration      float64          `json:"avgDuration"`
	ModelUsage       map[string]int64 `json:"modelUsage"`
	ProviderUsage    map[string]int64 `json:"providerUsage"`
	RequestsToday    int64            `json:"requestsToday"`
	RequestsThisWeek int64            `json:"requestsThisWeek"`

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
	TotalTokens   uint64 `json:"totalTokens"`
	TotalRequests uint64 `json:"totalRequests"`
	TotalProjects uint64 `json:"totalProjects"`

	TodayTokens   uint64 `json:"todayTokens"`
	TodayRequests uint64 `json:"todayRequests"`
	TodayProjects uint64 `json:"todayProjects"`
}
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
