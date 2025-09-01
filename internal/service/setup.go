package service

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
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
	s := &SetupService{
		db:  config.GetDB(),
		cfg: config.Load(),
	}
	if err := s.autoMigration(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrDatabaseMigrationFailed, err)
	}

	if err := s.initDefaultConfigs(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrInitDefaultConfigsFailed, err)
	}

	if err := s.createDefaultAdmin(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrCreateDefaultAdminFailed, err)
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

func (s *SetupService) SetConfig(config *model.ConfigModel) error {
	var existing model.ConfigModel
	err := s.db.Where("`key` = ?", config.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return s.db.Create(config).Error
	} else if err != nil {
		return err
	}
	existing.Data = config.Data
	return s.db.Save(existing).Error
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

func (s *SetupService) ParsePlanLimit(plan *model.PlanInfo) (*model.ApiLimit, error) {
	// plan.period 必须是：
	// 	7d, 14d, 1m、1y 等
	// 分别表示：
	// 	7 天, 14 天, 1 月, 1 年
	periodPattern := `^\d+[dmy]$`
	matched, err := regexp.MatchString(periodPattern, plan.Period)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrPeriodRegexError, err)
	}
	if !matched {
		return nil, fmt.Errorf("%w: %s", consts.ErrPlanPeriodFormatError, plan.Period)
	}
	limit := &model.ApiLimit{}

	// 解析 period 并转换为 time.Duration
	periodStr := plan.Period
	unit := periodStr[len(periodStr)-1:]      // 获取最后一个字符 (d/m/y)
	numberStr := periodStr[:len(periodStr)-1] // 获取数字部分
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", consts.ErrInvalidPeriodNumber, err)
	}

	// 根据单位转换为 time.Duration
	switch unit {
	case "d":
		limit.ExpireDays = number
	case "m":
		limit.ExpireDays = 30 // 按30天计算
	case "y":
		limit.ExpireDays = 365 // 按365天计算
	default:
		return nil, fmt.Errorf("%w: %s", consts.ErrUnsupportedPeriodUnit, unit)
	}

	// plan.usage 必须是：
	// 	 1k tokens, 2k requests, 3m projects 等
	// 分别表示：
	// 	 1k tokens, 2k API 请求, 3m 项目调用 等
	usagePattern := `^\d+[km]?\s+(tokens|requests|projects)$`
	matched, err = regexp.MatchString(usagePattern, plan.Usage)
	if err != nil {
		return nil, fmt.Errorf("usage regex error: %v", err)
	}
	if !matched {
		return nil, fmt.Errorf("plan usage format error: %s", plan.Usage)
	}

	// 解析 usage 字符串
	usageParts := strings.Fields(plan.Usage) // 按空格分割
	if len(usageParts) != 2 {
		return nil, fmt.Errorf("invalid usage format: %s", plan.Usage)
	}

	numberWithUnit := usageParts[0] // 如: "1k", "2m", "100"
	usageType := usageParts[1]      // 如: "tokens", "requests", "projects"

	// 解析数量和单位
	var baseNumber uint64
	var multiplier uint64 = 1

	if strings.HasSuffix(numberWithUnit, "k") {
		multiplier = 1000
		numberWithUnit = numberWithUnit[:len(numberWithUnit)-1]
	} else if strings.HasSuffix(numberWithUnit, "m") {
		multiplier = 1000000
		numberWithUnit = numberWithUnit[:len(numberWithUnit)-1]
	}

	parsedNumber, err := strconv.ParseUint(numberWithUnit, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid usage number: %v", err)
	}
	baseNumber = parsedNumber * multiplier

	// 根据使用类型设置相应的限制
	// 这里假设按月计算，如果需要按日计算可以除以30
	switch usageType {
	case "tokens":
		limit.MonthlyTokens = baseNumber
		limit.DailyTokens = baseNumber / 30
	case "requests":
		limit.MonthlyRequests = baseNumber
		limit.DailyRequests = baseNumber / 30
	case "projects":
		limit.MonthlyProjects = baseNumber
		limit.DailyProjects = baseNumber / 30
	default:
		return nil, fmt.Errorf("unsupported usage type: %s", usageType)
	}

	// 设置限制方法
	limit.LimitMethod = usageType
	return limit, nil
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
	expire_at := time.Now().AddDate(99, 0, 0)
	admin := model.UserModel{
		IsActive: true, Email: s.cfg.Admin.Username,
		Verified: true, Username: s.cfg.Admin.Username,
		UserRole: model.RoleAdmin, ExpireAt: &expire_at,
	}

	// 生成密码哈希
	pass, cost := []byte(s.cfg.Admin.Password), bcrypt.DefaultCost
	if data, err := bcrypt.GenerateFromPassword(pass, cost); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrPasswordEncryptionFailed, err)
	} else {
		admin.Password = string(data)
	}

	// 生成 API Key
	if key, err := support.GenerateAPIKey(); err != nil {
		return fmt.Errorf("failed to generate API key: %v", err)
	} else {
		admin.APIKey = key
	}

	// build plan limit
	if limit, err := s.GetPlanLimit(model.PlanSuper); err == nil {
		admin.ApiLimit, admin.UserPlan = limit, model.PlanSuper
	} else {
		log.Println("GetPlanLimit error: ", err)
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrCreateAdminUserFailed, err)
	}
	return nil
}

