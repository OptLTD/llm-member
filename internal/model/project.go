package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

// ProjectLLM 项目独立的 LLM 上游配置（OpenAI-compatible）
type ProjectLLM struct {
	Provider     string   `json:"provider,omitempty"`
	APIKey       string   `json:"apiKey,omitempty"`
	BaseURL      string   `json:"baseUrl,omitempty"`
	DefaultModel string   `json:"defaultModel,omitempty"`
	Models       []string `json:"models,omitempty"`
}

// ProjectSettings 项目独立配置（JSON 存储）
type ProjectSettings struct {
	AppName      string      `json:"appName,omitempty"`      // 项目展示名
	AuthCallback string      `json:"authCallback,omitempty"` // 认证回调 URL（覆盖全局 AUTH_CALLBACK）
	LLM          *ProjectLLM `json:"llm,omitempty"`          // LLM 上游（覆盖全局 LLM_*）
}

// ProjectModel 项目（租户）
type ProjectModel struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	Name   string `json:"name" gorm:"type:varchar(128);not null"`
	Slug   string `json:"slug" gorm:"type:varchar(64);uniqueIndex;not null"`
	Status string `json:"status" gorm:"type:varchar(20);default:active"` // active | disabled

	Settings string `json:"settings" gorm:"type:text"` // ProjectSettings JSON

	gorm.Model
}

func (ProjectModel) TableName() string { return "member_project" }

// GetSettings 解析项目设置
func (p *ProjectModel) GetSettings() ProjectSettings {
	var s ProjectSettings
	if p.Settings != "" {
		_ = json.Unmarshal([]byte(p.Settings), &s)
	}
	return s
}

// SetSettings 写入项目设置
func (p *ProjectModel) SetSettings(s ProjectSettings) {
	if b, err := json.Marshal(s); err == nil {
		p.Settings = string(b)
	}
}

// ProjectSettingsResponse 管理端展示用
type ProjectSettingsResponse struct {
	AppName      string      `json:"appName,omitempty"`
	AuthCallback string      `json:"authCallback,omitempty"`
	LLM          *ProjectLLM `json:"llm,omitempty"`
}

func ToProjectSettingsResponse(s ProjectSettings) ProjectSettingsResponse {
	return ProjectSettingsResponse{
		AppName:      s.AppName,
		AuthCallback: s.AuthCallback,
		LLM:          s.LLM,
	}
}
