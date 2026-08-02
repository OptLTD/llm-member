package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"

	"gorm.io/gorm"
)

type CouponService struct {
	db *gorm.DB
}

func NewCouponService() *CouponService {
	return &CouponService{db: config.GetDB()}
}

func (s *CouponService) GetByCode(projectID uint64, code string) (*model.CouponModel, error) {
	var c model.CouponModel
	err := s.db.Where("project_id = ? AND code = ?", projectID, strings.ToUpper(strings.TrimSpace(code))).First(&c).Error
	return &c, err
}

func (s *CouponService) Validate(projectID uint64, code string, userID uint64, plan model.PayPlan, amount float64) (*model.ValidateCouponResponse, error) {
	resp := &model.ValidateCouponResponse{
		OriginalAmount: amount,
		FinalAmount:    amount,
	}
	coupon, err := s.GetByCode(projectID, code)
	if err != nil {
		resp.Valid = false
		resp.Message = "券码不存在"
		return resp, nil
	}
	if coupon.Status != model.CouponStatusActive {
		resp.Message = "券码已失效"
		return resp, nil
	}
	now := time.Now()
	if coupon.ValidFrom != nil && now.Before(*coupon.ValidFrom) {
		resp.Message = "券码尚未生效"
		return resp, nil
	}
	if coupon.ValidTo != nil && now.After(*coupon.ValidTo) {
		resp.Message = "券码已过期"
		return resp, nil
	}
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		resp.Message = "券码已达使用上限"
		return resp, nil
	}
	if coupon.MinAmount > 0 && amount < coupon.MinAmount {
		resp.Message = fmt.Sprintf("订单金额需满 %.2f", coupon.MinAmount)
		return resp, nil
	}
	if coupon.ApplicablePlans != "" {
		ok := false
		for _, p := range strings.Split(coupon.ApplicablePlans, ",") {
			if strings.TrimSpace(p) == string(plan) {
				ok = true
				break
			}
		}
		if !ok {
			resp.Message = "券码不适用于当前套餐"
			return resp, nil
		}
	}

	discount := s.calcDiscount(coupon, amount)
	resp.Valid = true
	resp.DiscountAmount = discount
	resp.FinalAmount = amount - discount
	if resp.FinalAmount < 0 {
		resp.FinalAmount = 0
	}
	return resp, nil
}

func (s *CouponService) calcDiscount(coupon *model.CouponModel, amount float64) float64 {
	switch coupon.Type {
	case model.CouponTypePercent:
		return amount * coupon.Value / 100
	case model.CouponTypeFixed:
		if coupon.Value > amount {
			return amount
		}
		return coupon.Value
	default:
		return 0
	}
}

func (s *CouponService) Redeem(projectID uint64, code string, userID uint64, orderID string, discount float64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var coupon model.CouponModel
		if err := tx.Where("project_id = ? AND code = ?", projectID, strings.ToUpper(strings.TrimSpace(code))).First(&coupon).Error; err != nil {
			return err
		}
		if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
			return errors.New("券码已达使用上限")
		}
		coupon.UsedCount++
		if err := tx.Save(&coupon).Error; err != nil {
			return err
		}
		usage := &model.CouponUsageModel{
			CouponID:       coupon.ID,
			UserID:         userID,
			OrderID:        orderID,
			DiscountAmount: discount,
		}
		return tx.Create(usage).Error
	})
}

func (s *CouponService) RedeemForPlan(projectID uint64, code string, userID uint64) error {
	coupon, err := s.GetByCode(projectID, code)
	if err != nil {
		return err
	}
	if coupon.Type != model.CouponTypePlanExtend {
		return errors.New("该券码不是延期券")
	}
	user, err := NewUserService().GetUserByID(userID)
	if err != nil {
		return err
	}
	days := int(coupon.Value)
	if days <= 0 {
		days = 30
	}
	base := time.Now()
	if user.ExpireAt != nil && user.ExpireAt.After(base) {
		base = *user.ExpireAt
	}
	exp := base.AddDate(0, 0, days)
	user.ExpireAt = &exp
	if err := s.db.Save(user).Error; err != nil {
		return err
	}
	return s.Redeem(projectID, code, userID, "", 0)
}

func (s *CouponService) Create(projectID uint64, coupon *model.CouponModel) error {
	coupon.ProjectID = projectID
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))
	return s.db.Create(coupon).Error
}

func (s *CouponService) Update(id uint64, coupon *model.CouponModel) error {
	return s.db.Model(&model.CouponModel{}).Where("id = ?", id).Updates(map[string]any{
		"type":             coupon.Type,
		"value":            coupon.Value,
		"min_amount":       coupon.MinAmount,
		"max_uses":         coupon.MaxUses,
		"valid_from":       coupon.ValidFrom,
		"valid_to":         coupon.ValidTo,
		"applicable_plans": coupon.ApplicablePlans,
		"status":           coupon.Status,
	}).Error
}

func (s *CouponService) List(projectID uint64, page, size int) ([]model.CouponModel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	var list []model.CouponModel
	q := s.db.Model(&model.CouponModel{})
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (s *CouponService) ListUsages(couponID uint64) ([]model.CouponUsageModel, error) {
	var list []model.CouponUsageModel
	err := s.db.Where("coupon_id = ?", couponID).Order("id desc").Find(&list).Error
	return list, err
}
