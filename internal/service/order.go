package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/payment"

	"gorm.io/gorm"
)

// OrderService 支付服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建支付服务实例
func NewOrderService() *OrderService {
	return &OrderService{
		db: config.GetDB(),
	}
}

func (s *OrderService) CreateOrder(req *model.OrderRequest, plan *model.PlanInfo) (*model.OrderModel, error) {
	// 检查支付方式是否可用
	if !config.HasPaymentProvider(string(req.Method)) {
		return nil, fmt.Errorf("支付方式 %s 未启用", req.Method)
	}
	// 生成订单ID
	orderID, err := s.generateOrderID()
	if err != nil {
		return nil, fmt.Errorf("生成订单ID失败: %v", err)
	}

	// 创建订单记录
	order := &model.OrderModel{
		OrderID: orderID, UserID: *req.UserId,
		PayPlan: req.PayPlan, Amount: *req.Amount,
		Method: req.Method, Status: model.PaymentPending,
		ExpiredAt: time.Now().Add(30 * time.Minute), // 30分钟过期
	}
	// 保存订单到数据库
	if err = s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("保存订单失败: %v", err)
	}
	// 支付订单
	provider, err := payment.NewPayment(req.Method)
	if err != nil {
		return nil, fmt.Errorf("创建支付实例失败: %v", err)
	}
	if err = provider.Create(order); err != nil {
		return nil, fmt.Errorf("创建支付订单失败: %v", err)
	}
	if err = s.db.Save(order).Error; err != nil {
		return nil, fmt.Errorf("保存订单失败: %v", err)
	}
	return order, nil
}

// QueryOrder 获取支付订单
func (s *OrderService) QueryOrder(orderID string) (*model.OrderModel, error) {
	var order model.OrderModel
	if err := s.db.Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) QueryPayment(order *model.OrderModel) error {
	if order.Status != model.PaymentPending {
		return nil
	}
	provider, err := payment.NewPayment(order.Method)
	if err != nil {
		return fmt.Errorf("创建支付实例失败: %v", err)
	}
	if err = provider.Query(order); err != nil {
		return fmt.Errorf("查询订单状态失败: %v", err)
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
	order, err := s.QueryOrder(orderId)
	if err != nil {
		return fmt.Errorf("查询订单失败: %v", err)
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
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 更新用户套餐
	var updateUser = map[string]any{
		"user_plan": order.PayPlan, "api_limit": limit,
		"expire_at": time.Now().AddDate(0, 0, limit.ExpireDays),
	}
	var query = tx.Model(&model.UserModel{}).Where("id = ?", order.UserID)
	if err := query.Updates(updateUser).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户套餐失败: %v", err)
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
