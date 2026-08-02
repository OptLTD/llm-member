package model

import (
	"time"

	"gorm.io/gorm"
)

// LlmLogModel 聊天日志模型
type LlmLogModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID    uint64 `json:"userId" gorm:"column:user_id;index;not null;default:0"`
	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	ChatID    string `json:"chatId" gorm:"column:chat_id;type:varchar(64);index"`
	ProjID    string `json:"projId" gorm:"column:proj_id;type:varchar(64);index;"`
	TheModel  string `json:"model" gorm:"column:model;type:varchar(64);not null"`
	Provider  string `json:"provider" gorm:"column:provider;type:varchar(64);not null"`

	Messages any `json:"messages" gorm:"type:text;serializer:json"`
	Response any `json:"response" gorm:"type:text;serializer:json"`
	AllUsage any `json:"allUsage" gorm:"column:all_usage;type:text;serializer:json"`

	Duration  int64  `json:"duration" gorm:"not null;default:0"`
	Status    string `json:"status" gorm:"type:varchar(10);"`
	ErrorMsg  string `json:"errorMsg" gorm:"column:error_msg;type:varchar(256);"`
	ClientIP  string `json:"clientIp" gorm:"column:client_ip;type:varchar(64);"`
	UserAgent string `json:"userAgent" gorm:"column:user_agent;type:varchar(256);not null"`

	// 添加请求时间字段，用于按月分区查询
	ReqTime time.Time `json:"reqTime" gorm:"column:req_time;index;"`

	User UserModel `json:"user" gorm:"foreignKey:UserID"`

	gorm.Model
}

func (m LlmLogModel) TableName() string {
	return "llm_log"
}
