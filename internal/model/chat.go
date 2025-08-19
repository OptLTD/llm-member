package model

import (
	"time"

	"gorm.io/gorm"
)

// LlmLogModel 聊天日志模型
type LlmLogModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID   uint64 `json:"user_id" gorm:"column:user_id;index;not null;default:0"`
	ChatID   string `json:"chat_id" gorm:"column:chat_id;type:varchar(64);index"`
	ProjID   string `json:"proj_id" gorm:"column:proj_id;type:varchar(64);index;"`
	TheModel string `json:"model" gorm:"column:model;type:varchar(64);not null"`
	Provider string `json:"provider" gorm:"column:provider;type:varchar(64);not null"`

	Messages any `json:"messages" gorm:"type:text;serializer:json"`
	Response any `json:"response" gorm:"type:text;serializer:json"`
	AllUsage any `json:"all_usage" gorm:"type:text;serializer:json"`

	Duration  int64  `json:"duration" gorm:"not null;default:0"`
	Status    string `json:"status" gorm:"type:varchar(10);"`
	ErrorMsg  string `json:"error_msg" gorm:"type:varchar(256);"`
	ClientIP  string `json:"client_ip" gorm:"type:varchar(64);"`
	UserAgent string `json:"user_agent" gorm:"type:varchar(256);not null"`

	// 添加请求时间字段，用于按月分区查询
	ReqTime time.Time `json:"req_time" gorm:"column:req_time;index;"`

	User UserModel `json:"user" gorm:"foreignKey:UserID"`

	gorm.Model
}

func (m LlmLogModel) TableName() string {
	return "llm_log"
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model       string        `json:"model" binding:"required"`
	Messages    []ChatMessage `json:"messages" binding:"required"`
	Temperature *float32      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	TopP        *float32      `json:"top_p,omitempty"`
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

// ChatStreamResponse 流式聊天响应结构
type ChatStreamResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []ChatStreamChoice `json:"choices"`
	Usage   *Usage             `json:"usage,omitempty"` // 添加Usage字段
}

// ChatStreamChoice 流式聊天选择结构
type ChatStreamChoice struct {
	Index        int             `json:"index"`
	Delta        ChatStreamDelta `json:"delta"`
	FinishReason *string         `json:"finish_reason"`
	ToolCalls    []any           `json:"tool_calls,omitempty"`
}

// ChatStreamDelta 流式聊天增量结构
type ChatStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// LLModelInfo 模型信息结构
type LLModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Object   string `json:"object"`
	Provider string `json:"provider"`
}
