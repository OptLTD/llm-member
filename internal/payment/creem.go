package payment

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"
)

// CreemPayment Creem支付实现
type CreemPayment struct {
	config *config.Creem
	client *http.Client

	baseURL string // Creem API基础URL
}

// CreemOrder 订单信息结构
type CreemOrder struct {
	ID          string    `json:"id"`
	Mode        string    `json:"mode"`
	Object      string    `json:"object"`
	Customer    string    `json:"customer"`
	Product     string    `json:"product"`
	Transaction string    `json:"transaction"`
	Amount      int       `json:"amount"`
	SubTotal    int       `json:"sub_total"`
	TaxAmount   int       `json:"tax_amount"`
	AmountDue   int       `json:"amount_due"`
	AmountPaid  int       `json:"amount_paid"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreemCustomer 客户信息结构
type CreemCustomer struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Mode  string `json:"mode,omitempty"`

	Object  string `json:"object,omitempty"`
	Country string `json:"country,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// 添加产品信息结构体
type CreemProduct struct {
	ID                string    `json:"id"`
	Object            string    `json:"object"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	ImageURL          string    `json:"image_url"`
	Price             int       `json:"price"`
	Currency          string    `json:"currency"`
	BillingType       string    `json:"billing_type"`
	BillingPeriod     string    `json:"billing_period"`
	Status            string    `json:"status"`
	TaxMode           string    `json:"tax_mode"`
	TaxCategory       string    `json:"tax_category"`
	DefaultSuccessURL *string   `json:"default_success_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Mode              string    `json:"mode"`
}

// CreemCheckoutReq Creem创建支付请求结构
type CreemCheckoutReq struct {
	RequestID  string `json:"request_id"`
	ProductID  string `json:"product_id"`
	SuccessURL string `json:"success_url,omitempty"`
	Metadata   object `json:"metadata,omitempty"`
	Units      int    `json:"units,omitempty"`
	// customer
	Customer *CreemCustomer `json:"customer,omitempty"`
}

// CreemCheckoutRes Creem创建支付响应结构
type CreemCheckoutRes struct {
	ID      string `json:"id"`
	Mode    string `json:"mode"`
	Object  string `json:"object"`
	Status  string `json:"status"`
	Product string `json:"product"`
	Units   int    `json:"units"`

	RequestID string `json:"request_id"`
	Customer  string `json:"customer"`
	Metadata  object `json:"metadata"`

	CheckoutURL string `json:"checkout_url"`
	SuccessURL  string `json:"success_url"`
	// order
	Order *CreemOrder `json:"order"`
}

// CreemCheckoutStatus Creem查询支付状态响应结构
type CreemCheckoutStatus struct {
	ID     string `json:"id"`
	Mode   string `json:"mode"`
	Object string `json:"object"`
	Status string `json:"status"`
	Units  int    `json:"units"`

	RequestID string `json:"request_id"`
	Metadata  object `json:"metadata"`
	// data
	Product  *CreemProduct  `json:"product"`
	Customer *CreemCustomer `json:"customer"`
	// order
	Order *CreemOrder `json:"order,omitempty"`
}

// NewCreemPayment 创建Creem支付实例
func NewCreemPayment() *CreemPayment {
	baseURL := "https://api.creem.io/v1"
	if os.Getenv("APP_MODE") != "release" {
		baseURL = "https://test-api.creem.io/v1"
	}

	return &CreemPayment{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ensureClientReady 确保客户端已准备就绪
func (c *CreemPayment) ensureClientReady() error {
	if c.config != nil {
		return nil
	}

	// 获取配置
	provider := config.GetCreemConfig()
	if provider == nil {
		return consts.ErrPaymentProviderNotConfigured
	}

	c.config = provider
	return nil
}

// Create 创建Creem支付订单
func (c *CreemPayment) Create(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := c.ensureClientReady(); err != nil {
		return err
	}

	request := CreemCheckoutReq{
		RequestID: order.OrderID,
		Customer: &CreemCustomer{
			Email: order.User.Email,
		},
		Metadata: object{
			"order_id": order.OrderID,
			"user_id":  order.UserID,
			"pay_plan": string(order.PayPlan),
		},
		Units: 1, // 默认购买1个单位
		// SuccessURL: fmt.Sprintf(
		// 	"http://%s/success?order_id=%s",
		// 	"localhost", order.OrderID,
		// ),
	}
	switch order.PayPlan {
	case model.PlanBasic:
		request.ProductID = c.config.PlanBasicId
	case model.PlanExtra:
		request.ProductID = c.config.PlanExtraId
	case model.PlanUltra:
		request.ProductID = c.config.PlanUltraId
	case model.PlanSuper:
		request.ProductID = c.config.PlanSuperId
	default:
		return fmt.Errorf("%w: %v", consts.ErrPaymentPlanNotSupported, order.PayPlan)
	}

	// 发送HTTP请求到Creem API
	resp, err := c.makeAPIRequest("POST", "/checkouts", request)
	if err != nil {
		log.Printf("[creem][%s] payment creation failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentCreationFailed, err)
	}

	var checkoutResp CreemCheckoutRes
	if err := json.Unmarshal(resp, &checkoutResp); err != nil {
		log.Printf("[creem][%s] failed to parse response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 更新订单信息
	order.PayURL = checkoutResp.CheckoutURL
	order.ThridID = checkoutResp.ID
	order.Status = model.PaymentPending
	log.Printf(
		"[creem][%s][%s] payment created successfully, checkout URL: %s",
		order.OrderID, order.ThridID, checkoutResp.CheckoutURL,
	)
	return nil
}

// Webhook Creem支付回调验证
func (c *CreemPayment) Webhook(req *http.Request) (*Event, error) {
	if err := c.ensureClientReady(); err != nil {
		return nil, err
	}

	if c.config.WhSecret == "" {
		return nil, consts.ErrWebhookSecretNotConfigured
	}
	event, err := c.HandleWebhook(req, c.config.WhSecret)
	if err != nil {
		log.Printf("[creem] webhook verification failed: %v", err)
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookSignatureVerificationFailed, err)
	}

	log.Printf("[creem] received webhook event type: %s, id: %s", event.EventType, event.ID)

	orderID, amount, status := "", 0, "unknown"
	switch event.EventType {
	case "checkout.completed", "payment.succeeded", "order.completed":
		status = "success"
		// 从事件数据中提取订单ID和金额
		if data, ok := event.Object["order"].(object); ok {
			if amt, exists := data["amount"].(float64); exists {
				amount = int(amt)
			}
		}
		// 尝试从metadata中获取订单ID
		if metadata, ok := event.Object["metadata"].(object); ok {
			if oid, exists := metadata["order_id"].(string); exists {
				orderID = oid
			}
		}
		// 如果还是没有订单ID，尝试使用request_id
		if orderID == "" {
			if reqID, ok := event.Object["request_id"].(string); ok {
				orderID = reqID
			}
		}

	case "checkout.failed", "payment.failed", "order.failed":
		status = "failed"
		// 同样提取订单信息
		if metadata, ok := event.Object["metadata"].(object); ok {
			if oid, exists := metadata["order_id"].(string); exists {
				orderID = oid
			}
		}
		if orderID == "" {
			if reqID, ok := event.Object["request_id"].(string); ok {
				orderID = reqID
			}
		}

	case "checkout.cancelled", "payment.cancelled", "order.cancelled":
		status = "cancelled"
		if metadata, ok := event.Object["metadata"].(object); ok {
			if oid, exists := metadata["order_id"].(string); exists {
				orderID = oid
			}
		}
		if orderID == "" {
			if reqID, ok := event.Object["request_id"].(string); ok {
				orderID = reqID
			}
		}

	default:
		log.Printf("[creem] unknown webhook event type: %s", event.EventType)
		status = "unknown"
		// 尝试提取订单ID
		if metadata, ok := event.Object["metadata"].(object); ok {
			if oid, exists := metadata["order_id"].(string); exists {
				orderID = oid
			}
		}
	}

	if orderID == "" {
		log.Printf("[creem] warning: could not extract order_id from webhook event %s", event.ID)
		// 使用事件ID作为fallback
		orderID = event.ID
	}

	log.Printf("[creem] processed webhook: order_id=%s, status=%s, amount=%d", orderID, status, amount)

	return &Event{
		Type: event.EventType, Data: event.Object,
		OrderID: orderID, Amount: float64(amount),
		Status: status, Time: event.CreatedAt / 1000,
	}, nil
}

// Query 查询Creem支付状态
func (c *CreemPayment) Query(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := c.ensureClientReady(); err != nil {
		return err
	}

	// 使用checkout session ID查询支付状态
	var statusResp CreemCheckoutStatus
	url := fmt.Sprintf("/checkouts?checkout_id=%s", order.ThridID)
	if resp, err := c.makeAPIRequest("GET", url, nil); err != nil {
		log.Printf("[creem][%s] payment query failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentQueryFailed, err)
	} else if err := json.Unmarshal(resp, &statusResp); err != nil {
		log.Printf("[creem][%s] failed to parse query response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 根据Creem返回的状态更新订单状态
	// 检查checkout session状态和order状态
	var paymentStatus string
	if statusResp.Order != nil {
		paymentStatus = statusResp.Order.Status
	} else {
		paymentStatus = statusResp.Status
	}

	switch paymentStatus {
	case "completed", "paid":
		order.Status = model.PaymentSucceed
		log.Printf("[creem][%s] payment successful", order.OrderID)
		if order.SucceedAt == nil {
			now := time.Now()
			order.SucceedAt = &now
		}
	case "pending":
		order.Status = model.PaymentPending
		log.Printf("[creem][%s] payment pending", order.OrderID)
	case "cancelled", "expired", "canceled":
		order.Status = model.PaymentCanceled
		log.Printf("[creem][%s] payment cancelled/expired", order.OrderID)
	case "failed":
		order.Status = model.PaymentCanceled
		log.Printf("[creem][%s] payment failed", order.OrderID)
	default:
		log.Printf("[creem][%s] unknown payment status: %s", order.OrderID, paymentStatus)
	}

	return nil
}

// Close 关闭Creem支付（取消未完成的支付）
func (c *CreemPayment) Close(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := c.ensureClientReady(); err != nil {
		return err
	}

	log.Printf("[creem][%s] attempting to close payment", order.OrderID)

	// Creem可能不支持直接关闭checkout session
	// 但可以尝试调用相关API或者只是更新本地状态
	if order.Status == model.PaymentPending {
		// 这里可以调用Creem的取消API（如果有的话）
		// 目前只是更新本地状态
		order.Status = model.PaymentCanceled
		log.Printf("[creem][%s] payment closed locally", order.OrderID)
	} else {
		log.Printf("[creem][%s] payment cannot be closed, current status: %s", order.OrderID, order.Status)
	}

	return nil
}

// Refund 处理Creem退款
func (c *CreemPayment) Refund(order *model.OrderModel) error {
	// 检查配置和初始化客户端
	if err := c.ensureClientReady(); err != nil {
		return err
	}

	log.Printf("[creem][%s] processing refund", order.OrderID)

	// 检查订单状态是否可以退款
	if order.Status != model.PaymentSucceed {
		return fmt.Errorf("%w, current status: %s", consts.ErrOrderCannotBeRefunded, order.Status)
	}

	// 构建退款请求 - 根据Creem API文档调整
	refundReq := object{
		"order_id": order.OrderID, // 或者使用实际的Creem order ID
		"amount":   order.Amount,
		"reason":   "User requested refund",
	}

	// 发送退款请求 - 使用正确的API端点
	resp, err := c.makeAPIRequest("POST", "/refunds", refundReq)
	if err != nil {
		log.Printf("[creem][%s] refund failed: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrPaymentRefundFailed, err)
	}

	// 解析退款响应
	var refundResp object
	if err := json.Unmarshal(resp, &refundResp); err != nil {
		log.Printf("[creem][%s] failed to parse refund response: %v", order.OrderID, err)
		return fmt.Errorf("%w: %v", consts.ErrResponseParseFailed, err)
	}

	// 检查退款状态
	if status, ok := refundResp["status"].(string); ok && status == "success" {
		// 更新订单状态为已退款
		order.Status = model.PaymentRefunded
		log.Printf("[creem][%s] refund processed successfully", order.OrderID)
	} else {
		log.Printf("[creem][%s] refund response: %v", order.OrderID, refundResp)
		return consts.ErrRefundNotSuccessful
	}

	return nil
}

// makeAPIRequest 发送API请求到Creem
func (c *CreemPayment) makeAPIRequest(method, endpoint string, data any) ([]byte, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", consts.ErrRequestDataMarshalFailed, err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		log.Printf("[creem] API request body: %s", string(jsonData))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrRequestCreationFailed, err)
	}

	// 设置请求头 - Creem使用x-api-key认证
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.ApiKey) // 使用x-api-key认证
	req.Header.Set("User-Agent", "LLM-Member/1.0")

	log.Printf("[creem] API request: %s %s", method, url)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrRequestSendFailed, err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrResponseReadFailed, err)
	}

	log.Printf("[creem] API response status: %d, body: %s", resp.StatusCode, string(body))

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w with status %d: %s", consts.ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	return body, nil
}

// CreemWebhookEvent Creem webhook事件结构
// 根据Creem官方文档定义的webhook事件格式
// 常见的事件类型包括：
// - checkout.completed: 支付完成
// - checkout.failed: 支付失败
// - checkout.cancelled: 支付取消
// - payment.succeeded: 支付成功
// - payment.failed: 支付失败
// - order.completed: 订单完成
// - order.failed: 订单失败
type CreemWebhookEvent struct {
	ID        string `json:"id"`        // 事件唯一标识符
	EventType string `json:"eventType"` // 事件类型（Creem使用eventType而不是type）
	Object    object `json:"object"`    // 事件数据，包含订单、支付等信息

	CreatedAt int64 `json:"created_at"` // 事件创建时间（Unix时间戳毫秒）
}

// HandleWebhook 处理Creem webhook事件
func (c *CreemPayment) HandleWebhook(req *http.Request, webhookSecret string) (*CreemWebhookEvent, error) {
	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookBodyReadFailed, err)
	}

	// 记录原始webhook数据用于调试
	log.Printf("[creem] webhook received: method=%s, content-length=%d, user-agent=%s",
		req.Method, len(body), req.Header.Get("User-Agent"))

	// 验证HTTP方法
	if req.Method != "POST" {
		return nil, fmt.Errorf("%w: %s", consts.ErrInvalidHTTPMethod, req.Method)
	}

	// 验证Content-Type
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
		log.Printf("[creem] warning: unexpected content-type: %s", contentType)
	}

	// 验证webhook签名
	if err := c.verifyWebhookSignature(req, body, webhookSecret); err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookSignatureVerificationFailed, err)
	}

	// 解析webhook事件
	var event CreemWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[creem] failed to parse webhook JSON: %s", string(body))
		return nil, fmt.Errorf("%w: %v", consts.ErrWebhookEventParseFailed, err)
	}

	// 验证事件基本字段
	if event.ID == "" {
		return nil, consts.ErrWebhookEventMissingID
	}
	if event.EventType == "" {
		return nil, consts.ErrWebhookEventMissingEventType
	}

	// 检查事件时间戳，防止重放攻击（允许5分钟的时间差）
	if event.CreatedAt > 0 {
		eventTime := time.Unix(event.CreatedAt/1000, 0)
		timeDiff := time.Since(eventTime)
		if timeDiff > 5*time.Minute {
			log.Printf("[creem] warning: webhook event is too old: %v", timeDiff)
			// 注意：这里只是警告，不阻止处理，因为时钟可能不同步
		}
		if timeDiff < -1*time.Minute {
			log.Printf("[creem] warning: webhook event is from future: %v", timeDiff)
		}
	}

	log.Printf("[creem] successfully parsed webhook event: id=%s, type=%s, created_at=%d",
		event.ID, event.EventType, event.CreatedAt)
	return &event, nil
}

