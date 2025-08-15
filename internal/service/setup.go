package service

import (
	"encoding/json"

	"llm-member/internal/config"
	"llm-member/internal/model"

	"gorm.io/gorm"
)

type SetupService struct {
	db *gorm.DB
}

func NewSetupService() *SetupService {
	return &SetupService{db: config.GetDB()}
}

// GetConfig 获取配置项
func (s *SetupService) GetConfig(key string) (*model.ConfigModel, error) {
	var config model.ConfigModel
	err := s.db.Where("key = ?", key).First(&config).Error
	return &config, err
}

// GetByKind 获取指定分类的所有配置
func (s *SetupService) GetByKind(kind string) ([]model.ConfigModel, error) {
	var configs []model.ConfigModel
	err := s.db.Where("kind = ?", kind).Find(&configs).Error
	return configs, err
}

// GetConfigValue 获取配置值
func (s *SetupService) GetConfigValue(key string) (string, error) {
	config, err := s.GetConfig(key)
	if err != nil {
		return "", err
	}
	return config.Data, nil
}

// GetAsTarget 获取配置值并解析为JSON
func (s *SetupService) GetAsTarget(key string, target any) error {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}

// SetConfig 设置配置项
func (s *SetupService) SetConfig(key, data, kind string) error {
	var config model.ConfigModel
	err := s.db.Where("key = ?", key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新配置
		config = model.ConfigModel{
			Key: key, Kind: kind, Data: data,
		}
		return s.db.Create(&config).Error
	} else if err != nil {
		return err
	}

	// 更新现有配置
	config.Data = data
	config.Kind = kind
	return s.db.Save(&config).Error
}

// DeleteConfig 删除配置项
func (s *SetupService) DeleteConfig(key string) error {
	return s.db.Where("key = ?", key).Delete(&model.ConfigModel{}).Error
}

// GetAllConfigs 获取所有配置
func (s *SetupService) GetAllConfigs() ([]model.ConfigModel, error) {
	var configs []model.ConfigModel
	err := s.db.Find(&configs).Error
	return configs, err
}

// GetAllPlans 获取所有付费方案配置
func (s *SetupService) GetAllPlans() []model.PlanInfo {
	var plans []model.PlanInfo
	keys := []string{"basic", "extra", "ultra", "super"}
	for _, key := range keys {
		plan := model.PlanInfo{Plan: key}
		err := s.GetAsTarget("plan."+key, &plan)
		if err != nil {
			continue
		}
		plans = append(plans, plan)
	}
	return plans
}

// GetEnablePlans 获取已启用付费方案配置
func (s *SetupService) GetEnablePlans() []model.PlanInfo {
	var plans []model.PlanInfo
	keys := []string{"basic", "extra", "ultra", "super"}
	for _, key := range keys {
		plan := model.PlanInfo{Plan: key}
		err := s.GetAsTarget("plan."+key, &plan)
		if err != nil || !plan.Enabled {
			continue
		}
		plans = append(plans, plan)
	}
	return plans
}
