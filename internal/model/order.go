package model

import (
	"time"

	"gorm.io/gorm"
)

type method = PaymentMethod

// PlanInfo 套餐信息
type PlanInfo struct {
	Plan     string   `json:"plan"`
	Name     string   `json:"name"`
	Desc     string   `json:"desc"`
	Price    float64  `json:"price"`
	Enabled  bool     `json:"enabled"`
	Features []string `json:"features"`
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

	OrderID string  `json:"order_id" gorm:"uniqueIndex;not null"`
	UserID  uint64  `json:"user_id" gorm:"index"`
	PayPlan PayPlan `json:"pay_plan" gorm:"type:varchar(20)"`
	Amount  float64 `json:"amount" gorm:"not null"`
	Method  method  `json:"method" gorm:"type:varchar(20)"`

	PayURL    string     `json:"payurl"`
	QRCode    string     `json:"qrcode"`
	PaidAt    *time.Time `json:"paid_at"`
	ExpiredAt time.Time  `json:"expired_at"`

	Status OrderStatus `json:"status"` // pending, success, failed, canceled

	User UserModel `json:"user" gorm:"foreignKey:UserID"`

	gorm.Model
}

func (m OrderModel) TableName() string {
	return "llm_order"
}

type TokenModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Email string `json:"email" gorm:"index"`
	Token string `json:"token" gorm:"type:varchar(36)"`

	ExpiredAt time.Time `json:"expired_at"`
	gorm.Model
}

func (m TokenModel) TableName() string {
	return "llm_token"
}
