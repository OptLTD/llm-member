package model

import (
	"time"

	"gorm.io/gorm"
)

// DeviceModel 用户绑定设备
type DeviceModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	UserID     uint64    `json:"userId" gorm:"column:user_id;index;not null"`
	DeviceID   string    `json:"deviceId" gorm:"column:device_id;type:varchar(128);index;not null"`
	DeviceName string    `json:"deviceName" gorm:"column:device_name;type:varchar(256)"`
	Platform   string    `json:"platform" gorm:"column:platform;type:varchar(64)"`
	AppVersion string    `json:"appVersion" gorm:"column:app_version;type:varchar(32)"`
	LastSeenAt time.Time `json:"lastSeenAt" gorm:"column:last_seen_at"`

	gorm.Model
}

func (DeviceModel) TableName() string { return "member_device" }

type BindDeviceRequest struct {
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName,omitempty"`
	Platform   string `json:"platform,omitempty"`
	AppVersion string `json:"appVersion,omitempty"`
}