// verifyWebhookSignature 验证webhook签名
// 根据Creem官方文档，webhook签名验证流程：
// 1. 从请求头"creem-signature"获取签名
// 2. 使用HMAC-SHA256算法，以webhook密钥为key，请求体为消息计算签名
// 3. 将计算出的签名与接收到的签名进行比较
// 注意：签名比较使用hmac.Equal防止时序攻击
func (c *CreemPayment) verifyWebhookSignature(req *http.Request, body []byte, webhookSecret string) error {
	// 获取签名头 - Creem使用'creem-signature'头
	signatureHeader := req.Header.Get("creem-signature")
	if signatureHeader == "" {
		// 尝试其他可能的头名称
		signatureHeader = req.Header.Get("Creem-Signature")
		if signatureHeader == "" {
			return consts.ErrMissingWebhookSignatureHeader
		}
	}

	// 验证webhook密钥是否配置
	if webhookSecret == "" {
		return consts.ErrWebhookSecretNotConfigured
	}

	// 计算期望的签名
	// 使用HMAC-SHA256算法，密钥为webhookSecret，消息为请求体
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(body)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// 记录签名验证信息（调试用）
	log.Printf("[creem] signature verification: received=%s, expected=%s",
		signatureHeader, expectedSignature)

	// 安全比较签名（使用hmac.Equal防止时序攻击）
	if !hmac.Equal([]byte(signatureHeader), []byte(expectedSignature)) {
		return fmt.Errorf("%w: received=%s, expected=%s",
			consts.ErrSignatureMismatch, signatureHeader, expectedSignature)
	}

	log.Printf("[creem] webhook signature verification successful")
	return nil
}
