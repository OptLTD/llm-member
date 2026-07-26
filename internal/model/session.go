package model

import (
	"time"

	"gorm.io/gorm"
)

// SessionModel 登录会话（Bearer Token）
type SessionModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Token    string    `json:"-" gorm:"column:token;type:varchar(64);uniqueIndex;not null"`
	UserID   uint64    `json:"userId" gorm:"column:user_id;index;not null"`
	IsAdmin  bool      `json:"isAdmin" gorm:"column:is_admin;default:false"`
	ExpireAt time.Time `json:"expireAt" gorm:"column:expire_at;index;not null"`

	gorm.Model
}

func (SessionModel) TableName() string {
	return "member_session"
}
