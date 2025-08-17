package service

import (
	"encoding/json"
	"fmt"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"
	"llm-member/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SetupService struct {
	db *gorm.DB

	cfg *config.Config
}

func NewSetupService() *SetupService {
	return &SetupService{db: config.GetDB()}
}

func HandleInit() error {
	// 初始化token编码器
	support.InitTokenEncoders()
	
	s := &SetupService{
		db:  config.GetDB(),
		cfg: config.Load(),
	}
	if err := s.autoMigration(); err != nil {
		return err
	}
	if err := s.createDefaultAdmin(); err != nil {
		return err
	}
	if err := s.initDefaultConfigs(); err != nil {
		return err
	}
	return nil
}

// GetConfig 获取配置项
func (s *SetupService) GetConfig(key string) (*model.ConfigModel, error) {
	var config model.ConfigModel
	err := s.db.Where("`key` = ?", key).First(&config).Error
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
	err := s.db.Where("`key` = ?", key).First(&config).Error

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
	return s.db.Where("`key` = ?", key).Delete(&model.ConfigModel{}).Error
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

func (s *SetupService) autoMigration() error {
	// 自动迁移数据库表
	err := s.db.AutoMigrate(
		&model.UserModel{},
		&model.VerifyModel{},
		&model.OrderModel{},
		&model.LlmLogModel{},
		&model.ConfigModel{},
	)
	return err
}

// createDefaultAdmin 创建默认管理员用户
func (s *SetupService) createDefaultAdmin() error {
	var count int64
	if s.db.Model(&model.UserModel{}).
		Where("user_role = ?", "admin").
		Count(&count); count > 0 {
		return nil // 已存在管理员，跳过创建
	}

	// 创建默认管理员
	expired_at := time.Now().AddDate(99, 0, 0)
	admin := model.UserModel{
		DailyLimit: 10000, MonthlyLimit: 100000,
		IsActive: true, Email: s.cfg.Admin.Username,
		Verified: true, Username: s.cfg.Admin.Username,
		UserRole: model.RoleAdmin, ExpiredAt: &expired_at,
	}

	// 生成密码哈希
	pass, cost := []byte(s.cfg.Admin.Password), bcrypt.DefaultCost
	if data, err := bcrypt.GenerateFromPassword(pass, cost); err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	} else {
		admin.Password = string(data)
	}

	// 生成 API Key
	if key, err := support.GenerateAPIKey(); err != nil {
		return fmt.Errorf("failed to generate API key: %v", err)
	} else {
		admin.APIKey = key
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	return nil
}

// initDefaultConfigs 初始化默认配置
func (s *SetupService) initDefaultConfigs() error {
	configs := []model.ConfigModel{
		{Key: "plan.basic", Data: `{"name":"Basic 基础版","price":99,"desc":"适合个人用户和小型项目","features":["100万 tokens/月","基础模型访问","邮件支持","API文档"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.extra", Data: `{"name":"Extra 增强版","price":299,"desc":"适合中小企业和开发团队","features":["500万 tokens/月","所有模型访问","优先支持","使用统计分析","自定义限制"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.ultra", Data: `{"name":"Ultra 专业版","price":699,"desc":"适合大型企业和高频使用","features":["2000万 tokens/月","所有模型访问","24/7 专属支持","高级分析报告","SLA保障"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.super", Data: `{"name":"Super 旗舰版","price":1999,"desc":"适合超大型企业和极高频使用","features":["无限 tokens","所有模型访问","专属客户经理","定制化服务","最高优先级支持","企业级SLA"],"enabled":true}`, Kind: "plan"},
	}

	for _, config := range configs {
		var existing model.ConfigModel
		err := s.db.Where("`key` = ?", config.Key).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err = s.db.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create config %s: %v", config.Key, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check config %s: %v", config.Key, err)
		}
	}
	return nil
}
