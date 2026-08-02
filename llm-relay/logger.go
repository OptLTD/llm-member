package relay

import (
	"time"

	"gorm.io/gorm"
)

// LogRecord 中继日志的 gorm 模型，映射 llm_log 表。
// relay 只负责写入；查询/统计等业务侧需求由应用侧自行实现。
type LogRecord struct {
	ID uint64 `gorm:"primarykey"`

	ProjectID uint64 `gorm:"column:project_id;index;not null;default:0"`
	UserID    uint64 `gorm:"column:user_id;index;not null;default:0"`
	ChatID    string `gorm:"column:chat_id;type:varchar(64);index"`
	ProjID    string `gorm:"column:proj_id;type:varchar(64);index"`
	TheModel  string `gorm:"column:model;type:varchar(64);not null"`
	Provider  string `gorm:"column:provider;type:varchar(64);not null"`

	Messages any `gorm:"type:text;serializer:json"`
	Response any `gorm:"type:text;serializer:json"`
	AllUsage any `gorm:"column:all_usage;type:text;serializer:json"`

	Duration  int64  `gorm:"not null;default:0"`
	Status    string `gorm:"type:varchar(10)"`
	ErrorMsg  string `gorm:"column:error_msg;type:varchar(256)"`
	ClientIP  string `gorm:"column:client_ip;type:varchar(64)"`
	UserAgent string `gorm:"column:user_agent;type:varchar(256);not null"`

	// 请求时间，用于按月分区查询
	ReqTime time.Time `gorm:"column:req_time;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (LogRecord) TableName() string { return "llm_log" }

// GormLogger 基于 gorm 的日志记录器，实现 Logger 接口。
type GormLogger struct {
	db *gorm.DB
}

// NewGormLogger 创建日志记录器
func NewGormLogger(db *gorm.DB) *GormLogger {
	return &GormLogger{db: db}
}

// CreateLog 将中继日志条目持久化到 llm_log 表
func (l *GormLogger) CreateLog(entry *LogEntry) error {
	record := &LogRecord{
		ProjectID: entry.ProjectID,
		UserID:    entry.UserID,
		ChatID:    entry.ChatID,
		ProjID:    entry.ProjID,
		TheModel:  entry.Model,
		Provider:  entry.Provider,
		Messages:  entry.Messages,
		Response:  entry.Response,
		AllUsage:  entry.AllUsage,
		Duration:  entry.Duration,
		Status:    entry.Status,
		ErrorMsg:  entry.ErrorMsg,
		ClientIP:  entry.ClientIP,
		UserAgent: entry.UserAgent,
		ReqTime:   entry.ReqTime,
	}
	return l.db.Create(record).Error
}
