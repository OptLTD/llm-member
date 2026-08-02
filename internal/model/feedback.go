package model

import (
	"gorm.io/gorm"
)

const (
	FeedbackStatusOpen       = "open"
	FeedbackStatusProcessing = "processing"
	FeedbackStatusClosed     = "closed"
)

// FeedbackModel 客户反馈
type FeedbackModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID     uint64 `json:"userId" gorm:"column:user_id;index"`
	ProjectID  uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	Type       string `json:"type" gorm:"column:type;type:varchar(32);not null"`
	Title      string `json:"title" gorm:"column:title;type:varchar(256)"`
	Content    string `json:"content" gorm:"column:content;type:text;not null"`
	Rating     int    `json:"rating" gorm:"column:rating;default:0"`
	Contact    string `json:"contact" gorm:"column:contact;type:varchar(256)"`
	AppVersion string `json:"appVersion" gorm:"column:app_version;type:varchar(32)"`
	Status     string `json:"status" gorm:"column:status;type:varchar(20);default:open"`
	AdminReply string `json:"adminReply" gorm:"column:admin_reply;type:text"`

	gorm.Model
}

func (FeedbackModel) TableName() string { return "member_feedback" }

type SubmitFeedbackRequest struct {
	Type       string `json:"type" binding:"required"`
	Title      string `json:"title,omitempty"`
	Content    string `json:"content" binding:"required"`
	Rating     int    `json:"rating,omitempty"`
	Contact    string `json:"contact,omitempty"`
	AppVersion string `json:"appVersion,omitempty"`
}

type UpdateFeedbackRequest struct {
	Status     *string `json:"status,omitempty"`
	AdminReply *string `json:"adminReply,omitempty"`
}
