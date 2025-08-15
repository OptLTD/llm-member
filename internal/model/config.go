package model

import (
	"gorm.io/gorm"
)

// ConfigModel 通用配置项
type ConfigModel struct {
	ID uint64 `json:"id" gorm:"primaryKey"`

	Key  string `json:"key" gorm:"type:varchar(64);uniqueIndex;not null"`
	Kind string `json:"kind" gorm:"type:varchar(64);index;not null"`
	Data string `json:"data" gorm:"type:text;not null"`

	gorm.Model
}

func (m ConfigModel) TableName() string {
	return "llm_cfg"
}
