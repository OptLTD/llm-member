package model

import (
	"time"

	"gorm.io/gorm"
)

// UserModel 用户模型
type UserModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Phone    string `json:"phone" gorm:"uniqueIndex"`
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"` // 不在 JSON 中返回密码
	APIKey   string `json:"api_key" gorm:"uniqueIndex;not null"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
	Verified bool   `json:"verified" gorm:"default:false"`
	CurrRole Role   `json:"curr_role" gorm:"default:user"`

	// 当前套餐、过期时间
	CurrPlan  PayPlan    `json:"curr_plan" gorm:"type:varchar(20)"` // 用户套餐
	ExpiredAt *time.Time `json:"expired_at" gorm:"default:null"`    // 过期时间

	// 使用统计
	TotalTokens   int64 `json:"total_tokens" gorm:"default:0"`
	TotalRequests int64 `json:"total_requests" gorm:"default:0"`

	// 限制设置
	DailyLimit   int64 `json:"daily_limit" gorm:"default:1000"`    // 每日请求限制
	MonthlyLimit int64 `json:"monthly_limit" gorm:"default:10000"` // 每月请求限制

	gorm.Model
}

func (m UserModel) TableName() string {
	return "llm_user"
}

// CreateUserReq 创建用户请求
type CreateUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// SignUpRequest 用户端注册请求
type SignUpRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
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
	PayPlan  *PayPlan `json:"pay_plan,omitempty"`

	DailyLimit   *int64 `json:"daily_limit,omitempty"`
	MonthlyLimit *int64 `json:"monthly_limit,omitempty"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []UserModel `json:"users"`
	Total int64       `json:"total"`
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