func (s *SetupService) GetPlanLimit(name model.PayPlan) (*model.ApiLimit, error) {
	plan := &model.PlanInfo{Plan: string(name)}
	if err := s.GetAsTarget("plan."+plan.Plan, plan); err != nil {
		log.Println("[SETUP] GetAsTarget error: ", err.Error())
		return nil, fmt.Errorf("%w: %s", consts.ErrPlanNotFound, plan.Plan)
	}
	if !plan.Enabled {
		log.Println("[SETUP] GetPlanLimit error: ", "not enabled")
		return nil, fmt.Errorf("Plan %v is not enabled", plan.Plan)
	}
	if limit, err := s.ParsePlanLimit(plan); err != nil {
		log.Println("[SETUP] GetPlanLimit error: ", err.Error())
		return nil, fmt.Errorf("%w: %v", consts.ErrParsePlanLimitFailed, err)
	} else {
		return limit, nil
	}
}

func (s *SetupService) GetDefaultPlan() []model.PlanInfo {
	planInfo := []model.PlanInfo{
		{
			Plan: "basic", Name: "体验版",
			Period: "1d", Usage: "1k tokens",
			Brief: "适合个人用户和小型项目",
			Features: []string{
				"100万 tokens/月",
				"基础模型访问",
				"邮件支持",
				"API文档",
			},
		},
		{
			Plan: "extra", Name: "增强版",
			Period: "1m", Usage: "10k tokens",
			Price: 99, Brief: "适合中小企业和开发团队",
			Features: []string{
				"500万 tokens/月",
				"所有模型访问",
				"优先支持",
				"使用统计分析",
				"自定义限制",
			},
		},
		{
			Plan: "ultra", Name: "专业版",
			Period: "1m", Usage: "50k tokens",
			Price: 299, Brief: "适合大型企业和高频使用",
			Features: []string{
				"2000万 tokens/月",
				"所有模型访问",
				"24/7 专属支持",
				"高级分析报告",
				"SLA保障",
			},
		},
		{
			Plan: "super", Name: "旗舰版",
			Period: "1m", Usage: "100m tokens",
			Price: 1999, Brief: "适合超大型企业和极高频使用",
			Features: []string{
				"无限 tokens",
				"所有模型访问",
				"专属客户经理",
				"定制化服务",
				"最高优先级支持",
				"企业级SLA",
			},
		},
	}
	return planInfo
}

// initDefaultConfigs 初始化默认配置
func (s *SetupService) initDefaultConfigs() error {
	for _, plan := range s.GetDefaultPlan() {
		var config model.ConfigModel
		var key = "plan." + plan.Plan
		err := s.db.Where("`key` = ?", key).First(&config).Error
		if err != gorm.ErrRecordNotFound {
			continue
		}

		plan.Enabled = true
		config.Key, config.Kind = key, "plan"
		if data, err := json.Marshal(plan); err != nil {
			continue
		} else {
			config.Data = string(data)
		}
		if err := s.db.Create(&config).Error; err != nil {
			return fmt.Errorf("failed to create config %s: %v", config.Key, err)
		}
	}
	return nil
}
