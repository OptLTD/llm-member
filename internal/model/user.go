package model

import (
	"time"

	relay "github.com/OptLTD/member/llm-relay"

	"gorm.io/gorm"
)

// UserModel 用户模型
type UserModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Avatar    string `json:"avatar" gorm:"column:avatar;type:varchar(256)"`
	Email     string `json:"email" gorm:"type:varchar(256);index;not null"`
	Username  string `json:"username" gorm:"type:varchar(128);index;not null"`
	Password  string `json:"password" gorm:"type:varchar(64);not null"`
	APIKey    string `json:"apiKey" gorm:"column:api_key;type:varchar(68);index"`
	IsActive  bool   `json:"isActive" gorm:"column:is_active;default:true"`
	Verified  bool   `json:"verified" gorm:"column:verified;default:false"`
	UserRole  Role   `json:"userRole" gorm:"column:user_role;default:user"`
	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`

	// 当前套餐、过期时间
	ExpireAt *time.Time `json:"expireAt" gorm:"column:expire_at;default:null"`     // 过期时间
	UserPlan PayPlan    `json:"userPlan" gorm:"column:user_plan;type:varchar(20)"` // 用户套餐
	// 用量/配额已拆到独立表 llm_user_quota，此处仅作内存/响应携带
	ApiUsage *ApiUsage `json:"apiUsage" gorm:"-"`
	ApiLimit *ApiLimit `json:"apiLimit" gorm:"-"`

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
	RefCode  string `json:"refCode,omitempty"`
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
		ApiUsage:  u.ApiUsage,
		ApiLimit:  u.ApiLimit,
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

	AnalyticsTodayPV    int64 `json:"analyticsTodayPV,omitempty"`
	AnalyticsWeekEvents int64 `json:"analyticsWeekEvents,omitempty"`
}

// ApiUsage 用户用量（类型定义与序列化见 llm-relay 包）
type ApiUsage = relay.ApiUsage

// ApiLimit 用户配额（类型定义与序列化见 llm-relay 包）
type ApiLimit = relay.ApiLimit

// GetID 实现 relay.User 接口
func (u *UserModel) GetID() uint64 { return u.ID }

// GetProjectID 实现 relay.User 接口
func (u *UserModel) GetProjectID() uint64 { return u.ProjectID }

// GetLimitMethod 实现 relay.User 接口；未配置配额时返回空字符串
func (u *UserModel) GetLimitMethod() string {
	if u.ApiLimit == nil {
		return ""
	}
	return u.ApiLimit.LimitMethod
}
