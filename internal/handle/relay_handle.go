package handle

import (
	"context"
	"encoding/json"
	"llm-member/internal/model"
	"llm-member/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RelayHandle struct {
	logService   *service.LogService
	userService  *service.UserService
	relayService *service.RelayService
	setupService *service.SetupService
}

func NewRelayHandle() *RelayHandle {
	return &RelayHandle{
		logService:   service.NewLogService(),
		userService:  service.NewUserService(),
		relayService: service.NewRelayService(),
		setupService: service.NewSetupService(),
	}
}

// ChatCompletions 聊天完成处理
func (h *RelayHandle) ChatCompletions(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户信息
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	userModel := user.(*model.UserModel)
	startTime := time.Now()

	// 调用 LLM 服务
	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	// 检查是否为流式请求
	if req.Stream {
		h.handleStreamResponse(c, ctx, &req, userModel, startTime)
		return
	}

	// 非流式响应
	h.handleNonStreamResponse(c, ctx, &req, userModel, startTime)
}

// handleNonStreamResponse 处理非流式响应
func (h *RelayHandle) handleNonStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, userModel *model.UserModel, startTime time.Time) {
	response, err := h.relayService.ChatCompletions(ctx, req)
	duration := time.Since(startTime)

	// 准备日志数据
	messagesJSON, _ := json.Marshal(req.Messages)
	responseJSON := ""
	tokensUsed := 0
	success := err == nil

	if success {
		responseBytes, _ := json.Marshal(response)
		responseJSON = string(responseBytes)
		if response.Usage.TotalTokens > 0 {
			tokensUsed = response.Usage.TotalTokens
		}
	}

	// 记录日志
	logEntry := &model.LlmLogModel{
		UserID: userModel.ID, TheModel: req.Model,
		Provider: h.getProviderFromModel(req.Model),
		Messages: string(messagesJSON), Response: responseJSON,
		TokensUsed: tokensUsed, Duration: duration.Milliseconds(),
	}

	if err != nil {
		logEntry.ErrorMsg = err.Error()
	}

	h.logService.CreateLog(logEntry)

	// 更新用户统计
	if success {
		h.userService.UpdateUserStats(userModel.ID, tokensUsed)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleStreamResponse 处理流式响应
func (h *RelayHandle) handleStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, userModel *model.UserModel, startTime time.Time) {
	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	// 获取流式响应通道
	responseChan, errorChan := h.relayService.ChatCompletionsStream(ctx, req)

	// 准备日志数据
	messagesJSON, _ := json.Marshal(req.Messages)
	var responseContent strings.Builder
	tokensUsed := 0
	duration := time.Since(startTime)
	var streamErr error

	// 处理流式数据
	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// 流结束
				goto logAndFinish
			}

			// 发送数据到客户端
			data, _ := json.Marshal(response)
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()

			// 收集响应内容用于日志
			if len(response.Choices) > 0 {
				responseContent.WriteString(response.Choices[0].Delta.Content)
			}

		case err, ok := <-errorChan:
			if ok && err != nil {
				streamErr = err
				// 发送错误信息
				errorData := map[string]interface{}{
					"error": map[string]string{
						"message": err.Error(),
						"type":    "error",
					},
				}
				errorJSON, _ := json.Marshal(errorData)
				c.Writer.Write([]byte("data: "))
				c.Writer.Write(errorJSON)
				c.Writer.Write([]byte("\n\n"))
				c.Writer.Flush()
			}
			goto logAndFinish

		case <-ctx.Done():
			streamErr = ctx.Err()
			goto logAndFinish
		}
	}

logAndFinish:
	// 发送结束标记
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.Flush()

	// 记录日志
	duration = time.Since(startTime)
	logEntry := &model.LlmLogModel{
		UserID: userModel.ID, TheModel: req.Model,
		Provider: h.getProviderFromModel(req.Model),
		Messages: string(messagesJSON), Response: responseContent.String(),
		TokensUsed: tokensUsed, Duration: duration.Milliseconds(),
	}

	if streamErr != nil {
		logEntry.ErrorMsg = streamErr.Error()
	} else {
		// 更新用户统计
		h.userService.UpdateUserStats(userModel.ID, tokensUsed)
	}

	h.logService.CreateLog(logEntry)
}

func (h *RelayHandle) getProviderFromModel(model string) string {
	if model == "" {
		return "unknown"
	}

	switch {
	case strings.HasPrefix(model, "gpt-"):
		return "openai"
	case strings.HasPrefix(model, "claude-"):
		return "claude"
	case strings.HasPrefix(model, "qwen-") || strings.HasPrefix(model, "qwen2"):
		return "qwen"
	case strings.HasPrefix(model, "doubao-"):
		return "doubao"
	case strings.HasPrefix(model, "glm-"):
		return "bigmodel"
	case strings.HasPrefix(model, "grok-"):
		return "grok"
	case strings.HasPrefix(model, "gemini-"):
		return "gemini"
	case strings.Contains(model, "/"):
		// OpenRouter 模型格式: provider/model
		return "openrouter"
	case strings.HasPrefix(model, "Qwen") || strings.HasPrefix(model, "deepseek"):
		return "siliconflow"
	default:
		return "unknown"
	}
}
