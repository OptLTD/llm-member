package model

import (
	"gorm.io/gorm"
)

// ConfigModel 通用配置项
type ConfigModel struct {
	ID uint64 `json:"id" gorm:"primaryKey"`

	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	Key       string `json:"key" gorm:"type:varchar(64);index;not null"`
	Kind      string `json:"kind" gorm:"type:varchar(64);index;not null"`
	Data      string `json:"data" gorm:"type:text;not null"`

	gorm.Model
}

func (m ConfigModel) TableName() string {
	return "llm_cfg"
}
