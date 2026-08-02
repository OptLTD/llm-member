package relay

// FinishCallback 完成回调：请求结束（成功或失败）时调用
type FinishCallback func(err error, response *ChatResponse)

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model       string        `json:"model" binding:""`
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
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`

	FinishReason string `json:"finish_reason"`
}

// Usage 使用情况结构
type Usage struct {
	PromptTokens int `json:"promptTokens"`
	TotalTokens  int `json:"totalTokens"`

	CompletionTokens int `json:"completionTokens"`
}

// ChatStreamResponse 流式聊天响应结构
type ChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`

	Usage *Usage `json:"usage,omitempty"` // 添加Usage字段

	Choices []ChatStreamChoice `json:"choices"`
}

// ChatStreamChoice 流式聊天选择结构
type ChatStreamChoice struct {
	Index        int     `json:"index"`
	FinishReason *string `json:"finish_reason"`
	ToolCalls    []any   `json:"tool_calls,omitempty"`

	Delta ChatStreamDelta `json:"delta"`
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
