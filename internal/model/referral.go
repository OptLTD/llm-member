package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ReferralStatusPending    = "pending"
	ReferralStatusRegistered = "registered"
	ReferralStatusPaid       = "paid"
)

// ReferralCodeModel 用户邀请码
type ReferralCodeModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID    uint64 `json:"userId" gorm:"column:user_id;index;not null"`
	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	Code      string `json:"code" gorm:"column:code;type:varchar(32);index;not null"`
	Enabled   bool   `json:"enabled" gorm:"column:enabled;default:true"`

	gorm.Model
}

func (ReferralCodeModel) TableName() string { return "member_referral_code" }

// ReferralModel 邀请关系
type ReferralModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	InviterID    uint64 `json:"inviterId" gorm:"column:inviter_id;index;not null"`
	ProjectID    uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	InviteeID    uint64 `json:"inviteeId" gorm:"column:invitee_id;index;not null"`
	CampaignID   uint64 `json:"campaignId" gorm:"column:campaign_id;index"`
	Status       string `json:"status" gorm:"column:status;type:varchar(20);default:pending"`
	RewardStatus string `json:"rewardStatus" gorm:"column:reward_status;type:varchar(20);default:none"`

	gorm.Model
}

func (ReferralModel) TableName() string { return "member_referral" }

// CampaignModel 分享活动
type CampaignModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Name       string     `json:"name" gorm:"column:name;type:varchar(128);not null"`
	ProjectID  uint64     `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	RewardType string     `json:"rewardType" gorm:"column:reward_type;type:varchar(32)"` // plan_extend | coupon
	Rules      string     `json:"rules" gorm:"column:rules;type:text"`                   // JSON
	StartAt    *time.Time `json:"startAt" gorm:"column:start_at"`
	EndAt      *time.Time `json:"endAt" gorm:"column:end_at"`
	Enabled    bool       `json:"enabled" gorm:"column:enabled;default:true"`

	gorm.Model
}

func (CampaignModel) TableName() string { return "member_campaign" }

// ShareLogModel 分享点击
type ShareLogModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Code       string `json:"code" gorm:"column:code;type:varchar(32);index"`
	ProjectID  uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	CampaignID uint64 `json:"campaignId" gorm:"column:campaign_id;index"`
	Channel    string `json:"channel" gorm:"column:channel;type:varchar(64)"`
	IPHash     string `json:"ipHash" gorm:"column:ip_hash;type:varchar(64)"`
	Converted  bool   `json:"converted" gorm:"column:converted;default:false"`

	gorm.Model
}

func (ShareLogModel) TableName() string { return "member_share_log" }
