package payment

import (
	"fmt"
	"llm-member/internal/model"
)

type IPayment interface {
	Create(order *model.OrderModel) error
	Close(order *model.OrderModel) error
	Query(order *model.OrderModel) error
	Refund(order *model.OrderModel) error
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
