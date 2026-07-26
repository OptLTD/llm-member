package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	CouponTypePercent    = "percent"
	CouponTypeFixed      = "fixed"
	CouponTypePlanExtend = "plan_extend"
	CouponStatusActive   = "active"
	CouponStatusDisabled = "disabled"
)

// CouponModel 券码
type CouponModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Code            string    `json:"code" gorm:"column:code;type:varchar(64);uniqueIndex;not null"`
	Type            string    `json:"type" gorm:"column:type;type:varchar(20);not null"`
	Value           float64   `json:"value" gorm:"column:value;type:double;not null"`
	MinAmount       float64   `json:"minAmount" gorm:"column:min_amount;type:double;default:0"`
	MaxUses         int       `json:"maxUses" gorm:"column:max_uses;default:0"` // 0 = unlimited
	UsedCount       int       `json:"usedCount" gorm:"column:used_count;default:0"`
	ValidFrom       *time.Time `json:"validFrom" gorm:"column:valid_from"`
	ValidTo         *time.Time `json:"validTo" gorm:"column:valid_to"`
	ApplicablePlans string    `json:"applicablePlans" gorm:"column:applicable_plans;type:text"` // comma plans
	Status          string    `json:"status" gorm:"column:status;type:varchar(20);default:active"`

	gorm.Model
}

func (CouponModel) TableName() string { return "member_coupon" }

// CouponUsageModel 核销记录
type CouponUsageModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	CouponID       uint64  `json:"couponId" gorm:"column:coupon_id;index;not null"`
	UserID         uint64  `json:"userId" gorm:"column:user_id;index;not null"`
	OrderID        string  `json:"orderId" gorm:"column:order_id;type:varchar(256);index"`
	DiscountAmount float64 `json:"discountAmount" gorm:"column:discount_amount;type:double"`

	gorm.Model
}

func (CouponUsageModel) TableName() string { return "member_coupon_usage" }

type ValidateCouponRequest struct {
	Code    string  `json:"code" binding:"required"`
	PayPlan PayPlan `json:"payPlan" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

type ValidateCouponResponse struct {
	Valid          bool    `json:"valid"`
	Message        string  `json:"message,omitempty"`
	OriginalAmount float64 `json:"originalAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	FinalAmount    float64 `json:"finalAmount"`
}

type RedeemCouponRequest struct {
	Code string `json:"code" binding:"required"`
}
