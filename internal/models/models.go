package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Username    string `json:"username" gorm:"uniqueIndex;not null"`
	Email       string `json:"email" gorm:"uniqueIndex;not null"`
	Password    string `json:"-" gorm:"not null"` // 不在 JSON 中返回密码
	APIKey      string `json:"api_key" gorm:"uniqueIndex;not null"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	IsAdmin     bool   `json:"is_admin" gorm:"default:false"`
	
	// 使用统计
	TotalRequests int64 `json:"total_requests" gorm:"default:0"`
	TotalTokens   int64 `json:"total_tokens" gorm:"default:0"`
	
	// 限制设置
	DailyLimit   int64 `json:"daily_limit" gorm:"default:1000"`   // 每日请求限制
	MonthlyLimit int64 `json:"monthly_limit" gorm:"default:10000"` // 每月请求限制
}

// ChatLog 聊天日志模型
type ChatLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	UserID      uint   `json:"user_id" gorm:"index"`
	User        User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Model       string `json:"model" gorm:"not null"`
	Provider    string `json:"provider" gorm:"not null"`
	Messages    string `json:"messages" gorm:"type:text"`
	Response    string `json:"response" gorm:"type:text"`
	TokensUsed  int    `json:"tokens_used"`
	Duration    int64  `json:"duration"` // 毫秒
	Status      string `json:"status"`   // success, error
	ErrorMsg    string `json:"error_msg"`
	ClientIP    string `json:"client_ip"`
	UserAgent   string `json:"user_agent"`
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model       string         `json:"model" binding:"required"`
	Messages    []ChatMessage  `json:"messages" binding:"required"`
	Temperature *float32       `json:"temperature,omitempty"`
	MaxTokens   *int           `json:"max_tokens,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	TopP        *float32       `json:"top_p,omitempty"`
}

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

// ChatChoice 聊天选择结构
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 使用情况结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo 模型信息结构
type ModelInfo struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User    *User  `json:"user"`
	APIKey  string `json:"api_key"`
	Message string `json:"message"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token   string `json:"token"`
	User    *User  `json:"user"`
	Message string `json:"message"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email        string `json:"email,omitempty"`
	IsActive     *bool  `json:"is_active,omitempty"`
	DailyLimit   *int64 `json:"daily_limit,omitempty"`
	MonthlyLimit *int64 `json:"monthly_limit,omitempty"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

// StatsResponse 统计响应结构
type StatsResponse struct {
	TotalRequests    int64            `json:"total_requests"`
	TotalTokens      int64            `json:"total_tokens"`
	SuccessRate      float64          `json:"success_rate"`
	AvgDuration      float64          `json:"avg_duration"`
	ModelUsage       map[string]int64 `json:"model_usage"`
	ProviderUsage    map[string]int64 `json:"provider_usage"`
	RequestsToday    int64            `json:"requests_today"`
	RequestsThisWeek int64            `json:"requests_this_week"`
}