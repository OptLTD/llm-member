package model

import (
	"gorm.io/gorm"
)

// ConfigModel 通用配置项
type ConfigModel struct {
	ID uint64 `json:"id" gorm:"primaryKey"`

	Key  string `json:"key" gorm:"uniqueIndex;not null"`
	Kind string `json:"kind" gorm:"index"`
	Data string `json:"data" gorm:"type:text"`

	gorm.Model
}

func (m ConfigModel) TableName() string {
	return "llm_cfg"
}
