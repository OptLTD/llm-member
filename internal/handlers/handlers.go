package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"swiflow-auth/internal/config"
	"swiflow-auth/internal/models"
	"swiflow-auth/internal/services"
)

type Handlers struct {
	authService *services.AuthService
	llmService  *services.LLMService
	logService  *services.LogService
	userService *services.UserService
	config      *config.Config
}

func New(llmService *services.LLMService, logService *services.LogService, authService *services.AuthService, userService *services.UserService, cfg *config.Config) *Handlers {
	return &Handlers{
		llmService:  llmService,
		logService:  logService,
		authService: authService,
		userService: userService,
		config:      cfg,
	}
}

// Login 用户登录
func (h *Handlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:   token,
		User:    user,
		Message: "登录成功",
	})
}

// Logout 用户登出
func (h *Handlers) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		h.authService.Logout(token)
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// Register 用户注册
func (h *Handlers) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.RegisterResponse{
		User:    user,
		APIKey:  user.APIKey,
		Message: "注册成功",
	})
}

// GetUsers 获取用户列表（管理员功能）
func (h *Handlers) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.GetAllUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UserListResponse{
		Users: users,
		Total: total,
	})
}

// UpdateUser 更新用户信息（管理员功能）
func (h *Handlers) UpdateUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(uint(userID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "用户信息更新成功",
	})
}

// RegenerateAPIKey 重新生成用户 API Key
func (h *Handlers) RegenerateAPIKey(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := h.userService.RegenerateAPIKey(uint(userID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"api_key": user.APIKey,
		"message": "API Key 重新生成成功",
	})
}

// DeleteUser 删除用户（管理员功能）
func (h *Handlers) DeleteUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.userService.DeleteUser(uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// ToggleUserStatus 切换用户状态（管理员功能）
func (h *Handlers) ToggleUserStatus(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ToggleUserStatus(uint(userID), req.IsActive); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action := "禁用"
	if req.IsActive {
		action = "启用"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("用户%s成功", action)})
}

// GetCurrentUser 获取当前用户信息
func (h *Handlers) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// TestChat 测试用聊天接口（使用 Web Token 认证）
func (h *Handlers) TestChat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取用户信息（从 Web Token 认证中间件）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	// 获取用户详细信息
	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	startTime := time.Now()

	// 调用 LLM 服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, err := h.llmService.ChatCompletions(ctx, &req)
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
	logEntry := &models.ChatLog{
		UserID:     user.ID,
		Model:      req.Model,
		Provider:   h.getProviderFromModel(req.Model),
		Messages:   string(messagesJSON),
		Response:   responseJSON,
		TokensUsed: tokensUsed,
		Duration:   duration.Milliseconds(),
	}

	if err != nil {
		logEntry.ErrorMsg = err.Error()
	}

	h.logService.CreateLog(logEntry)

	// 更新用户统计
	if success {
		h.userService.UpdateUserStats(user.ID, tokensUsed)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChatCompletions 聊天完成处理
func (h *Handlers) ChatCompletions(c *gin.Context) {
	var req models.ChatRequest
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

	userModel := user.(*models.User)
	startTime := time.Now()

	// 调用 LLM 服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, err := h.llmService.ChatCompletions(ctx, &req)
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
	logEntry := &models.ChatLog{
		UserID:     userModel.ID,
		Model:      req.Model,
		Provider:   h.getProviderFromModel(req.Model),
		Messages:   string(messagesJSON),
		Response:   responseJSON,
		TokensUsed: tokensUsed,
		Duration:   duration.Milliseconds(),
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

// GetModels 获取可用模型
func (h *Handlers) GetModels(c *gin.Context) {
	models := h.llmService.GetModels()
	c.JSON(http.StatusOK, gin.H{"data": models})
}

// GetLogs 获取日志
func (h *Handlers) GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.logService.GetLogs(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       logs,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// DeleteLog 删除日志
func (h *Handlers) DeleteLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	if err := h.logService.DeleteLog(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}

// GetStats 获取统计信息
func (h *Handlers) GetStats(c *gin.Context) {
	stats, err := h.logService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetConfig 获取配置
func (h *Handlers) GetConfig(c *gin.Context) {
	models := h.llmService.GetModels()
	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// UpdateConfig 更新配置
func (h *Handlers) UpdateConfig(c *gin.Context) {
	// 这里可以实现配置更新逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

func (h *Handlers) getProviderFromModel(model string) string {
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