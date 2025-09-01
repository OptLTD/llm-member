package payment

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/paypal"
)

// PaypalPayment PayPal支付实现
type PaypalPayment struct {
	client *paypal.Client
	config *config.PaymentProvider
}

// NewPaypalPayment 创建PayPal支付实例
func NewPaypalPayment() (*PaypalPayment, error) {
	provider := config.GetPaymentProvider("paypal")
	if provider == nil {
		return nil, consts.ErrPaymentProviderNotConfigured
	}

	// 创建PayPal客户端
	client, err := paypal.NewClient(provider.AppID, provider.Token, false) // false表示沙箱环境
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrPaymentClientCreationFailed, err)
	}

	return &PaypalPayment{
		client: client,
		config: provider,
	}, nil
}

// Create 创建PayPal支付订单
func (p *PaypalPayment) Create(order *model.OrderModel) error {
	// 根据套餐类型设置价格
	var amount float64
	switch order.PayPlan {
	case model.PlanBasic:
		amount = 1.99 // PayPal使用美元
	case model.PlanExtra:
		amount = 2.99
	case model.PlanUltra:
		amount = 5.99
	case model.PlanSuper:
		amount = 14.99
	default:
		amount = 1.99
	}

	// 创建支付请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("intent", "CAPTURE")
	bodyMap.Set("purchase_units", []map[string]interface{}{
		{
			"reference_id": order.OrderID,
			"amount": map[string]interface{}{
				"currency_code": "USD",
				"value":         fmt.Sprintf("%.2f", amount),
			},
			"description": fmt.Sprintf("%s Plan", order.PayPlan),
		},
	})
	bodyMap.Set("application_context", map[string]interface{}{
		"return_url": "http://localhost:8080/payment/return",
		"cancel_url": "http://localhost:8080/payment/cancel",
	})

	// 发起支付请求
	result, err := p.client.CreateOrder(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[paypal][%s] failed to create order: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentCreationFailed, err)
	}

	// 获取支付链接
	var paymentURL string
	if result.Response != nil && result.Response.Links != nil {
		for _, link := range result.Response.Links {
			if link.Rel == "approve" {
				paymentURL = link.Href
				break
			}
		}
	}

	// 更新订单状态
	order.PayURL = paymentURL
	order.Amount = amount
	order.Status = model.PaymentPending
	order.CreatedAt = time.Now()

	return nil
}

// Query 查询PayPal支付状态
func (p *PaypalPayment) Query(order *model.OrderModel) error {
	// PayPal通常通过webhook或redirect处理状态更新
	// 这里提供基本的查询实现
	if order.Status == model.PaymentPending {
		// 实际应用中需要调用PayPal API查询订单状态
		log.Printf("[paypal][%s] querying order status", order.OrderID)
	}
	return nil
}

// Close 关闭PayPal支付订单
func (p *PaypalPayment) Close(order *model.OrderModel) error {
	// PayPal订单关闭逻辑
	order.Status = model.PaymentCanceled
	return nil
}

// Refund PayPal退款
func (p *PaypalPayment) Refund(order *model.OrderModel) error {
	// PayPal退款需要通过PayPal后台或API实现
	// 这里提供基本的状态更新
	log.Printf("[paypal][%s] processing refund, amount: %.2f USD", order.OrderID, order.Amount)

	order.Status = "refunded"
	return nil
}

// Webhook PayPal支付回调验证
func (p *PaypalPayment) Webhook(req *http.Request) (*Event, error) {
	// TODO: 实现PayPal webhook验证逻辑
	log.Printf("[paypal] webhook verification not implemented")
	return nil, consts.ErrPaymentWebhookNotImplemented
}
