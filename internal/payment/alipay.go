package payment

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

// AlipayPayment 支付宝支付实现
type AlipayPayment struct {
	client *alipay.Client
	config *config.PaymentProvider
}

// NewAlipayPayment 创建支付宝支付实例
func NewAlipayPayment() (*AlipayPayment, error) {
	provider := config.GetPaymentProvider("alipay")
	if provider == nil {
		return nil, fmt.Errorf("alipay payment provider not configured")
	}

	// 创建支付宝客户端
	client, err := alipay.NewClient(provider.AppID, provider.Token, false) // false表示沙箱环境
	if err != nil {
		return nil, fmt.Errorf("failed to create alipay client: %v", err)
	}

	// 设置支付宝公钥证书和根证书（如果有的话）
	// client.SetCertSnFromPath("appCertPublicKey.crt", "alipayRootCert.crt", "alipayCertPublicKey_RSA2.crt")

	return &AlipayPayment{
		client: client,
		config: provider,
	}, nil
}

// Create 创建支付订单
func (p *AlipayPayment) Create(order *model.OrderModel) error {
	// 生成订单ID
	orderID := fmt.Sprintf("alipay_%d", time.Now().UnixNano())

	// 创建支付请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("subject", fmt.Sprintf("%s套餐", order.PayPlan))
	bodyMap.Set("out_trade_no", orderID)
	bodyMap.Set("product_code", "QUICK_WAP_WAY")
	bodyMap.Set("total_amount", fmt.Sprintf("%.2f", order.Amount))
	bodyMap.Set("return_url", "http://localhost:8080/payment?status=success")
	bodyMap.Set("notify_url", "http://localhost:8080/api/payment/alipay/notify")

	// 生成支付URL
	payURL, err := p.client.TradeWapPay(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[alipay][%s] failed to create payment: %v", order.OrderID, err)
		return fmt.Errorf("failed to create alipay payment: %v", err)
	}

	log.Printf("[alipay][%s] payment created successfully", order.OrderID)

	order.PayURL = payURL
	order.Method = model.PaymentAlipay
	order.Status = model.PaymentPending
	order.ExpiredAt = time.Now().Add(30 * time.Minute)
	return nil
}

// Query 查询支付状态
func (p *AlipayPayment) Query(order *model.OrderModel) error {
	// 创建查询请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("out_trade_no", order.OrderID)

	result, err := p.client.TradeQuery(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[alipay][%s] payment query failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to query alipay payment: %v", err)
	}

	// 更新订单状态
	if result.Response != nil {
		switch result.Response.TradeStatus {
		case "TRADE_SUCCESS", "TRADE_FINISHED":
			order.Status = model.PaymentSucceed
			log.Printf("[alipay][%s] payment successful", order.OrderID)
			if order.SucceedAt == nil {
				now := time.Now()
				order.SucceedAt = &now
			}
		case "TRADE_CLOSED":
			order.Status = model.PaymentCanceled
			log.Printf("[alipay][%s] payment closed", order.OrderID)
		case "WAIT_BUYER_PAY":
			order.Status = model.PaymentPending
			log.Printf("[alipay][%s] payment pending", order.OrderID)
		default:
			order.Status = model.PaymentCanceled
			log.Printf("[alipay][%s] payment failed, status: %s", order.OrderID, result.Response.TradeStatus)
		}
	}

	return nil
}

// Close 关闭支付订单
func (p *AlipayPayment) Close(order *model.OrderModel) error {
	// 创建关闭请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("out_trade_no", order.OrderID)

	_, err := p.client.TradeClose(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[alipay][%s] payment close failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to close alipay payment: %v", err)
	}

	log.Printf("[alipay][%s] payment closed successfully", order.OrderID)
	order.Status = model.PaymentCanceled
	return nil
}

// Refund 退款
func (p *AlipayPayment) Refund(order *model.OrderModel) error {
	// 创建退款请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("out_trade_no", order.OrderID)
	bodyMap.Set("refund_amount", strconv.FormatFloat(order.Amount, 'f', 2, 64))
	bodyMap.Set("refund_reason", "用户申请退款")

	_, err := p.client.TradeRefund(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[alipay][%s] refund failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to refund alipay payment: %v", err)
	}

	log.Printf("[alipay][%s] refund processed successfully", order.OrderID)
	order.Status = model.PaymentRefunded
	return nil
}

// Webhook 支付宝支付回调验证
func (p *AlipayPayment) Webhook(req *http.Request) (*Event, error) {
	// TODO: 实现支付宝webhook验证逻辑
	log.Printf("[alipay] webhook verification not implemented")
	return nil, fmt.Errorf("alipay webhook verification not implemented")
}
