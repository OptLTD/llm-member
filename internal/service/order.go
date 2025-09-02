package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"
	"llm-member/internal/payment"

	"gorm.io/gorm"
)

// OrderService 支付服务
type OrderService struct {
	db *gorm.DB
}

// getProviderFromMethod 根据支付方式获取提供商类型
func (s *OrderService) getProviderFromMethod(method model.PaymentMethod) consts.PaymentProvider {
	switch method {
	case "stripe":
		return consts.ProviderStripe
	case "alipay":
		return consts.ProviderAlipay
	case "wechat":
		return consts.ProviderWechat
	case "paypal":
		return consts.ProviderPaypal
	case "creem":
		return consts.ProviderCreem
	default:
		return consts.PaymentProvider(string(method))
	}
}

// NewOrderService 创建支付服务实例
func NewOrderService() *OrderService {
	return &OrderService{
		db: config.GetDB(),
	}
}

func (s *OrderService) CreateOrder(req *model.OrderRequest, plan *model.PlanInfo) (*model.OrderModel, error) {
	// 检查支付方式是否可用
	if !config.HasPayment(string(req.Method)) {
		return nil, fmt.Errorf("%w %s", consts.ErrPaymentMethodNotEnabled, req.Method)
	}
	// 生成订单ID, 创建订单记录
	order := &model.OrderModel{
		UserID:  *req.UserId,
		PayPlan: req.PayPlan, Amount: *req.Amount,
		Method: req.Method, Status: model.PaymentPending,
		ExpiredAt: time.Now().Add(30 * time.Minute), // 30分钟过期
	}
	if orderID, err := s.generateOrderID(); err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrOrderIDGenerationFailed, err)
	} else {
		order.OrderID = orderID
	}

	// 保存订单到数据库
	if err := s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrOrderSaveFailed, err)
	}
	// 支付订单
	providerType := s.getProviderFromMethod(req.Method)
	provider := payment.NewPayment(req.Method)
	if err := provider.Create(order); err != nil {
		consts.LogDetailedError(providerType, consts.ErrorTypeCreation, err, "create payment order")
		return nil, consts.GetFriendlyError(err)
	}
	if err := s.db.Save(order).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrOrderSaveFailed, err)
	}
	return order, nil
}

// FindOrder 获取支付订单
func (s *OrderService) FindOrder(orderID string) (*model.OrderModel, error) {
	var order model.OrderModel
	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// VerifyCallback 验证支付回调
func (s *OrderService) VerifyCallback(name string, req *http.Request) (*model.OrderModel, error) {
	method := model.PaymentMethod(name)
	providerType := s.getProviderFromMethod(method)
	provider := payment.NewPayment(method)

	var orderId string
	// 使用支付提供商的Webhook方法验证回调
	if event, err := provider.Webhook(req); err != nil {
		consts.LogDetailedError(
			providerType, consts.ErrorTypeConfig, err,
			"verify payment webhook",
		)
		return nil, consts.GetFriendlyError(err)
	} else if event == nil || event.OrderID == "" {
		return nil, consts.ErrPaymentWebhookMissingParams
	} else {
		orderId = event.OrderID
		log.Printf(
			"[webhook] 收到支付事件: Type=%s, OrderID=%s, Status=%s, Amount=%.2f",
			event.Type, event.OrderID, event.Status, event.Amount,
		)
	}

	if order, err := s.FindOrder(orderId); err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrOrderQueryFailed, err)
	} else {
		return order, nil
	}
}

func (s *OrderService) QueryPayment(order *model.OrderModel) error {
	if order.Status != model.PaymentPending {
		return nil
	}

	providerType := s.getProviderFromMethod(order.Method)
	provider := payment.NewPayment(order.Method)
	if err := provider.Query(order); err != nil {
		consts.LogDetailedError(
			providerType, consts.ErrorTypeQuery, err,
			"query payment status",
		)
		return consts.GetFriendlyError(err)
	}
	return nil
}

// UpdatePaymentStatus 更新支付状态
func (s *OrderService) UpdateStatus(orderID string, status string) error {
	updates := map[string]any{
		"status": status,
	}

	if status == string(model.PaymentSucceed) {
		now := time.Now()
		updates["succeed_at"] = &now
	}

	return s.db.Model(&model.OrderModel{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

// ProcessPaymentSuccess 处理支付成功
func (s *OrderService) PaySuccess(orderId string, limit *model.ApiLimit) error {
	order, err := s.FindOrder(orderId)
	if err != nil {
		return fmt.Errorf("%w: %v", consts.ErrOrderQueryFailed, err)
	}
	if order.Status == model.PaymentSucceed {
		return nil // 已经处理过了
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	var updateOrder = map[string]any{
		"status": model.PaymentSucceed, "succeed_at": time.Now(),
	}
	if err := tx.Model(order).Updates(updateOrder).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: %v", consts.ErrOrderStatusUpdateFailed, err)
	}

	// 更新用户套餐
	var updateUser = map[string]any{
		"user_plan": order.PayPlan, "api_limit": limit,
		"expire_at": time.Now().AddDate(0, 0, limit.ExpireDays),
	}
	var query = tx.Model(&model.UserModel{}).Where("id = ?", order.UserID)
	if err := query.Updates(updateUser).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: %v", consts.ErrUserPlanUpdateFailed, err)
	}

	tx.Commit()
	return nil
}

// GetAllOrders 获取用户订单列表
func (s *OrderService) QueryOrders(req *model.PaginateRequest) (*model.PaginateResponse, error) {
	var total int64
	var orders []model.OrderModel

	query := s.db.Model(&model.OrderModel{})
	query.Where(req.Query).Order("created_at DESC")
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 获取分页数据
	offset := int((req.Page - 1) * req.Size)
	if err := query.Offset(offset).Limit(int(req.Size)).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	// 构造响应
	response := &model.PaginateResponse{
		Data: orders, Page: req.Page, Size: req.Size, Total: total,
		Count: uint((total + int64(req.Size) - 1) / int64(req.Size)),
	}
	return response, nil
}

// generateOrderID 生成订单ID
func (s *OrderService) generateOrderID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("PAY%d%s", time.Now().Unix(), hex.EncodeToString(bytes)[:8]), nil
}
