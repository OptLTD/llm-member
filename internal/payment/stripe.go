package payment

import (
	"fmt"
	"log"

	"llm-member/internal/config"
	"llm-member/internal/model"
)

// StripePayment Stripe支付实现
type StripePayment struct {
	config *config.PaymentProvider
}

// NewStripePayment 创建Stripe支付实例
func NewStripePayment() (IPayment, error) {
	provider := config.GetPaymentProvider("stripe")
	if provider == nil {
		return nil, fmt.Errorf("stripe payment provider not configured")
	}

	return &StripePayment{
		config: provider,
	}, nil
}

// Create 创建Stripe支付订单
func (s *StripePayment) Create(order *model.OrderModel) error {
	// 根据套餐类型设置价格（以美元为单位，Stripe主要支持美元）
	var amount float64
	switch order.PayPlan {
	case model.PlanBasic:
		amount = 14.99 // $14.99
	case model.PlanExtra:
		amount = 42.99 // $42.99
	case model.PlanUltra:
		amount = 142.99 // $142.99
	case model.PlanSuper:
		amount = 285.99 // $285.99
	default:
		// 自定义金额，从订单中获取并转换为美元（假设汇率1:7）
		if order.Amount > 0 {
			amount = order.Amount / 7.0 // 简单汇率转换
		} else {
			amount = 1.99 // 默认$1.99
		}
	}

	// 生成Stripe支付URL（简化实现）
	order.PayURL = fmt.Sprintf("https://checkout.stripe.com/pay/%s?amount=%.2f&currency=usd", order.OrderID, amount)
	order.Status = model.PaymentPending

	log.Printf("[stripe][%s] payment created with amount $%.2f", order.OrderID, amount)
	return nil
}

// Query 查询Stripe支付状态
func (s *StripePayment) Query(order *model.OrderModel) error {
	// 简化实现：模拟查询支付状态
	// 在实际项目中，这里应该调用Stripe API查询支付状态
	log.Printf("[stripe][%s] querying payment status", order.OrderID)

	// 这里可以根据实际需要实现Stripe API调用
	// 暂时保持订单状态不变
	return nil
}

// Close 关闭Stripe支付订单
func (s *StripePayment) Close(order *model.OrderModel) error {
	// 简化实现：记录关闭操作
	log.Printf("[stripe][%s] order closed", order.OrderID)
	return nil
}

// Refund 处理Stripe退款
func (s *StripePayment) Refund(order *model.OrderModel) error {
	// 简化实现：记录退款操作
	log.Printf("[stripe][%s] refund requested", order.OrderID)
	return nil
}
