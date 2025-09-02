package payment

import (
	"llm-member/internal/consts"
	"llm-member/internal/model"
	"net/http"
)

// 定义一个通用的map类型别名
type object = map[string]any

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

// UnsupportedPayment 不支持的支付方式实现
type UnsupportedPayment struct {
	method model.PaymentMethod
}

func (u *UnsupportedPayment) Create(order *model.OrderModel) error {
	return consts.ErrPaymentMethodNotSupported
}

func (u *UnsupportedPayment) Close(order *model.OrderModel) error {
	return consts.ErrPaymentMethodNotSupported
}

func (u *UnsupportedPayment) Query(order *model.OrderModel) error {
	return consts.ErrPaymentMethodNotSupported
}

func (u *UnsupportedPayment) Refund(order *model.OrderModel) error {
	return consts.ErrPaymentMethodNotSupported
}

func (u *UnsupportedPayment) Webhook(req *http.Request) (*Event, error) {
	return nil, consts.ErrPaymentMethodNotSupported
}

func NewPayment(method model.PaymentMethod) IPayment {
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
		return &MockPayment{}
	default:
		// 返回一个错误实现，在调用接口时返回不支持的支付方式错误
		return &UnsupportedPayment{method: method}
	}
}
