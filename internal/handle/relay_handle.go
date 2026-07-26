package handle

import (
	"context"
	"encoding/json"
	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/service"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RelayHandle struct {
	logService   *service.LogService
	userService  *service.UserService
	relayService *service.RelayService
	setupService *service.SetupService
	tokenService *service.TokenService
	quotaService *service.QuotaService
	statsService *service.StatsService
}

func NewRelayHandle() *RelayHandle {
	return &RelayHandle{
		logService:   service.NewLogService(),
		userService:  service.NewUserService(),
		relayService: service.NewRelayService(),
		setupService: service.NewSetupService(),
		tokenService: service.NewTokenService(),
		quotaService: service.NewQuotaService(),
		statsService: service.NewStatsService(),
	}
}

// 定义callback函数类型
type FinishCallback func(err error, response *model.ChatResponse)

func (h *RelayHandle) ChatCompletions(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var startTime = time.Now()
	var userInfo *model.UserModel
	if user, exists := c.Get("user"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	} else {
		userInfo = user.(*model.UserModel)
	}

	if err := h.tokenService.CheckUsage(userInfo); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	reserveAmount := uint64(1)
	if userInfo.ApiLimit != nil && userInfo.ApiLimit.LimitMethod == "tokens" {
		if n, err := h.tokenService.CountMsgsToken(req.Messages, req.Model, req.Stream); err == nil && n > 0 {
			reserveAmount = uint64(n)
		}
	}

	reservation, err := h.quotaService.Reserve(userInfo, reserveAmount)
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	storeBody := os.Getenv("LOG_STORE_BODY") == "true" || os.Getenv("LOG_STORE_BODY") == "1"

	finishCallback := func(err error, resp *model.ChatResponse) {
		if err != nil {
			_ = h.quotaService.Rollback(userInfo, reservation)
		} else {
			actual := reservation.Amount
			if resp != nil && userInfo.ApiLimit != nil && userInfo.ApiLimit.LimitMethod == "tokens" {
				if resp.Usage.TotalTokens > 0 {
					actual = uint64(resp.Usage.TotalTokens)
				}
			}
			_ = h.quotaService.Commit(userInfo, reservation, actual)
		}

		duration := time.Since(startTime).Milliseconds()
		provider := h.relayService.GetProvider(req.Model)
		logEntry := &model.LlmLogModel{
			UserID: userInfo.ID, Duration: duration,
			Provider: provider, TheModel: req.Model,
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
		_ = h.logService.CreateLog(logEntry)
	}

	if req.Stream {
		h.handleStreamResponse(c, ctx, &req, finishCallback)
		return
	}
	h.handleNonStreamResponse(c, ctx, &req, finishCallback)
}

// handleNonStreamResponse 处理非流式响应
func (h *RelayHandle) handleNonStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, callback FinishCallback) {
	response, err := h.relayService.ChatCompletions(ctx, req)
	if callback != nil {
		go callback(err, response)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *RelayHandle) handleStreamResponse(c *gin.Context, ctx context.Context, req *model.ChatRequest, callback FinishCallback) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	respChan, errorChan := h.relayService.ChatCompletionsStream(ctx, req)

	var streamErr error
	var accumulatedContent strings.Builder
	var providerUsage *model.Usage
	var response = &model.ChatResponse{
		Object: "chat.completion", Usage: model.Usage{},
		Model: req.Model, Choices: []model.ChatChoice{},
	}

	logAndFinish := func(err error) {
		c.Writer.Write([]byte("data: [DONE]\n\n"))
		c.Writer.Flush()

		if callback != nil {
			response.Choices = append(response.Choices, model.ChatChoice{
				Index: 0, FinishReason: "stop",
				Message: model.ChatMessage{
					Role: "assistant", Content: accumulatedContent.String(),
				},
			})

			if providerUsage != nil && providerUsage.TotalTokens > 0 {
				response.Usage = *providerUsage
			} else {
				promptTokens, _ := h.tokenService.CountMsgsToken(
					req.Messages, req.Model, true,
				)
				completionTokens := h.tokenService.CountTextToken(
					accumulatedContent.String(), req.Model,
				)
				response.Usage = model.Usage{
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
func (h *RelayHandle) ListModels(c *gin.Context) {
	models := h.relayService.GetModels()
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}
