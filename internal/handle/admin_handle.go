package handle

import (
	"fmt"
	"net/http"
	"strconv"

	"llm-member/internal/model"
	"llm-member/internal/service"
	"llm-member/internal/support"

	"github.com/gin-gonic/gin"
)

type AdminHandle struct {
	logService   *service.LogService
	userService  *service.UserService
	orderService *service.OrderService
	relayService *service.RelayService
	setupService *service.SetupService
	statsService *service.StatsService
}

func NewAdminHandle() *AdminHandle {
	return &AdminHandle{
		logService:   service.NewLogService(),
		userService:  service.NewUserService(),
		orderService: service.NewOrderService(),
		relayService: service.NewRelayService(),
		setupService: service.NewSetupService(),
		statsService: service.NewStatsService(),
	}
}

// CreateUser 用户注册
func (h *AdminHandle) CreateUser(c *gin.Context) {
	var req model.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &model.UserModel{
		Username: req.Username,
		Email:    req.Email,
	}
	user.APIKey, _ = support.GenerateAPIKey()
	err := h.userService.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model.SignUpResponse{
		User: user, APIKey: user.APIKey, Message: "注册成功",
	})
}

// UpdateUser 更新用户信息（管理员功能）
func (h *AdminHandle) UpdateUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserId = userID
	if user, err := h.userService.UpdateUser(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"user": user.ToUserResponse(),
		})
	}
}

// RegenerateAPIKey 重新生成用户 API Key
func (h *AdminHandle) GenerateAPIKey(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := h.userService.RegenerateKey(uint(userID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": user.APIKey, "user": user.ToUserResponse(),
	})
}

// DeleteUser 删除用户（管理员功能）
func (h *AdminHandle) DeleteUser(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.userService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// ToggleUserStatus 切换用户状态（管理员功能）
func (h *AdminHandle) ToggleUserStatus(c *gin.Context) {
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

// Current 获取当前用户信息
func (h *AdminHandle) Current(c *gin.Context) {
	var user_id uint64
	if userID, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	} else {
		user_id = userID.(uint64)
	}

	user, err := h.userService.GetUserByID(user_id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetModels 获取可用模型
func (h *AdminHandle) GetModels(c *gin.Context) {
	models := h.relayService.GetModels()
	c.JSON(http.StatusOK, gin.H{"data": models})
}

// GetStats 获取统计信息
func (h *AdminHandle) GetStats(c *gin.Context) {
	stats, err := h.statsService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetLogs 获取日志
func (h *AdminHandle) GetLogs(c *gin.Context) {
	var req model.PaginateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Page = 1
		req.Size = 10
	}

	// 参数校验
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 调用服务层方法
	response, err := h.logService.GetLogs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetUsers 获取用户列表（管理员功能）
func (h *AdminHandle) GetUsers(c *gin.Context) {
	var req model.PaginateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Page = 1
		req.Size = 10
	}

	// 参数校验
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 调用服务层方法
	response, err := h.userService.QueryUsers(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetOrders 获取订单
func (h *AdminHandle) GetOrders(c *gin.Context) {
	var req model.PaginateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Page = 1
		req.Size = 10
	}

	// 参数校验
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// 调用服务层方法
	response, err := h.orderService.QueryOrders(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
