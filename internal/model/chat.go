package model

import "gorm.io/gorm"

// LlmLogModel 聊天日志模型
type LlmLogModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID     uint64 `json:"user_id" gorm:"index"`
	TheModel   string `json:"model" gorm:"not null"`
	Provider   string `json:"provider" gorm:"not null"`
	Messages   string `json:"messages" gorm:"type:text"`
	Response   string `json:"response" gorm:"type:text"`
	TokensUsed int    `json:"tokens_used"`
	Duration   int64  `json:"duration"` // 毫秒
	Status     string `json:"status"`   // success, error
	ErrorMsg   string `json:"error_msg"`
	ClientIP   string `json:"client_ip"`
	UserAgent  string `json:"user_agent"`

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
}

// ChatStreamChoice 流式聊天选择结构
type ChatStreamChoice struct {
	Index        int             `json:"index"`
	Delta        ChatStreamDelta `json:"delta"`
	FinishReason *string         `json:"finish_reason"`
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
