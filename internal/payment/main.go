package payment

import (
	"fmt"
	"llm-member/internal/model"
	"net/http"
)

// Event 支付回调事件
type Event struct {
	Type    string  `json:"type"`      // 事件类型
	OrderID string  `json:"order_id"`  // 订单ID
	Status  string  `json:"status"`    // 支付状态
	Amount  float64 `json:"amount"`    // 金额
	Time    int64   `json:"timestamp"` // 时间戳

	Data object `json:"data"` // 原始数据
}

type IPayment interface {
	Create(order *model.OrderModel) error
	Close(order *model.OrderModel) error
	Query(order *model.OrderModel) error
	Refund(order *model.OrderModel) error
	Webhook(req *http.Request) (*Event, error)
}

func NewPayment(method model.PaymentMethod) (IPayment, error) {
	switch method {
	case "paypal":
		return NewPaypalPayment()
	case "alipay":
		return NewAlipayPayment()
	case "wechat":
		return NewWechatPayment()
	case "creem":
		return NewCreemPayment()
	case "stripe":
		return NewStripePayment()
	case "mock":
		return &MockPayment{}, nil
	default:
		return nil, fmt.Errorf("payment method %s not supported", method)
	}
}
