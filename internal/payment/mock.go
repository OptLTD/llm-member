package payment

import (
	"log"
	"time"

	"llm-member/internal/model"
)

type MockPayment struct {
}

// Create 创建模拟支付订单，直接标记为成功
func (m *MockPayment) Create(order *model.OrderModel) error {
	// 模拟显示支付二维码
	log.Printf("[mock][%s] payment created successfully", order.OrderID)
	order.Status = model.PaymentPending
	order.PayURL = "mock://payment/success"
	order.QRCode = "/api/order/qrcode/" + order.OrderID
	return nil
}

// Query 查询模拟支付状态，始终返回成功
func (m *MockPayment) Query(order *model.OrderModel) error {
	// 模拟查询成功, 手机扫码之后
	log.Printf("[mock][%s] querying payment status", order.OrderID)
	if order.Status == model.PaymentPending {
		order.Status = model.PaymentSucceed
		now := time.Now()
		order.PaidAt = &now
		log.Printf("[mock][%s] payment successful", order.OrderID)
	}
	return nil
}

// Close 关闭模拟支付订单
func (m *MockPayment) Close(order *model.OrderModel) error {
	// 模拟关闭成功
	log.Printf("[mock][%s] payment closed successfully", order.OrderID)
	order.Status = model.PaymentCanceled
	return nil
}

// Refund 模拟退款，直接标记为成功
func (m *MockPayment) Refund(order *model.OrderModel) error {
	// 模拟退款成功
	log.Printf("[mock][%s] refund processed successfully", order.OrderID)
	order.Status = model.PaymentRefunded
	return nil
}
