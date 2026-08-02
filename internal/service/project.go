package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/model"
	relay "github.com/OptLTD/member/llm-relay"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// CtxProjectID 上下文中的项目 ID
	CtxProjectID = "ProjectID"
	// CtxProject 上下文中的项目对象
	CtxProject = "Project"
)

// ProjectService 项目（租户）服务
type ProjectService struct {
	db *gorm.DB
}

func NewProjectService() *ProjectService {
	return &ProjectService{db: config.GetDB()}
}

// DefaultSlug 默认项目 slug
const DefaultSlug = "default"

// GetDefault 获取默认项目（不存在则创建）
func (s *ProjectService) GetDefault() (*model.ProjectModel, error) {
	var p model.ProjectModel
	err := s.db.Where("slug = ?", DefaultSlug).First(&p).Error
	if err == nil {
		return &p, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	p = model.ProjectModel{Name: "默认项目", Slug: DefaultSlug, Status: "active"}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByID 获取项目
func (s *ProjectService) GetByID(id uint64) (*model.ProjectModel, error) {
	var p model.ProjectModel
	if err := s.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// GetBySlug 获取项目
func (s *ProjectService) GetBySlug(slug string) (*model.ProjectModel, error) {
	var p model.ProjectModel
	if err := s.db.Where("slug = ?", slug).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// List 项目列表（active 优先）
func (s *ProjectService) List() ([]model.ProjectModel, error) {
	var list []model.ProjectModel
	err := s.db.Order("id asc").Find(&list).Error
	return list, err
}

// Create 创建项目
func (s *ProjectService) Create(p *model.ProjectModel) error {
	p.Slug = strings.TrimSpace(strings.ToLower(p.Slug))
	if p.Slug == "" {
		return errors.New("slug 不能为空")
	}
	if p.Status == "" {
		p.Status = "active"
	}
	return s.db.Create(p).Error
}

// Update 更新项目（不含 settings 由调用方管理）
func (s *ProjectService) Update(id uint64, p *model.ProjectModel) error {
	return s.db.Model(&model.ProjectModel{}).Where("id = ?", id).Updates(map[string]any{
		"name":   p.Name,
		"slug":   strings.TrimSpace(strings.ToLower(p.Slug)),
		"status": p.Status,
	}).Error
}

// Delete 删除项目（软删除）
func (s *ProjectService) Delete(id uint64) error {
	return s.db.Delete(&model.ProjectModel{}, id).Error
}

// SaveSettings 保存项目设置
func (s *ProjectService) SaveSettings(id uint64, settings model.ProjectSettings) error {
	p, err := s.GetByID(id)
	if err != nil {
		return err
	}
	p.SetSettings(settings)
	return s.db.Save(p).Error
}

// Resolve 从请求解析项目：
//  1. X-Project-Id header（数字）
//  2. X-Project-Slug header
//  3. query project（id 或 slug）
//  4. 默认项目
func (s *ProjectService) Resolve(c *gin.Context) (*model.ProjectModel, error) {
	if v := c.GetHeader("X-Project-Id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			if p, err := s.GetByID(id); err == nil {
				return p, nil
			}
		}
	}
	if v := c.GetHeader("X-Project-Slug"); v != "" {
		if p, err := s.GetBySlug(v); err == nil {
			return p, nil
		}
	}
	if v := c.Query("project"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			if p, err := s.GetByID(id); err == nil {
				return p, nil
			}
		}
		if p, err := s.GetBySlug(v); err == nil {
			return p, nil
		}
	}
	return s.GetDefault()
}

// ProjectMiddleware 解析项目并写入 gin context（供管理端与公共端点使用）
func ProjectMiddleware(ps *ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, err := ps.Resolve(c)
		if err != nil {
			p, _ = ps.GetDefault()
		}
		if p != nil {
			c.Set(CtxProjectID, p.ID)
			c.Set(CtxProject, p)
		}
		c.Next()
	}
}

// ProjectID 从 context 取项目 ID
func ProjectID(c *gin.Context) uint64 {
	if v, ok := c.Get(CtxProjectID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

// Project 从 context 取项目对象
func Project(c *gin.Context) *model.ProjectModel {
	if v, ok := c.Get(CtxProject); ok {
		if p, ok := v.(*model.ProjectModel); ok {
			return p
		}
	}
	return nil
}

// LLMConfig 项目 LLM 配置（有项目级配置用项目级，否则环境变量）
func (s *ProjectService) LLMConfig(projectID uint64) *relay.LLMConfig {
	if p, err := s.GetByID(projectID); err == nil {
		st := p.GetSettings()
		if st.LLM != nil && st.LLM.APIKey != "" {
			cfg := &relay.LLMConfig{
				Provider:     st.LLM.Provider,
				APIKey:       st.LLM.APIKey,
				BaseURL:      st.LLM.BaseURL,
				DefaultModel: st.LLM.DefaultModel,
				Models:       st.LLM.Models,
			}
			if cfg.Provider == "" {
				cfg.Provider = "openai"
			}
			if cfg.BaseURL == "" {
				cfg.BaseURL = "https://api.openai.com/v1"
			}
			if len(cfg.Models) == 0 && cfg.DefaultModel != "" {
				cfg.Models = []string{cfg.DefaultModel}
			}
			return cfg
		}
	}
	return relay.ConfigFromEnv()
}

// ToProjectSettingsResponse 转换为管理端展示结构
func ToProjectSettingsResponse(st model.ProjectSettings) model.ProjectSettingsResponse {
	return model.ProjectSettingsResponse{
		AppName:      st.AppName,
		AuthCallback: st.AuthCallback,
		LLM:          st.LLM,
	}
}

// AuthCallback 项目认证回调（有项目级配置用项目级，否则全局）
func (s *ProjectService) AuthCallback(projectID uint64) string {
	if p, err := s.GetByID(projectID); err == nil {
		if cb := p.GetSettings().AuthCallback; cb != "" {
			return cb
		}
	}
	return config.GetAuthCallback()
}
