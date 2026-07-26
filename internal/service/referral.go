package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"

	"gorm.io/gorm"
)

type ReferralService struct {
	db *gorm.DB
}

func NewReferralService() *ReferralService {
	return &ReferralService{db: config.GetDB()}
}

func (s *ReferralService) genCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

func (s *ReferralService) GetOrCreateCode(userID uint64) (*model.ReferralCodeModel, error) {
	var rc model.ReferralCodeModel
	err := s.db.Where("user_id = ?", userID).First(&rc).Error
	if err == nil {
		return &rc, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	rc = model.ReferralCodeModel{
		UserID:  userID,
		Code:    s.genCode(),
		Enabled: true,
	}
	for i := 0; i < 5; i++ {
		if err := s.db.Create(&rc).Error; err != nil {
			rc.Code = s.genCode()
			continue
		}
		return &rc, nil
	}
	return nil, errors.New("failed to create referral code")
}

func (s *ReferralService) ResolveCode(code string) (*model.ReferralCodeModel, error) {
	var rc model.ReferralCodeModel
	err := s.db.Where("code = ? AND enabled = ?", strings.ToUpper(strings.TrimSpace(code)), true).First(&rc).Error
	return &rc, err
}

func (s *ReferralService) LogShare(code, channel, ipHash string, campaignID uint64) error {
	log := &model.ShareLogModel{
		Code:       strings.ToUpper(strings.TrimSpace(code)),
		Channel:    channel,
		IPHash:     ipHash,
		CampaignID: campaignID,
	}
	return s.db.Create(log).Error
}

func (s *ReferralService) BindReferral(inviteeID uint64, refCode string) error {
	rc, err := s.ResolveCode(refCode)
	if err != nil {
		return nil // silent ignore invalid code
	}
	if rc.UserID == inviteeID {
		return nil
	}
	var existing model.ReferralModel
	if err := s.db.Where("invitee_id = ?", inviteeID).First(&existing).Error; err == nil {
		return nil
	}
	ref := &model.ReferralModel{
		InviterID: rc.UserID,
		InviteeID: inviteeID,
		Status:    model.ReferralStatusRegistered,
	}
	if err := s.db.Create(ref).Error; err != nil {
		return err
	}
	s.db.Model(&model.ShareLogModel{}).
		Where("code = ? AND converted = ?", rc.Code, false).
		Update("converted", true)
	if config.HasModule(config.ModuleAnalytics) {
		_ = NewAnalyticsService().TrackInternal(inviteeID, "referral.register", map[string]any{
			"inviterId": rc.UserID,
			"code":      rc.Code,
		})
	}
	return nil
}

func (s *ReferralService) OnPaymentSuccess(userID uint64) error {
	var ref model.ReferralModel
	err := s.db.Where("invitee_id = ? AND status = ?", userID, model.ReferralStatusRegistered).First(&ref).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	ref.Status = model.ReferralStatusPaid
	if err := s.db.Save(&ref).Error; err != nil {
		return err
	}
	return s.grantReward(&ref)
}

func (s *ReferralService) grantReward(ref *model.ReferralModel) error {
	if ref.RewardStatus == "granted" {
		return nil
	}
	inviter, err := NewUserService().GetUserByID(ref.InviterID)
	if err != nil {
		return err
	}
	base := time.Now()
	if inviter.ExpireAt != nil && inviter.ExpireAt.After(base) {
		base = *inviter.ExpireAt
	}
	exp := base.AddDate(0, 0, 7)
	inviter.ExpireAt = &exp
	if err := s.db.Save(inviter).Error; err != nil {
		return err
	}
	ref.RewardStatus = "granted"
	return s.db.Save(ref).Error
}

func (s *ReferralService) Stats(userID uint64) (map[string]any, error) {
	code, _ := s.GetOrCreateCode(userID)
	var invited, paid int64
	s.db.Model(&model.ReferralModel{}).Where("inviter_id = ?", userID).Count(&invited)
	s.db.Model(&model.ReferralModel{}).Where("inviter_id = ? AND status = ?", userID, model.ReferralStatusPaid).Count(&paid)
	return map[string]any{
		"code":    code.Code,
		"invited": invited,
		"paid":    paid,
	}, nil
}

func (s *ReferralService) ListCampaigns() ([]model.CampaignModel, error) {
	var list []model.CampaignModel
	err := s.db.Order("id desc").Find(&list).Error
	return list, err
}

func (s *ReferralService) SaveCampaign(c *model.CampaignModel) error {
	if c.ID == 0 {
		return s.db.Create(c).Error
	}
	return s.db.Save(c).Error
}

func (s *ReferralService) ListReferrals(page, size int) ([]model.ReferralModel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	var total int64
	var list []model.ReferralModel
	s.db.Model(&model.ReferralModel{}).Count(&total)
	err := s.db.Order("id desc").Offset((page-1)*size).Limit(size).Find(&list).Error
	return list, total, err
}
