package payment

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"
)

type MockPayment struct {
}

// Create 创建模拟支付订单，直接标记为成功
func (m *MockPayment) Create(order *model.OrderModel) error {
	log.Printf("[mock][%s] payment created successfully", order.OrderID)
	order.Status = model.PaymentPending
	order.PayURL = "mock://payment/success"
	order.QRCode = "/api/order/qrcode/" + order.OrderID
	return nil
}

// Query 查询模拟支付状态，始终返回成功
func (m *MockPayment) Query(order *model.OrderModel) error {
	log.Printf("[mock][%s] querying payment status", order.OrderID)
	if order.Status == model.PaymentPending {
		order.Status = model.PaymentSucceed
		now := time.Now()
		order.SucceedAt = &now
		log.Printf("[mock][%s] payment successful", order.OrderID)
	}
	return nil
}

// Close 关闭模拟支付订单
func (m *MockPayment) Close(order *model.OrderModel) error {
	log.Printf("[mock][%s] payment closed successfully", order.OrderID)
	order.Status = model.PaymentCanceled
	return nil
}

// Refund 模拟退款，直接标记为成功
func (m *MockPayment) Refund(order *model.OrderModel) error {
	log.Printf("[mock][%s] refund successful", order.OrderID)
	order.Status = model.PaymentRefunded
	return nil
}

// Webhook 模拟支付回调：仅 debug/test 模式且校验密钥
func (m *MockPayment) Webhook(req *http.Request) (*Event, error) {
	mode := os.Getenv("APP_MODE")
	if mode != "debug" && mode != "test" {
		return nil, fmt.Errorf("%w: mock webhook disabled in production", consts.ErrPaymentWebhookError)
	}
	secret := os.Getenv("MOCK_WEBHOOK_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("%w: MOCK_WEBHOOK_SECRET not configured", consts.ErrWebhookSecretNotConfigured)
	}
	got := req.Header.Get("X-Mock-Webhook-Secret")
	if got == "" {
		got = req.URL.Query().Get("secret")
	}
	if got != secret {
		return nil, fmt.Errorf("%w: invalid mock webhook secret", consts.ErrWebhookSignatureVerificationFailed)
	}

	log.Printf("[mock] webhook verification successful")
	orderID := req.URL.Query().Get("order_id")
	if orderID == "" {
		orderID = req.Header.Get("X-Order-Id")
	}
	if orderID == "" {
		return nil, consts.ErrPaymentWebhookMissingParams
	}

	return &Event{
		Type:    "payment.success",
		OrderID: orderID,
		Status:  "success",
		Data:    make(object),
		Time:    time.Now().Unix(),
	}, nil
}
