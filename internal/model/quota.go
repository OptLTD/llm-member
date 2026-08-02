package model

import (
	"gorm.io/gorm"
)

// UserQuotaModel 用户用量/配额独立表
type UserQuotaModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID    uint64    `json:"userId" gorm:"column:user_id;index;not null"`
	ProjectID uint64    `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	ApiUsage  *ApiUsage `json:"apiUsage" gorm:"column:api_usage;serializer:json"`
	ApiLimit  *ApiLimit `json:"apiLimit" gorm:"column:api_limit;serializer:json"`

	gorm.Model
}

func (UserQuotaModel) TableName() string { return "llm_user_quota" }
