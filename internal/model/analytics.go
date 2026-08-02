package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type EventProps map[string]any

func (p EventProps) Value() (driver.Value, error) {
	if p == nil {
		return "{}", nil
	}
	return json.Marshal(p)
}

func (p *EventProps) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(data, p)
}

// SiteModel 统计站点
type SiteModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	SiteKey   string `json:"siteKey" gorm:"column:site_key;type:varchar(64);index;not null"`
	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	Name      string `json:"name" gorm:"column:name;type:varchar(128);not null"`
	Domains   string `json:"domains" gorm:"column:domains;type:text"` // comma-separated whitelist
	OwnerID   uint64 `json:"ownerId" gorm:"column:owner_id;index"`
	Status    string `json:"status" gorm:"column:status;type:varchar(20);default:active"`
	EmbedCode string `json:"embedCode" gorm:"-"`

	gorm.Model
}

func (SiteModel) TableName() string { return "member_site" }

// EventModel 原始事件
type EventModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	SiteID      uint64     `json:"siteId" gorm:"column:site_id;index"`
	ProjectID   uint64     `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	UserID      uint64     `json:"userId" gorm:"column:user_id;index"`
	EventType   string     `json:"eventType" gorm:"column:event_type;type:varchar(64);index;not null"`
	PageURL     string     `json:"pageUrl" gorm:"column:page_url;type:text"`
	Referrer    string     `json:"referrer" gorm:"column:referrer;type:text"`
	UtmSource   string     `json:"utmSource" gorm:"column:utm_source;type:varchar(128)"`
	UtmMedium   string     `json:"utmMedium" gorm:"column:utm_medium;type:varchar(128)"`
	UtmCampaign string     `json:"utmCampaign" gorm:"column:utm_campaign;type:varchar(128)"`
	VisitorID   string     `json:"visitorId" gorm:"column:visitor_id;type:varchar(128);index"`
	SessionID   string     `json:"sessionId" gorm:"column:session_id;type:varchar(128);index"`
	DeviceID    string     `json:"deviceId" gorm:"column:device_id;type:varchar(128);index"`
	Props       EventProps `json:"props" gorm:"column:props;serializer:json"`
	IPHash      string     `json:"ipHash" gorm:"column:ip_hash;type:varchar(64)"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"column:created_at;index"`
}

func (EventModel) TableName() string { return "member_event" }

// EventDailyModel 日聚合
type EventDailyModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	SiteID    uint64 `json:"siteId" gorm:"column:site_id;index;not null"`
	ProjectID uint64 `json:"projectId" gorm:"column:project_id;index;not null;default:0"`
	Day       string `json:"day" gorm:"column:day;type:varchar(10);index;not null"` // YYYY-MM-DD
	PV        int64  `json:"pv" gorm:"column:pv;default:0"`
	UV        int64  `json:"uv" gorm:"column:uv;default:0"`
	Events    int64  `json:"events" gorm:"column:events;default:0"`
	TopSource string `json:"topSource" gorm:"column:top_source;type:text"` // JSON

	gorm.Model
}

func (EventDailyModel) TableName() string { return "member_event_daily" }

type TrackRequest struct {
	SiteKey     string         `json:"siteKey,omitempty"` // 单实例部署可省略，服务端使用默认站点
	EventType   string         `json:"eventType" binding:"required"`
	PageURL     string         `json:"pageUrl,omitempty"`
	Referrer    string         `json:"referrer,omitempty"`
	UtmSource   string         `json:"utmSource,omitempty"`
	UtmMedium   string         `json:"utmMedium,omitempty"`
	UtmCampaign string         `json:"utmCampaign,omitempty"`
	VisitorID   string         `json:"visitorId,omitempty"`
	SessionID   string         `json:"sessionId,omitempty"`
	DeviceID    string         `json:"deviceId,omitempty"`
	Props       map[string]any `json:"props,omitempty"`
}
