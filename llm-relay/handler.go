package relay

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetUserFunc 从 gin 上下文中取出当前用户；默认读取 "user" 键并断言为 User。
type GetUserFunc func(c *gin.Context) (User, bool)

// ServiceForFunc 按请求解析 Service（例如按项目动态配置上游）
type ServiceForFunc func(c *gin.Context) *Service

// Handler LLM 中继的 HTTP 处理器。
type Handler struct {
	service      *Service
	svcFor       ServiceForFunc
	tokenCounter *TokenCounter
	quota        Quota
	logs         Logger
	getUser      GetUserFunc
}

// HandlerOption 处理器选项
type HandlerOption func(*Handler)

// HandlerWithServiceFor 自定义按请求解析 Service（例如按项目 LLM 配置）
func HandlerWithServiceFor(fn ServiceForFunc) HandlerOption {
	return func(h *Handler) {
		if fn != nil {
			h.svcFor = fn
		}
	}
}

// HandlerWithUser 自定义用户提取方式
func HandlerWithUser(getUser GetUserFunc) HandlerOption {
	return func(h *Handler) {
		if getUser != nil {
			h.getUser = getUser
		}
	}
}

// NewHandler 创建中继处理器。
// 默认从 gin 上下文 "user" 键取用户。
func NewHandler(service *Service, quota Quota, logs Logger, opts ...HandlerOption) *Handler {
	h := &Handler{
		service:      service,
		tokenCounter: NewTokenCounter(),
		quota:        quota,
		logs:         logs,
		getUser: func(c *gin.Context) (User, bool) {
			v, ok := c.Get("user")
			if !ok {
				return nil, false
			}
			u, ok := v.(User)
			return u, ok
		},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// svc 返回本次请求使用的 Service
func (h *Handler) svc(c *gin.Context) *Service {
	if h.svcFor != nil {
		return h.svcFor(c)
	}
	return h.service
}

// ChatCompletions 处理 /v1/chat/completions
func (h *Handler) ChatCompletions(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime = time.Now()
	var userInfo User
	if user, exists := h.getUser(c); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	} else {
		userInfo = user
	}

	if err := h.quota.CheckUsage(userInfo); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	reserveAmount := uint64(1)
	if userInfo.GetLimitMethod() == "tokens" {
		if n, err := h.tokenCounter.CountMsgsToken(req.Messages, req.Model, req.Stream); err == nil && n > 0 {
			reserveAmount = uint64(n)
		}
	}

	reservation, err := h.quota.Reserve(userInfo, reserveAmount)
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	storeBody := os.Getenv("LOG_STORE_BODY") == "true" || os.Getenv("LOG_STORE_BODY") == "1"

	finishCallback := func(err error, resp *ChatResponse) {
		if err != nil {
			_ = h.quota.Rollback(userInfo, reservation)
		} else {
			actual := reserveAmount
			if resp != nil && userInfo.GetLimitMethod() == "tokens" {
				if resp.Usage.TotalTokens > 0 {
					actual = uint64(resp.Usage.TotalTokens)
				}
			}
			_ = h.quota.Commit(userInfo, reservation, actual)
		}

		duration := time.Since(startTime).Milliseconds()
		provider := h.svc(c).GetProvider(req.Model)
		logEntry := &LogEntry{
			ProjectID: userInfo.GetProjectID(), UserID: userInfo.GetID(), Duration: duration,
			Provider: provider, Model: req.Model,
			ReqTime: time.Now(), ClientIP: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
		}
		if storeBody {
			logEntry.Messages = req.Messages
			logEntry.Response = resp
		}
		if c.GetHeader("X-Project-Id") != "" {
			logEntry.ProjID = c.GetHeader("X-Project-Id")
		}
		if err != nil {
			logEntry.Status = "failure"
			msg := err.Error()
			if len(msg) > 256 {
				msg = msg[:256]
			}
			logEntry.ErrorMsg = msg
		} else if resp != nil {
			logEntry.Status = "success"
			logEntry.ChatID = resp.ID
			logEntry.AllUsage = resp.Usage
		}
		_ = h.logs.CreateLog(logEntry)
	}

	if req.Stream {
		h.handleStreamResponse(c, ctx, &req, finishCallback)
		return
	}
	h.handleNonStreamResponse(c, ctx, &req, finishCallback)
}

// handleNonStreamResponse 处理非流式响应
func (h *Handler) handleNonStreamResponse(c *gin.Context, ctx context.Context, req *ChatRequest, callback FinishCallback) {
	response, err := h.svc(c).ChatCompletions(ctx, req)
	if callback != nil {
		go callback(err, response)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) handleStreamResponse(c *gin.Context, ctx context.Context, req *ChatRequest, callback FinishCallback) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	respChan, errorChan := h.svc(c).ChatCompletionsStream(ctx, req)

	var streamErr error
	var accumulatedContent strings.Builder
	var providerUsage *Usage
	var response = &ChatResponse{
		Object: "chat.completion", Usage: Usage{},
		Model: req.Model, Choices: []ChatChoice{},
	}

	logAndFinish := func(err error) {
		c.Writer.Write([]byte("data: [DONE]\n\n"))
		c.Writer.Flush()

		if callback != nil {
			response.Choices = append(response.Choices, ChatChoice{
				Index: 0, FinishReason: "stop",
				Message: ChatMessage{
					Role: "assistant", Content: accumulatedContent.String(),
				},
			})

			if providerUsage != nil && providerUsage.TotalTokens > 0 {
				response.Usage = *providerUsage
			} else {
				promptTokens, _ := h.tokenCounter.CountMsgsToken(
					req.Messages, req.Model, true,
				)
				completionTokens := h.tokenCounter.CountTextToken(
					accumulatedContent.String(), req.Model,
				)
				response.Usage = Usage{
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
					TotalTokens:      promptTokens + completionTokens,
				}
			}
			go callback(err, response)
		}
	}

	for {
		select {
		case resp, ok := <-respChan:
			if !ok {
				logAndFinish(streamErr)
				return
			}

			if response.ID == "" {
				response.ID = resp.ID
				response.Created = resp.Created
			}

			if resp.Usage != nil && resp.Usage.TotalTokens > 0 {
				providerUsage = resp.Usage
			}

			for _, choice := range resp.Choices {
				if choice.Delta.Content != "" {
					accumulatedContent.WriteString(choice.Delta.Content)
				}
			}

			data, _ := json.Marshal(resp)
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()
		case err, ok := <-errorChan:
			if ok && err != nil {
				streamErr = err
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

// ListModels OpenAI 兼容的模型列表
func (h *Handler) ListModels(c *gin.Context) {
	models := h.svc(c).GetModels()
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}
