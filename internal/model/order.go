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
	PayPlan PayPlan `json:"pay_plan" binding:"required"`
	Method  method  `json:"method" binding:"required"`

	Amount *float64 `json:"amount,omitempty"` // 基础版可自定义金额
	UserId *uint64  `json:"-,omitempty"`
}

// OrderResponse 支付响应
type OrderResponse struct {
	OrderID string  `json:"order_id"`
	PayURL  string  `json:"pay_url"`
	Amount  float64 `json:"amount"`
	QRCode  string  `json:"qrcode"`
	Status  string  `json:"status"`
	Method  string  `json:"method"`
}

// OrderModel 支付订单
type OrderModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID  uint64  `json:"user_id" gorm:"index;not null"`
	OrderID string  `json:"order_id" gorm:"type:varchar(64);uniqueIndex;not null"`
	PayPlan PayPlan `json:"pay_plan" gorm:"type:varchar(20)"`
	Amount  float64 `json:"amount" gorm:"not null"`
	Method  method  `json:"method" gorm:"type:varchar(20)"`
	PayURL  string  `json:"payurl"`
	QRCode  string  `json:"qrcode"`

	SucceedAt *time.Time `json:"succeed_at"`
	ExpiredAt time.Time  `json:"expired_at"`

	Status OrderStatus `json:"status"` // pending, success, failed, canceled

	User UserModel `json:"user" gorm:"foreignKey:UserID"`

	gorm.Model
}

func (m OrderModel) TableName() string {
	return "llm_order"
}

type VerifyModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Email string `json:"email" gorm:"index"`
	Token string `json:"token" gorm:"type:varchar(64)"`

	ExpireAt time.Time `json:"expire_at"`
	gorm.Model
}

func (m VerifyModel) TableName() string {
	return "llm_verify"
}
