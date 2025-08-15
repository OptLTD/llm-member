package model

// Role 角色
type Role string

const (
	RoleUser  Role = "user"  // 用户
	RoleAdmin Role = "admin" // 管理员
)

// PayPlan 套餐类型
type PayPlan string

const (
	PlanBasic PayPlan = "basic" // Basic
	PlanExtra PayPlan = "extra" // Extra
	PlanUltra PayPlan = "ultra" // Ultra
	PlanSuper PayPlan = "super" // Super
)

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentAlipay PaymentMethod = "alipay"
	PaymentWechat PaymentMethod = "wechat"
	PaymentUnion  PaymentMethod = "union"
	PaymentMock   PaymentMethod = "mock"
	PaymentStripe PaymentMethod = "stripe"
)

type OrderStatus string

const (
	PaymentPending  OrderStatus = "pending"
	PaymentSucceed  OrderStatus = "succeed"
	PaymentRefunded OrderStatus = "refunded"
	PaymentCanceled OrderStatus = "canceled"
)

// PaginateRequest 分页请求结构体
type PaginateRequest struct {
	Page uint `json:"page" binding:"min=1"`         // 页码，从1开始
	Size uint `json:"size" binding:"min=1,max=100"` // 每页大小，最大100

	Query map[string]any `json:"query"`
}

// PaginateResponse 分页响应结构体
type PaginateResponse struct {
	Data  any   `json:"data"`  // 数据列表
	Page  uint  `json:"page"`  // 当前页码
	Size  uint  `json:"size"`  // 每页大小
	Count uint  `json:"count"` // 总页数
	Total int64 `json:"total"` // 总记录数
}
