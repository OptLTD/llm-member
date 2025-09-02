package model

import (
	"time"

	"gorm.io/gorm"
)

type method = PaymentMethod

// PlanInfo 套餐信息
type PlanInfo struct {
	Plan     string   `json:"plan" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Brief    string   `json:"brief" binding:"required"`
	Price    float64  `json:"price" binding:"required"`
	Usage    string   `json:"usage" binding:"required"`  // 用量
	Period   string   `json:"period" binding:"required"` // 周期
	Enabled  bool     `json:"enabled" binding:"required"`
	Features []string `json:"features" binding:"required"`
}

// OrderRequest 支付请求
type OrderRequest struct {
	PayPlan PayPlan `json:"payPlan" binding:"required"`
	Method  method  `json:"method" binding:"required"`

	Amount *float64 `json:"amount,omitempty"` // 基础版可自定义金额
	UserId *uint64  `json:"-,omitempty"`
}

// OrderResponse 支付响应
type OrderResponse struct {
	OrderID string  `json:"orderId"`
	PayURL  string  `json:"payUrl"`
	Amount  float64 `json:"amount"`
	QRCode  string  `json:"qrcode"`
	Status  string  `json:"status"`
	Method  string  `json:"method"`
}

// OrderModel 支付订单
type OrderModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID  uint64  `json:"userId" gorm:"column:user_id;index;not null"`
	OrderID string  `json:"orderId" gorm:"column:order_id;type:varchar(256);uniqueIndex;not null"`
	ThridID string  `json:"thridId" gorm:"column:thrid_id;type:varchar(256);index"`
	PayPlan PayPlan `json:"payPlan" gorm:"column:pay_plan;type:varchar(20)"`
	Amount  float64 `json:"amount" gorm:"column:amount;type:double;not null"`
	Method  method  `json:"method" gorm:"column:method;type:varchar(20)"`
	PayURL  string  `json:"payurl" gorm:"column:pay_url;type:text"`
	QRCode  string  `json:"qrcode" gorm:"column:qrcode;type:text"`

	SucceedAt *time.Time `json:"succeedAt" gorm:"column:succeed_at"`
	ExpiredAt time.Time  `json:"expiredAt" gorm:"column:expired_at"`

	Status OrderStatus `json:"status" gorm:"column:status;type:varchar(10)"` // pending, success, failed, canceled

	User UserModel `json:"user" gorm:"column:user_id;foreignKey:UserID"`

	gorm.Model
}

func (m OrderModel) TableName() string {
	return "llm_order"
}

type VerifyModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Email string `json:"email" gorm:"column:email;index"`
	Token string `json:"token" gorm:"column:token;type:varchar(64)"`

	ExpireAt time.Time `json:"expireAt"  gorm:"column:expire_at"`
	gorm.Model
}

func (m VerifyModel) TableName() string {
	return "llm_verify"
}
