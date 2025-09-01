package payment

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/go-pay/xlog"
)

// WechatPayment 微信支付实现
type WechatPayment struct {
	client *wechat.ClientV3
	config *config.WechatProvider
}

// NewWechatPayment 创建微信支付实例
func NewWechatPayment() (*WechatPayment, error) {
	provider := config.GetWechatProvider()
	if provider == nil {
		return nil, fmt.Errorf("wechat payment provider not configured")
	}

	// 验证必要的配置参数
	if provider.MchID == "" || provider.SerialNo == "" || provider.APIv3Key == "" || provider.PrivateKey == "" {
		return nil, fmt.Errorf("wechat payment v3 configuration incomplete")
	}

	// 创建微信支付v3客户端
	client, err := wechat.NewClientV3(provider.MchID, provider.SerialNo, provider.APIv3Key, provider.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create wechat v3 client: %v", err)
	}

	// 启用自动验签（使用平台证书自动获取）
	err = client.AutoVerifySign()
	if err != nil {
		xlog.Warn("Failed to enable auto verify sign:", err)
	}

	// 开启调试模式（可选）
	client.DebugSwitch = gopay.DebugOff

	return &WechatPayment{
		client: client,
		config: provider,
	}, nil
}

// Create 创建微信支付订单
func (w *WechatPayment) Create(order *model.OrderModel) error {
	// 根据套餐类型设置价格（微信支付使用分为单位）
	var amount int
	switch order.PayPlan {
	case model.PlanBasic:
		amount = 990 // 9.9元 = 990分
	case model.PlanExtra:
		amount = 1990 // 19.9元 = 1990分
	case model.PlanUltra:
		amount = 3990 // 39.9元 = 3990分
	case model.PlanSuper:
		amount = 9990 // 99.9元 = 9990分
	default:
		amount = 990
	}

	// 设置订单过期时间（10分钟后）
	expire := time.Now().Add(10 * time.Minute).Format(time.RFC3339)

	// 创建支付请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("appid", w.config.AppID).
		Set("mchid", w.config.MchID).
		Set("description", fmt.Sprintf("%s套餐", order.PayPlan)).
		Set("out_trade_no", order.OrderID).
		Set("time_expire", expire).
		Set("notify_url", w.config.NotifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", amount).
				Set("currency", "CNY")
		})

	// 调用微信支付v3 JSAPI下单接口
	wxRsp, err := w.client.V3TransactionJsapi(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[wechat][%s] payment creation failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to create wechat payment: %v", err)
	}

	// 检查响应状态
	if wxRsp.Code != 200 {
		log.Printf("[wechat][%s] payment creation error: %s", order.OrderID, wxRsp.Error)
		return fmt.Errorf("wechat payment error: %s", wxRsp.Error)
	}

	log.Printf("[wechat][%s] payment created successfully, prepay_id: %s", order.OrderID, wxRsp.Response.PrepayId)

	// 生成JSAPI支付签名
	jsapi, err := w.client.PaySignOfJSAPI(w.config.AppID, wxRsp.Response.PrepayId)
	if err != nil {
		log.Printf("[wechat][%s] failed to generate JSAPI pay sign: %v", order.OrderID, err)
		return fmt.Errorf("failed to generate pay sign: %v", err)
	}

	// 更新订单状态
	order.PayURL = fmt.Sprintf("jsapi://%s", jsapi.Package) // 存储支付包信息
	order.Status = model.PaymentPending
	order.Amount = float64(amount) / 100 // 转换为元
	order.CreatedAt = time.Now()

	log.Printf("[wechat][%s] payment order created, amount: %.2f yuan", order.OrderID, order.Amount)
	return nil
}

