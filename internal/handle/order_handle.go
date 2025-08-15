package handle

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type OrderHandler struct {
	userService  *service.UserService
	setupService *service.SetupService
	orderService *service.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		userService:  service.NewUserService(),
		setupService: service.NewSetupService(),
		orderService: service.NewOrderService(),
	}
}

// GetPaymentMethods 获取可用的支付方式
func (h *OrderHandler) GetPaymentMethods(c *gin.Context) {
	methods := []gin.H{}

	// 检查各支付方式是否可用
	if config.HasPaymentProvider("mock") {
		methods = append(methods, gin.H{
			"method": "mock", "name": "模拟支付",
			"icon": "fab fa-debug", "color": "#1677FF",
		})
	}
	if config.HasPaymentProvider("alipay") {
		methods = append(methods, gin.H{
			"method": "alipay", "name": "支付宝",
			"icon": "fab fa-alipay", "color": "#1677FF",
		})
	}

	if config.HasPaymentProvider("wechat") {
		methods = append(methods, gin.H{
			"method": "wechat", "name": "微信支付",
			"icon": "fab fa-weixin", "color": "#07C160",
		})
	}

	if config.HasPaymentProvider("union") {
		methods = append(methods, gin.H{
			"method": "union", "name": "银联支付",
			"icon": "fas fa-credit-card", "color": "#E60012",
		})
	}

	if config.HasPaymentProvider("stripe") {
		methods = append(methods, gin.H{
			"method": "stripe", "name": "Stripe",
			"icon": "stripe", "color": "#008000",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"methods": methods,
	})
}

// ShowPaymentQrcode 获取可用的支付方式
func (h *OrderHandler) ShowPaymentQrcode(c *gin.Context) {
	if orderID := c.Param("id"); orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID不能为空"})
		return
	}

	order, err := h.orderService.QueryOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if uid, exists := c.Get("user_id"); !exists || uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	} else if order.UserID != uid.(uint64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此订单"})
		return
	}
	qrContent := order.PayURL
	if order.Method == model.PaymentMock {
		qrContent = `
			this is a mock payment, nice to meet you!
			click 'check button' to get success status!
		`
	} else if order.QRCode != "" {
		qrContent = order.QRCode
	}

	// 生成二维码图片
	qrCode, err := qrcode.New(qrContent, qrcode.Medium)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
		return
	}

	// 设置二维码大小
	qrCode.DisableBorder = false

	// 生成PNG格式的二维码图片
	var buf bytes.Buffer
	if err := qrCode.Write(256, &buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码图片失败"})
		return
	}

	// 返回PNG图片
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

// CreatePaymentOrder 创建支付订单
func (h *OrderHandler) CreatePaymentOrder(c *gin.Context) {
	var uid uint64
	if id, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	} else if uid, _ = id.(uint64); uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req model.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		req.UserId = &uid // 注入 user id
	}

	user, err := h.userService.GetUserByID(uint64(*req.UserId))
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}
	if user.ExpiredAt != nil && user.ExpiredAt.After(time.Now()) {
		c.JSON(http.StatusOK, gin.H{"error": "您已订阅套餐，不能重复订阅",
			"plan": user.CurrPlan, "expired_at": user.ExpiredAt,
		})
		return
	}

	plan := model.PlanInfo{Plan: string(req.PayPlan)}
	err = h.setupService.GetAsTarget("plan."+plan.Plan, &plan)
	if err != nil || plan.Price == 0 || plan.Enabled == false {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取套餐价格失败"})
		return
	}

	// 创建支付订单（这里模拟支付接口）
	order, err := h.orderService.CreateOrder(&req, &plan)
	response := model.OrderResponse{
		Amount: order.Amount, QRCode: order.QRCode,
		OrderID: order.OrderID, PayURL: order.PayURL,
		Status: string(order.Status), Method: string(order.Method),
	}
	c.JSON(http.StatusOK, response)
}

// QueryPaymentOrder 获取支付订单详情
func (h *OrderHandler) QueryPaymentOrder(c *gin.Context) {
	if orderID := c.Param("id"); orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID不能为空"})
		return
	}

	order, err := h.orderService.QueryOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if uid, exists := c.Get("user_id"); !exists || uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	} else if order.UserID != uid.(uint64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此订单"})
		return
	}

	currStatus := order.Status
	err = h.orderService.QueryPayment(order)
	if err == nil && currStatus != order.Status {
		if err := h.orderService.PaySuccess(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, order)
}

// DoPaymentCallback 支付回调处理
func (h *OrderHandler) DoPaymentCallback(c *gin.Context) {
	if orderId := c.Param("id"); orderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID不能为空"})
		return
	}

	order, err := h.orderService.QueryOrder(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	// 这里应该验证支付平台的回调签名
	// 为了演示，我们直接标记为支付成功
	if err := h.orderService.PaySuccess(order.OrderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "支付成功"})
}
