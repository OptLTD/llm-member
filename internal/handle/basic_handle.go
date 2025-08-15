package handle

import (
	"net/http"

	"llm-member/internal/model"
	"llm-member/internal/service"

	"github.com/gin-gonic/gin"
)

type BasicHandle struct {
	userService  *service.UserService
	orderService *service.OrderService
	setupService *service.SetupService
}

func NewBasicHandle() *BasicHandle {
	userService := service.NewUserService()
	orderService := service.NewOrderService()
	setupService := service.NewSetupService()
	return &BasicHandle{
		userService:  userService,
		orderService: orderService,
		setupService: setupService,
	}
}

// GetUserProfile 获取用户个人资料
func (h *BasicHandle) GetUserProfile(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(user_id.(uint64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 清除敏感信息
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// SetUserProfile 更新用户个人资料
func (h *BasicHandle) SetUserProfile(c *gin.Context) {
	updateReq := &model.UpdateUserRequest{}
	if userID, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	} else {
		updateReq.UserId, _ = userID.(uint64)
	}

	var req struct {
		Email string `json:"email,omitempty"`
		Phone string `json:"phone,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建更新请求
	if req.Email != "" {
		updateReq.Email = req.Email
	}
	if req.Phone != "" {
		updateReq.Phone = req.Phone
	}

	// 调用UserService更新用户信息
	user, err := h.userService.UpdateUser(updateReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 清除敏感信息
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "个人资料更新成功",
	})
}

// GetUserUsage 获取用户使用统计
func (h *BasicHandle) GetUserUsage(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取时间范围参数
	period := c.DefaultQuery("period", "month") // day, week, month, year

	// 简单的使用统计，实际应该在UserService中实现GetUserUsageStats方法
	user, err := h.userService.GetUserByID(user_id.(uint64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取使用统计失败"})
		return
	}

	usage := gin.H{
		"total_tokens":   user.TotalTokens,
		"total_requests": user.TotalRequests,
		"daily_limit":    user.DailyLimit,
		"monthly_limit":  user.MonthlyLimit,
		"current_plan":   user.CurrPlan,
		"period":         period,
	}

	c.JSON(http.StatusOK, usage)
}

// GetUserAPIKeys 获取用户API密钥
func (h *BasicHandle) GetUserAPIKeys(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.GetUserByID(user_id.(uint64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key":       user.APIKey,
		"created_at":    user.CreatedAt,
		"daily_limit":   user.DailyLimit,
		"monthly_limit": user.MonthlyLimit,
	})
}

// GetUserOrders 用户订单
func (h *BasicHandle) GetUserOrders(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	data, err := h.userService.GetUserOrders(user_id.(uint64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": data})
}

// RegenerateKey 重新生成用户API密钥
func (h *BasicHandle) RegenerateKey(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	user, err := h.userService.RegenerateKey(user_id.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新生成API密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": user.APIKey,
		"message": "API密钥重新生成成功",
	})
}