// Query 查询微信支付状态
func (w *WechatPayment) Query(order *model.OrderModel) error {
	// 调用微信支付v3查询订单接口
	wxRsp, err := w.client.V3TransactionQueryOrder(
		context.Background(), wechat.OutTradeNo, order.OrderID,
	)
	if err != nil {
		log.Printf("[wechat][%s] payment query failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to query wechat payment: %v", err)
	}

	// 检查响应状态
	if wxRsp.Code != 200 {
		log.Printf("[wechat][%s] payment query error: %s", order.OrderID, wxRsp.Error)
		return fmt.Errorf("wechat payment query error: %s", wxRsp.Error)
	}

	// 根据微信返回的交易状态更新订单状态
	switch wxRsp.Response.TradeState {
	case "SUCCESS":
		order.Status = model.PaymentSucceed
		log.Printf("[wechat][%s] payment successful", order.OrderID)
	case "REFUND":
		order.Status = model.PaymentRefunded
		log.Printf("[wechat][%s] payment refunded", order.OrderID)
	case "NOTPAY":
		order.Status = model.PaymentPending
		log.Printf("[wechat][%s] payment pending", order.OrderID)
	case "CLOSED":
		order.Status = model.PaymentCanceled
		log.Printf("[wechat][%s] payment closed", order.OrderID)
	case "REVOKED":
		order.Status = model.PaymentCanceled
		log.Printf("[wechat][%s] payment revoked", order.OrderID)
	case "USERPAYING":
		order.Status = model.PaymentPending
		log.Printf("[wechat][%s] payment user paying", order.OrderID)
	case "PAYERROR":
		order.Status = model.PaymentCanceled
		log.Printf("[wechat][%s] payment error", order.OrderID)
	default:
		log.Printf("[wechat][%s] unknown payment status: %s", order.OrderID, wxRsp.Response.TradeState)
	}

	return nil
}

// Close 关闭微信支付订单
func (w *WechatPayment) Close(order *model.OrderModel) error {
	// 创建关闭请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("mchid", w.config.MchID)

	// 调用微信支付v3关闭订单接口
	wxRsp, err := w.client.V3TransactionCloseOrder(context.Background(), order.OrderID)
	if err != nil {
		log.Printf("[wechat][%s] payment close failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to close wechat payment: %v", err)
	}

	// 检查响应状态
	if wxRsp.Code != 204 { // 关闭订单成功返回204
		log.Printf("[wechat][%s] payment close error: %s", order.OrderID, wxRsp.Error)
		return fmt.Errorf("wechat payment close error: %s", wxRsp.Error)
	}

	log.Printf("[wechat][%s] payment closed successfully", order.OrderID)
	order.Status = model.PaymentCanceled
	return nil
}

// Refund 微信支付退款
func (w *WechatPayment) Refund(order *model.OrderModel) error {
	// 创建退款请求参数
	bodyMap := make(gopay.BodyMap)
	bodyMap.Set("out_trade_no", order.OrderID).
		Set("out_refund_no", "refund_"+order.OrderID).
		Set("reason", "用户申请退款").
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			// 转换为分
			bm.Set("refund", int(order.Amount*100)).
				Set("total", int(order.Amount*100)).
				Set("currency", "CNY")
		})

	// 调用微信支付v3退款接口
	wxRsp, err := w.client.V3Refund(context.Background(), bodyMap)
	if err != nil {
		log.Printf("[wechat][%s] refund failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to process wechat refund: %v", err)
	}

	// 检查响应状态
	if wxRsp.Code != 200 {
		log.Printf("[wechat][%s] refund error: %s", order.OrderID, wxRsp.Error)
		return fmt.Errorf("wechat refund error: %s", wxRsp.Error)
	}

	log.Printf("[wechat][%s] refund processed successfully, refund_id: %s", order.OrderID, wxRsp.Response.RefundId)
	order.Status = "refunded"
	return nil
}

// Webhook 微信支付回调验证
func (w *WechatPayment) Webhook(req *http.Request) (*Event, error) {
	// TODO: 实现微信支付webhook验证逻辑
	log.Printf("[wechat] webhook verification not implemented")
	return nil, fmt.Errorf("wechat webhook verification not implemented")
}
