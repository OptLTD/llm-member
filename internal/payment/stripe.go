package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"
)

// StripePayment Stripe支付实现
type StripePayment struct {
	config *config.Stripe
	client *http.Client

	baseURL string // Stripe API基础URL
}

type StripeCheckoutSession struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	URL    string `json:"url"`
	Mode   string `json:"mode"`
	Status string `json:"status"`

	Currency string `json:"currency"`
	Customer string `json:"customer"`
	Metadata object `json:"metadata"`
	Created  int64  `json:"created"`

	AmountTotal int64  `json:"amount_total"`
	ExpiresAt   int64  `json:"expires_at"`
	SuccessURL  string `json:"success_url,omitempty"`
	CancelURL   string `json:"cancel_url,omitempty"`

	PaymentStatus string `json:"payment_status"`
	PaymentIntent string `json:"payment_intent"`

	LineItems *StripeLineItemList `json:"line_items,omitempty"`
}

type StripeLineItemList struct {
	Object string `json:"object"`

	Data []StripeLineItem `json:"data"`
}

type StripeLineItem struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Quantity    int64  `json:"quantity"`
	AmountTotal int64  `json:"amount_total"`
	Currency    string `json:"currency"`
	Description string `json:"description"`

	Price StripePrice `json:"price"`
}

// StripePrice Stripe Price对象
type StripePrice struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Object string `json:"object"`
	Active bool   `json:"active"`

	Currency   string `json:"currency"`
	UnitAmount int64  `json:"unit_amount"`
	Recurring  *bool  `json:"recurring"`
}

// StripeCreateSessionRequest 创建Checkout Session请求
type StripeCreateSessionRequest struct {
	Mode string `json:"mode"`

	SuccessURL string `json:"success_url,omitempty"`
	CancelURL  string `json:"cancel_url,omitempty"`
	Metadata   object `json:"metadata,omitempty"`
	// 客户创建设置
	CustomerCreation string `json:"customer_creation,omitempty"`
	// 支付方式类型
	PaymentMethodTypes []string `json:"payment_method_types,omitempty"`
	// 账单地址收集设置
	BillingAddressCollection string `json:"billing_address_collection,omitempty"`
	// 客户设置
	Customer *StripeCustomerData `json:"customer,omitempty"`
	// 订单项
	LineItems []StripeCreateLineItem `json:"line_items"`
	// 自动税收设置
	AutomaticTax *StripeAutomaticTax `json:"automatic_tax,omitempty"`
	// 过期时间（Unix时间戳）
	ExpiresAt int64 `json:"expires_at,omitempty"`
	// 客户端引用ID
	ClientReferenceID string `json:"client_reference_id,omitempty"`
}

// StripeCreateLineItem 创建Line Item
type StripeCreateLineItem struct {
	Price string `json:"price,omitempty"`

	Quantity int64 `json:"quantity"`

	PriceData *StripePriceData `json:"price_data,omitempty"`
}

// StripePriceData 价格数据
type StripePriceData struct {
	Currency   string `json:"currency"`
	UnitAmount int64  `json:"unit_amount"`

	ProductData StripeProductData `json:"product_data"`
}

// StripeProductData 产品数据
type StripeProductData struct {
	Name string `json:"name"`

	Description string `json:"description,omitempty"`
}

// StripeCustomerData 客户数据
type StripeCustomerData struct {
	Enabled bool `json:"enabled"`
}

// StripeAutomaticTax 自动税收设置
type StripeAutomaticTax struct {
	Enabled bool `json:"enabled"`
}

// StripeRefund Stripe退款对象
type StripeRefund struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Reason   string `json:"reason"`
	Status   string `json:"status"`
	Metadata object `json:"metadata"`
	Created  int64  `json:"created"`

	PaymentIntent string `json:"payment_intent"`
}

