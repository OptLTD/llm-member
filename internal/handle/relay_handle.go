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

// 定义callback函数类型
type FinishCallback func(err error, response *strings.Builder, usage *model.Usage)

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

	// 创建通用的日志记录callback
	finishCallback := func(err error, resp *strings.Builder, usage *model.Usage) {
		messagesJSON, _ := json.Marshal(req.Messages)
		duration := time.Since(startTime).Milliseconds()
		logEntry := &model.LlmLogModel{
			UserID: userModel.ID, TheModel: req.Model,
			Provider: h.getProviderFromModel(req.Model),
			Messages: string(messagesJSON), Response: resp.String(),
			TokensUsed: usage.TotalTokens, Duration: duration,
		}
		if err != nil {
			logEntry.ErrorMsg = err.Error()
		} else {
			h.userService.UpdateUserStats(userModel.ID, usage.TotalTokens)
		}
		h.logService.CreateLog(logEntry)
	}

	// 检查是否为流式请求
	if req.Stream {
		h.handleStreamResponse(c, ctx, &req, finishCallback)
		return
	}

	// 非流式响应
	h.handleNonStreamResponse(c, ctx, &req, finishCallback)
}

// handleNonStreamResponse 处理非流式响应
func (h *RelayHandle) handleNonStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, callback FinishCallback) {
	response, err := h.relayService.ChatCompletions(ctx, req)
	// 准备callback所需的数据
	var respBuilder strings.Builder
	var respUsage *model.Usage

	if err == nil {
		// 将响应转换为字符串用于日志记录
		responseBytes, _ := json.Marshal(response)
		respBuilder.WriteString(string(responseBytes))
		respUsage = &response.Usage
	}

	// 调用callback处理日志
	if callback != nil {
		callback(err, &respBuilder, respUsage)
	}

	// 返回响应给客户端
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelayHandle) handleStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, callback FinishCallback) {
	// 设置流式响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no")

	// 获取流式响应通道
	respChan, errorChan := h.relayService.ChatCompletionsStream(ctx, req)

	// 准备日志数据
	var streamErr error
	var response strings.Builder
	var tokenUsage *model.Usage

	// 定义完成处理的内部函数
	logAndFinish := func(err error) {
		// 发送结束标记
		c.Writer.Write([]byte("data: [DONE]\n\n"))
		c.Writer.Flush()

		// 调用外部传入的callback
		if callback != nil {
			callback(err, &response, tokenUsage)
		}
	}

	// 处理流式数据
	for {
		select {
		case resp, ok := <-respChan:
			if !ok { // 流结束
				logAndFinish(streamErr)
				return
			}

			// 检查是否包含usage信息（通常在最后一个响应中）
			if resp != nil && resp.Usage != nil {
				tokenUsage = resp.Usage
				// 如果这是只包含Usage信息的响应，不发送给客户端
				if resp.ID == "" && len(resp.Choices) == 0 {
					continue
				}
			}

			// 发送数据到客户端
			data, _ := json.Marshal(resp)
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()

			// 收集响应内容用于日志
			if len(resp.Choices) > 0 {
				response.WriteString(resp.Choices[0].Delta.Content)
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
			logAndFinish(streamErr)
			return
		case <-ctx.Done():
			streamErr = ctx.Err()
			logAndFinish(streamErr)
			return
		}
	}
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
