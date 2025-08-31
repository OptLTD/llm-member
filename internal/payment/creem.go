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
	"strconv"
	"strings"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
)

type object = map[string]any

// CreemPayment Creem支付实现
type CreemPayment struct {
	config *config.PaymentProvider
	client *http.Client
}

// CreemCustomer 客户信息结构
type CreemCustomer struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}

// CreemCheckoutRequest Creem创建支付请求结构
type CreemCheckoutRequest struct {
	RequestID  string `json:"request_id"`
	ProductID  string `json:"product_id"`
	SuccessURL string `json:"success_url,omitempty"`
	Metadata   object `json:"metadata,omitempty"`
	Units      int    `json:"units,omitempty"`
	// customer
	Customer *CreemCustomer `json:"customer,omitempty"`
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

// CreemCheckoutResponse Creem创建支付响应结构
type CreemCheckoutResponse struct {
	ID          string `json:"id"`
	Mode        string `json:"mode"`
	Object      string `json:"object"`
	Status      string `json:"status"`
	RequestID   string `json:"request_id"`
	Product     string `json:"product"`
	Units       int    `json:"units"`
	Customer    string `json:"customer"`
	CheckoutURL string `json:"checkout_url"`
	SuccessURL  string `json:"success_url"`
	Metadata    object `json:"metadata"`
	// order
	Order *CreemOrder `json:"order"`
}

// CreemCheckoutStatusResponse Creem查询支付状态响应结构
type CreemCheckoutStatusResponse struct {
	ID        string `json:"id"`
	Mode      string `json:"mode"`
	Object    string `json:"object"`
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	Product   string `json:"product"`
	Customer  string `json:"customer"`
	Metadata  object `json:"metadata"`
	// order
	Order *CreemOrder `json:"order,omitempty"`
}

// NewCreemPayment 创建Creem支付实例
func NewCreemPayment() (IPayment, error) {
	provider := config.GetPaymentProvider("creem")
	if provider == nil {
		return nil, fmt.Errorf("creem payment provider not configured")
	}

	return &CreemPayment{
		config: provider,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Create 创建Creem支付订单
func (c *CreemPayment) Create(order *model.OrderModel) error {
	// 根据套餐类型设置产品ID（需要在Creem后台预先创建产品）
	var productID string
	switch order.PayPlan {
	case model.PlanBasic:
		productID = "prod_basic_plan" // 需要在Creem后台创建对应产品，格式为prod_xxx
	case model.PlanExtra:
		productID = "prod_extra_plan"
	case model.PlanUltra:
		productID = "prod_ultra_plan"
	case model.PlanSuper:
		productID = "prod_super_plan"
	default:
		productID = "prod_basic_plan"
	}

	// 创建支付请求
	request := CreemCheckoutRequest{
		RequestID: order.OrderID, // 使用订单ID作为request_id来跟踪支付
		ProductID: productID,
		Customer: &CreemCustomer{
			ID: strconv.FormatUint(order.UserID, 10),
		},
		SuccessURL: fmt.Sprintf("https://your-domain.com/payment/success?order_id=%s", order.OrderID),
		Metadata: object{
			"order_id": order.OrderID,
			"plan":     string(order.PayPlan),
			"user_id":  order.UserID,
		},
		Units: 1, // 默认购买1个单位
	}

	// 发送HTTP请求到Creem API
	resp, err := c.makeAPIRequest("POST", "/checkouts", request)
	if err != nil {
		log.Printf("[creem][%s] payment creation failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to create creem payment: %v", err)
	}

	var checkoutResp CreemCheckoutResponse
	if err := json.Unmarshal(resp, &checkoutResp); err != nil {
		log.Printf("[creem][%s] failed to parse response: %v", order.OrderID, err)
		return fmt.Errorf("failed to parse creem response: %v", err)
	}

	// 更新订单信息
	order.PayURL = checkoutResp.CheckoutURL
	order.QRCode = checkoutResp.CheckoutURL // Creem使用URL进行支付
	order.Status = model.PaymentPending

	log.Printf("[creem][%s] payment created successfully, checkout URL: %s", order.OrderID, checkoutResp.CheckoutURL)
	return nil
}

// Query 查询Creem支付状态
func (c *CreemPayment) Query(order *model.OrderModel) error {
	// 使用checkout session ID查询支付状态
	// 注意：这里需要存储checkout session ID，或者使用其他方式查询
	// 由于Creem API可能不支持通过request_id直接查询，这里使用一个假设的endpoint
	url := fmt.Sprintf("/checkouts/%s", order.OrderID) // 假设可以通过request_id查询
	resp, err := c.makeAPIRequest("GET", url, nil)
	if err != nil {
		log.Printf("[creem][%s] payment query failed: %v", order.OrderID, err)
		return fmt.Errorf("failed to query creem payment: %v", err)
	}

	var statusResp CreemCheckoutStatusResponse
	if err := json.Unmarshal(resp, &statusResp); err != nil {
		log.Printf("[creem][%s] failed to parse query response: %v", order.OrderID, err)
		return fmt.Errorf("failed to parse creem query response: %v", err)
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
	log.Printf("[creem][%s] processing refund", order.OrderID)

	// 检查订单状态是否可以退款
	if order.Status != model.PaymentSucceed {
		return fmt.Errorf("order %s cannot be refunded, current status: %s", order.OrderID, order.Status)
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
		return fmt.Errorf("failed to process creem refund: %v", err)
	}

	// 解析退款响应
	var refundResp object
	if err := json.Unmarshal(resp, &refundResp); err != nil {
		log.Printf("[creem][%s] failed to parse refund response: %v", order.OrderID, err)
		return fmt.Errorf("failed to parse creem refund response: %v", err)
	}

	// 检查退款状态
	if status, ok := refundResp["status"].(string); ok && status == "success" {
		// 更新订单状态为已退款
		order.Status = model.PaymentRefunded
		log.Printf("[creem][%s] refund processed successfully", order.OrderID)
	} else {
		log.Printf("[creem][%s] refund response: %v", order.OrderID, refundResp)
		return fmt.Errorf("refund was not successful")
	}

	return nil
}

// makeAPIRequest 发送API请求到Creem
func (c *CreemPayment) makeAPIRequest(method, endpoint string, data any) ([]byte, error) {
	baseURL := "https://api.creem.io/v1" // Creem API基础URL
	url := baseURL + endpoint

	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		log.Printf("[creem] API request body: %s", string(jsonData))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头 - Creem使用x-api-key认证
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.Token) // 使用x-api-key认证
	req.Header.Set("User-Agent", "LLM-Member/1.0")

	log.Printf("[creem] API request: %s %s", method, url)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("[creem] API response status: %d, body: %s", resp.StatusCode, string(body))

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// CreemWebhookEvent Creem webhook事件结构
type CreemWebhookEvent struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Type   string `json:"type"`
	Data   object `json:"data"`

	CreatedAt time.Time `json:"created_at"`
}

// HandleWebhook 处理Creem webhook事件
func (c *CreemPayment) HandleWebhook(req *http.Request, webhookSecret string) (*CreemWebhookEvent, error) {
	// 读取请求体
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook body: %v", err)
	}

	// 验证webhook签名
	if err := c.verifyWebhookSignature(req, body, webhookSecret); err != nil {
		return nil, fmt.Errorf("webhook signature verification failed: %v", err)
	}

	// 解析webhook事件
	var event CreemWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %v", err)
	}

	log.Printf("[creem] received webhook event: %s, type: %s", event.ID, event.Type)
	return &event, nil
}

// verifyWebhookSignature 验证Creem webhook签名
func (c *CreemPayment) verifyWebhookSignature(req *http.Request, body []byte, webhookSecret string) error {
	// 获取签名头
	signatureHeader := req.Header.Get("X-Creem-Signature")
	if signatureHeader == "" {
		return fmt.Errorf("missing webhook signature")
	}

	// 解析签名
	// Creem签名格式通常是: t=timestamp,v1=signature
	parts := strings.Split(signatureHeader, ",")
	var timestamp, signature string
	for _, part := range parts {
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			signature = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || signature == "" {
		return fmt.Errorf("invalid signature format")
	}

	// 构建签名字符串
	signedPayload := timestamp + "." + string(body)

	// 计算HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// 比较签名
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