// StripeRefundRequest Stripe退款请求
type StripeRefundRequest struct {
	PaymentIntent string `json:"payment_intent"`

	Amount   int64  `json:"amount,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Metadata object `json:"metadata,omitempty"`
}

// StripeRefundResponse Stripe退款响应
type StripeRefundResponse struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Reason   string `json:"reason"`
	Status   string `json:"status"`
	Metadata object `json:"metadata"`
	Created  int64  `json:"created"`

	PaymentIntent string `json:"payment_intent"`
}

// StripeWebhookEvent Stripe Webhook事件
type StripeWebhookEvent struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Type    string `json:"type"`
	Created int64  `json:"created"`

	Data StripeEventData `json:"data"`
}

// StripeEventData Stripe事件数据
type StripeEventData struct {
	Object object `json:"object"`
}

// NewStripePayment 创建Stripe支付实例
func NewStripePayment() *StripePayment {
	baseURL := "https://api.stripe.com/v1"
	return &StripePayment{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ensureClientReady 确保客户端已准备就绪
func (s *StripePayment) ensureClientReady() error {
	if s.config != nil {
		return nil
	}

	// 获取配置
	provider := config.GetStripeConfig()
	if provider == nil {
		return consts.ErrPaymentProviderNotConfigured
	}

	s.config = provider
	return nil
}

func (s *StripePayment) getLineItems(order *model.OrderModel) []StripeCreateLineItem {

	var plan_name string
	var product_id string
	switch order.PayPlan {
	case model.PlanBasic:
		plan_name = "Basic Plan"
		product_id = s.config.PlanBasicId
	case model.PlanExtra:
		plan_name = "Extra Plan"
		product_id = s.config.PlanExtraId
	case model.PlanUltra:
		plan_name = "Ultra Plan"
		product_id = s.config.PlanUltraId
	case model.PlanSuper:
		plan_name = "Super Plan"
		product_id = s.config.PlanSuperId
	}
	if product_id != "" {
		return []StripeCreateLineItem{{
			Quantity: 1, Price: product_id,
		}}
	}

	priceData := &StripePriceData{
		Currency: "usd", UnitAmount: int64(order.Amount) * 100,
		ProductData: StripeProductData{
			Name: plan_name, Description: fmt.Sprintf(
				"LLM Member %s - Order %s", plan_name, order.OrderID,
			),
		},
	}
	return []StripeCreateLineItem{{
		Quantity: 1, PriceData: priceData,
	}}
}

// Create 创建Stripe支付订单
func (s *StripePayment) Create(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return err
	}

	// priceData := &StripePriceData{
	// 	Currency: "usd", UnitAmount: int64(order.Amount) * 100,
	// 	ProductData: StripeProductData{
	// 		Name: planName, Description: fmt.Sprintf(
	// 			"LLM Member %s - Order %s", planName, order.OrderID,
	// 		),
	// 	},
	// }

	// 构建Checkout Session请求
	request := StripeCreateSessionRequest{
		Mode: "payment", LineItems: s.getLineItems(order),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		// 设置支付方式类型
		PaymentMethodTypes: []string{"card", "alipay"},
		ClientReferenceID:  order.OrderID,
		Metadata: object{
			"order_id": order.OrderID,
			"user_id":  order.UserID,
			"pay_plan": string(order.PayPlan),
		},
		SuccessURL: fmt.Sprintf(
			"http://%s/success?order_id=%s",
			"localhost", order.OrderID,
		),
		// 设置客户创建
		// CustomerCreation: "if_required",
		// 账单收集
		// BillingAddressCollection: "auto",
	}

	// 发送HTTP请求到Stripe API
	resp, err := s.makeAPIRequest("POST", "/checkout/sessions", request)
	if err != nil {
		log.Printf("[stripe][%s] payment creation failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentCreationFailed, err)
	}

	var sessionResp StripeCheckoutSession
	if err := json.Unmarshal(resp, &sessionResp); err != nil {
		log.Printf("[stripe][%s] failed to parse response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 更新订单信息
	order.PayURL = sessionResp.URL
	order.ThridID = sessionResp.ID
	order.Status = model.PaymentPending
	log.Printf(
		"[stripe][%s][%s] payment created successfully, checkout URL: %s",
		order.OrderID, order.ThridID, sessionResp.URL,
	)
	return nil
}

// Query 查询Stripe支付状态
func (s *StripePayment) Query(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return err
	}

	// 使用checkout session ID查询支付状态
	if order.ThridID == "" {
		log.Printf("[stripe][%s] no checkout session ID found", order.OrderID)
		return fmt.Errorf("%w: missing checkout session ID", consts.ErrPaymentQueryFailed)
	}

	var sessionResp StripeCheckoutSession
	url := fmt.Sprintf("/checkout/sessions/%s", order.ThridID)
	if resp, err := s.makeAPIRequest("GET", url, nil); err != nil {
		log.Printf("[stripe][%s] payment query failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentQueryFailed, err)
	} else if err := json.Unmarshal(resp, &sessionResp); err != nil {
		log.Printf("[stripe][%s] failed to parse query response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 根据Stripe返回的状态更新订单状态
	// 优先使用payment_status，如果不存在则使用session status
	// Stripe Checkout Session状态: open, complete, expired
	// Stripe Payment状态: unpaid, paid, no_payment_required
	var sessionStatus = sessionResp.Status
	var paymentStatus = sessionResp.PaymentStatus

	log.Printf("[stripe][%s] session status: %s, payment status: %s", order.OrderID, sessionStatus, paymentStatus)

	// 根据状态优先级进行判断
	switch {
	case paymentStatus == "paid" || sessionStatus == "complete":
		// 支付成功
		order.Status = model.PaymentSucceed
		log.Printf("[stripe][%s] payment successful", order.OrderID)
		if order.SucceedAt == nil {
			now := time.Now()
			order.SucceedAt = &now
		}

	case sessionStatus == "expired":
		// 会话过期
		order.Status = model.PaymentCanceled
		log.Printf("[stripe][%s] checkout session expired", order.OrderID)

	case paymentStatus == "unpaid" || sessionStatus == "open":
		// 支付待处理
		order.Status = model.PaymentPending
		log.Printf("[stripe][%s] payment pending", order.OrderID)

	case paymentStatus == "no_payment_required":
		// 无需支付（免费订单）
		order.Status = model.PaymentSucceed
		log.Printf("[stripe][%s] no payment required", order.OrderID)
		if order.SucceedAt == nil {
			now := time.Now()
			order.SucceedAt = &now
		}

	default:
		// 未知状态，记录详细信息但不更改订单状态
		log.Printf("[stripe][%s] unknown status combination - session: %s, payment: %s",
			order.OrderID, sessionStatus, paymentStatus)
		return fmt.Errorf("unknown payment status: session=%s, payment=%s", sessionStatus, paymentStatus)
	}

	// 更新订单的第三方ID和金额信息
	if sessionResp.AmountTotal > 0 {
		order.Amount = float64(sessionResp.AmountTotal) / 100 // Stripe使用分为单位
	}

	// 如果支付成功，尝试获取line items详细信息
	if order.Status == model.PaymentSucceed {
		if lineItems, err := s.GetLineItems(order.ThridID); err != nil {
			log.Printf("[stripe][%s] failed to fetch line items: %v", order.OrderID, err)
			// 不影响主要流程，继续执行
		} else if len(lineItems.Data) > 0 {
			// 记录line items信息用于调试和审计
			for i, item := range lineItems.Data {
				log.Printf("[stripe][%s] line item %d: %s, amount: %d %s, quantity: %d",
					order.OrderID, i+1, item.Description, item.AmountTotal, item.Currency, item.Quantity)
			}
		}
	}

	return nil
}

// Close 关闭Stripe支付订单
func (s *StripePayment) Close(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return err
	}

	// 检查订单状态是否可以关闭
	if order.Status != model.PaymentPending {
		log.Printf("[stripe][%s] order cannot be closed, current status: %s", order.OrderID, order.Status)
		return fmt.Errorf("%w: order status is %s", consts.ErrPaymentCloseFailed, order.Status)
	}

	if order.ThridID == "" {
		log.Printf("[stripe][%s] no checkout session ID found for closure", order.OrderID)
		return fmt.Errorf("%w: missing checkout session ID", consts.ErrPaymentCloseFailed)
	}

	// 使用Stripe Checkout Session过期API
	// POST /v1/checkout/sessions/{session}/expire
	url := fmt.Sprintf("/checkout/sessions/%s/expire", order.ThridID)

	log.Printf("[stripe][%s] attempting to expire checkout session: %s", order.OrderID, order.ThridID)

	if resp, err := s.makeAPIRequest("POST", url, nil); err != nil {
		// 记录详细错误信息
		log.Printf("[stripe][%s] failed to expire checkout session: %v", order.OrderID, err)

		// 尝试查询当前状态，如果已经过期则直接标记为取消
		if queryErr := s.Query(order); queryErr == nil {
			if order.Status == model.PaymentCanceled {
				log.Printf("[stripe][%s] session already expired", order.OrderID)
				return nil
			}
		}

		// 如果查询也失败，返回原始错误
		return fmt.Errorf("%w: %v", consts.ErrPaymentCloseFailed, err)
	} else {
		// 解析响应以确认过期成功
		var expiredSession StripeCheckoutSession
		if err := json.Unmarshal(resp, &expiredSession); err != nil {
			log.Printf("[stripe][%s] failed to parse expire response: %v", order.OrderID, err)
			// 即使解析失败，也认为过期成功
		} else {
			log.Printf("[stripe][%s] session expired successfully, status: %s", order.OrderID, expiredSession.Status)
		}

		order.Status = model.PaymentCanceled
		log.Printf("[stripe][%s] payment session closed successfully", order.OrderID)
		return nil
	}
}

// Refund 处理Stripe退款
func (s *StripePayment) Refund(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return err
	}

	// 检查订单是否可以退款
	if order.Status != model.PaymentSucceed {
		log.Printf("[stripe][%s] order cannot be refunded, current status: %s", order.OrderID, order.Status)
		return fmt.Errorf("%w: order status is %s", consts.ErrOrderCannotBeRefunded, order.Status)
	}

	if order.ThridID == "" {
		log.Printf("[stripe][%s] no payment intent ID found for refund", order.OrderID)
		return fmt.Errorf("%w: missing payment intent ID", consts.ErrPaymentRefundFailed)
	}

	// 构建退款请求
	refundReq := StripeRefundRequest{
		PaymentIntent: order.ThridID, Reason: "requested_by_customer",
		Metadata: object{"order_id": order.OrderID},
		Amount:   int64(order.Amount * 100), // 转换为分

	}

	// 调用Stripe退款API
	var refundResp StripeRefundResponse
	if resp, err := s.makeAPIRequest("POST", "/refunds", refundReq); err != nil {
		log.Printf("[stripe][%s] refund request failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentRefundFailed, err)
	} else if err := json.Unmarshal(resp, &refundResp); err != nil {
		log.Printf("[stripe][%s] failed to parse refund response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 检查退款状态
	if refundResp.Status == "succeeded" {
		order.Status = model.PaymentRefunded
		log.Printf("[stripe][%s] refund successful, refund_id: %s", order.OrderID, refundResp.ID)
	} else {
		log.Printf("[stripe][%s] refund failed, status: %s", order.OrderID, refundResp.Status)
		return fmt.Errorf("%w: refund status is %s", consts.ErrRefundNotSuccessful, refundResp.Status)
	}

	return nil
}

// Webhook 处理Stripe Webhook回调
func (s *StripePayment) Webhook(req *http.Request) (*Event, error) {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return nil, err
	}
	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("[stripe] failed to read webhook body: %v", err)
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookBodyReadFailed, err)
	}

	// 验证Stripe Webhook签名
	signature := req.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Printf("[stripe] missing stripe signature header")
		return nil, fmt.Errorf("%w: missing stripe signature", consts.ErrWebhookSignatureVerificationFailed)
	}

	// 验证签名
	if err := s.verifyWebhookSignature(body, signature); err != nil {
		log.Printf("[stripe] webhook signature verification failed: %v", err)
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookSignatureVerificationFailed, err)
	}

	// 解析Webhook事件
	var webhookEvent StripeWebhookEvent
	if err := json.Unmarshal(body, &webhookEvent); err != nil {
		log.Printf("[stripe] failed to parse webhook event: %v", err)
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookEventParseFailed, err)
	}

	log.Printf("[stripe] received webhook event: %s, id: %s", webhookEvent.Type, webhookEvent.ID)

	// 处理不同类型的事件
	switch webhookEvent.Type {
	case "checkout.session.completed":
		return s.handleCheckoutSessionCompleted(&webhookEvent)
	case "checkout.session.expired":
		return s.handleCheckoutSessionExpired(&webhookEvent)
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(&webhookEvent)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(&webhookEvent)
	default:
		log.Printf("[stripe] unhandled webhook event type: %s", webhookEvent.Type)
		return nil, nil // 忽略未处理的事件类型
	}
}

// GetLineItems 获取Checkout Session的line items详细信息
func (s *StripePayment) GetLineItems(sessionID string) (*StripeLineItemList, error) {
	// 检查配置和初始化客户端
	if err := s.ensureClientReady(); err != nil {
		return nil, err
	}

	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	// 构建API请求URL
	url := fmt.Sprintf("/checkout/sessions/%s/line_items", sessionID)
	log.Printf("[stripe] fetching line items for session: %s", sessionID)
	resp, err := s.makeAPIRequest("GET", url, nil)
	if err != nil {
		log.Printf("[stripe] failed to fetch line items: %v", err)
		return nil, fmt.Errorf("failed to fetch line items: %w", err)
	}

	var lineItems StripeLineItemList
	if err := json.Unmarshal(resp, &lineItems); err != nil {
		log.Printf("[stripe] failed to parse line items response: %v", err)
		return nil, fmt.Errorf("failed to parse line items: %w", err)
	}

	log.Printf("[stripe] successfully fetched %d line items", len(lineItems.Data))
	return &lineItems, nil
}

// verifyWebhookSignature 验证Stripe Webhook签名
func (s *StripePayment) verifyWebhookSignature(payload []byte, signature string) error {
	// 解析签名头
	// Stripe-Signature: t=1492774577,v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd
	parts := strings.Split(signature, ",")
	var timestamp, v1Signature string

	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			v1Signature = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || v1Signature == "" {
		return fmt.Errorf("invalid signature format")
	}

	// 检查时间戳（防止重放攻击）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %v", err)
	}

	// 检查时间戳是否在5分钟内
	if time.Now().Unix()-ts > 300 {
		return fmt.Errorf("timestamp too old")
	}

	// 构建签名字符串
	signedPayload := timestamp + "." + string(payload)

	// 使用HMAC-SHA256验证签名
	if s.config.WhSecret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	mac := hmac.New(sha256.New, []byte(s.config.WhSecret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(v1Signature), []byte(expectedSignature)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// handleCheckoutSessionCompleted 处理checkout session完成事件
func (s *StripePayment) handleCheckoutSessionCompleted(event *StripeWebhookEvent) (*Event, error) {
	session := event.Data.Object
	if session == nil {
		return nil, fmt.Errorf("invalid session data")
	}

	// 从session metadata中获取订单ID
	metadata, ok := session["metadata"].(object)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	orderID, ok := metadata["order_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order_id in metadata")
	}

	// 获取支付金额
	amountTotal, _ := session["amount_total"].(float64)

	return &Event{
		Type:    "payment.succeeded",
		OrderID: orderID,
		Status:  "succeeded",
		Amount:  amountTotal / 100, // Stripe金额以分为单位，转换为元
		Time:    time.Now().Unix(),
		Data: object{
			"provider":   "stripe",
			"session_id": session["id"],
			"event_id":   event.ID,
		},
	}, nil
}

// handleCheckoutSessionExpired 处理checkout session过期事件
func (s *StripePayment) handleCheckoutSessionExpired(event *StripeWebhookEvent) (*Event, error) {
	session := event.Data.Object
	if session == nil {
		return nil, fmt.Errorf("invalid session data")
	}

	metadata, ok := session["metadata"].(object)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	orderID, ok := metadata["order_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order_id in metadata")
	}

	return &Event{
		Type:    "payment.expired",
		OrderID: orderID,
		Status:  "expired",
		Amount:  0,
		Time:    time.Now().Unix(),
		Data: object{
			"provider":   "stripe",
			"session_id": session["id"],
			"event_id":   event.ID,
		},
	}, nil
}

// handlePaymentIntentSucceeded 处理支付成功事件
func (s *StripePayment) handlePaymentIntentSucceeded(event *StripeWebhookEvent) (*Event, error) {
	paymentIntent := event.Data.Object
	if paymentIntent == nil {
		return nil, fmt.Errorf("invalid payment intent data")
	}

	metadata, ok := paymentIntent["metadata"].(object)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	orderID, ok := metadata["order_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order_id in metadata")
	}

	amount, _ := paymentIntent["amount"].(float64)

	return &Event{
		Type:    "payment.succeeded",
		OrderID: orderID,
		Status:  "succeeded",
		Amount:  amount / 100, // Stripe金额以分为单位，转换为元
		Time:    time.Now().Unix(),
		Data: object{
			"provider":          "stripe",
			"payment_intent_id": paymentIntent["id"],
			"event_id":          event.ID,
		},
	}, nil
}

// handlePaymentIntentFailed 处理支付失败事件
func (s *StripePayment) handlePaymentIntentFailed(event *StripeWebhookEvent) (*Event, error) {
	paymentIntent := event.Data.Object
	if paymentIntent == nil {
		return nil, fmt.Errorf("invalid payment intent data")
	}

	metadata, ok := paymentIntent["metadata"].(object)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	orderID, ok := metadata["order_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order_id in metadata")
	}

	amount, _ := paymentIntent["amount"].(float64)

	return &Event{
		Type:    "payment.failed",
		OrderID: orderID,
		Status:  "failed",
		Amount:  amount / 100, // Stripe金额以分为单位，转换为元
		Time:    time.Now().Unix(),
		Data: object{
			"provider":          "stripe",
			"payment_intent_id": paymentIntent["id"],
			"event_id":          event.ID,
		},
	}, nil
}

// structToURLValues 将结构体转换为URL编码格式
func (s *StripePayment) structToURLValues(data any) (url.Values, error) {
	values := url.Values{}
	if data == nil {
		return values, nil
	}

	// 将结构体先转换为JSON，再转换为map
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var dataMap object
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, err
	}

	// 递归处理嵌套结构
	s.flattenMap(dataMap, "", values)
	return values, nil
}

// flattenMap 递归展平嵌套的map结构
func (s *StripePayment) flattenMap(data object, prefix string, values url.Values) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "[" + key + "]"
		}

		switch v := value.(type) {
		case object:
			s.flattenMap(v, fullKey, values)
		case []any:
			for i, item := range v {
				itemKey := fullKey + "[" + strconv.Itoa(i) + "]"
				if itemMap, ok := item.(object); ok {
					s.flattenMap(itemMap, itemKey, values)
				} else {
					values.Set(itemKey, s.formatValue(item))
				}
			}
		case nil:
			// 跳过nil值
		default:
			values.Set(fullKey, s.formatValue(v))
		}
	}
}

// formatValue 格式化值为字符串，特别处理布尔值
func (s *StripePayment) formatValue(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		// 检查是否为整数值（可能是从JSON反序列化的int64）
		if v == float32(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)
	case float64:
		// 检查是否为整数值（可能是从JSON反序列化的int64）
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// makeAPIRequest 发送API请求到Stripe
func (s *StripePayment) makeAPIRequest(method, endpoint string, data any) ([]byte, error) {
	apiURL := s.baseURL + endpoint

	var reqBody io.Reader
	var contentType string

	if data != nil {
		// Stripe API要求使用application/x-www-form-urlencoded格式
		values, err := s.structToURLValues(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", consts.ErrRequestDataMarshalFailed, err)
		}
		formData := values.Encode()
		reqBody = strings.NewReader(formData)
		contentType = "application/x-www-form-urlencoded"
		log.Printf("[stripe] API request body: %s", formData)
	} else {
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequest(method, apiURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrRequestCreationFailed, err)
	}

	// 设置请求头 - Stripe使用Bearer认证和form-urlencoded格式
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+s.config.SecretKey)
	req.Header.Set("User-Agent", "LLM-Member/1.0")
	log.Printf("[stripe] API request: %s %s", method, apiURL)

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrRequestSendFailed, err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrResponseReadFailed, err)
	}
	log.Printf("[stripe] API response status: %d, body: %s", resp.StatusCode, string(body))

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w with status %d: %s", consts.ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	return body, nil
}
