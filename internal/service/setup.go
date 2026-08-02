package service

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	cfgPkg "github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"
	"github.com/OptLTD/member/internal/support"
	relay "github.com/OptLTD/member/llm-relay"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SetupService struct {
	db *gorm.DB

	cfg *cfgPkg.Config
}

func NewSetupService() *SetupService {
	return &SetupService{db: cfgPkg.GetDB()}
}

func HandleInit() error {
	s := &SetupService{
		db:  cfgPkg.GetDB(),
		cfg: cfgPkg.Load(),
	}
	// 先清理索引（旧单列唯一索引 + 旧复合索引），再 AutoMigrate 重建
	if err := s.dropIndexes(); err != nil {
		log.Println("[SETUP] drop indexes:", err)
	}
	if err := s.autoMigration(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrDatabaseMigrationFailed, err)
	}
	// 用原生 SQL 建按项目复合唯一索引（gorm AutoMigrate 不负责唯一性）
	if err := s.ensureUniqueIndexes(); err != nil {
		log.Println("[SETUP] ensure unique indexes:", err)
	}

	defaultProject, err := NewProjectService().GetDefault()
	if err != nil {
		return fmt.Errorf("failed to init default project: %v", err)
	}

	// 旧库迁移：把旧数据归入默认项目
	if err := s.migrateProjectData(defaultProject.ID); err != nil {
		return fmt.Errorf("failed to migrate project data: %v", err)
	}

	if err := s.initDefaultConfigs(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrInitDefaultConfigsFailed, err)
	}

	if err := s.loadModulesConfig(); err != nil {
		return fmt.Errorf("failed to load modules config: %v", err)
	}

	if err := s.createDefaultAdmin(); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrCreateDefaultAdminFailed, err)
	}
	if err := s.migrateQuotaFromUser(); err != nil {
		log.Println("[SETUP] migrate quota failed:", err)
	}
	if err := cfgPkg.InitRedis(s.cfg.Redis); err != nil {
		log.Println("Failed to initialize redis:", err)
	}
	if err := EnsureDefaultSite(); err != nil {
		log.Println("Failed to initialize default analytics site:", err)
	}
	return nil
}

// GetConfig 获取配置项（按项目）
func (s *SetupService) GetConfig(projectID uint64, key string) (*model.ConfigModel, error) {
	var config model.ConfigModel
	err := s.db.Where("project_id = ? AND "+cfgPkg.QuoteIdent(s.db, "key")+" = ?", projectID, key).First(&config).Error
	return &config, err
}

// GetByKind 获取指定分类的所有配置（按项目）
func (s *SetupService) GetByKind(projectID uint64, kind string) ([]model.ConfigModel, error) {
	var configs []model.ConfigModel
	err := s.db.Where("project_id = ? AND kind = ?", projectID, kind).Find(&configs).Error
	return configs, err
}

// GetConfigValue 获取配置值（按项目）
func (s *SetupService) GetConfigValue(projectID uint64, key string) (string, error) {
	config, err := s.GetConfig(projectID, key)
	if err != nil {
		return "", err
	}
	return config.Data, nil
}

// GetAsTarget 获取配置值并解析为JSON（按项目）
func (s *SetupService) GetAsTarget(projectID uint64, key string, target any) error {
	value, err := s.GetConfigValue(projectID, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}

func (s *SetupService) SetConfig(projectID uint64, config *model.ConfigModel) error {
	var existing model.ConfigModel
	err := s.db.Where("project_id = ? AND "+cfgPkg.QuoteIdent(s.db, "key")+" = ?", projectID, config.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		config.ProjectID = projectID
		return s.db.Create(config).Error
	} else if err != nil {
		return err
	}
	existing.Data = config.Data
	return s.db.Save(&existing).Error
}

// DeleteConfig 删除配置项（按项目）
func (s *SetupService) DeleteConfig(projectID uint64, key string) error {
	return s.db.Where("project_id = ? AND "+cfgPkg.QuoteIdent(s.db, "key")+" = ?", projectID, key).Delete(&model.ConfigModel{}).Error
}

// GetAllConfigs 获取所有配置
func (s *SetupService) GetAllConfigs() ([]model.ConfigModel, error) {
	var configs []model.ConfigModel
	err := s.db.Find(&configs).Error
	return configs, err
}

// GetAllPlans 获取所有付费方案配置
func (s *SetupService) GetAllPlans(projectID uint64) []model.PlanInfo {
	var plans []model.PlanInfo
	keys := []string{"basic", "extra", "ultra", "super"}
	for _, key := range keys {
		plan := model.PlanInfo{Plan: key}
		err := s.GetAsTarget(projectID, "plan."+key, &plan)
		if err != nil {
			continue
		}
		plans = append(plans, plan)
	}
	return plans
}

// GetEnablePlans 获取已启用付费方案配置
func (s *SetupService) GetEnablePlans(projectID uint64) []model.PlanInfo {
	var plans []model.PlanInfo
	keys := []string{"basic", "extra", "ultra", "super"}
	for _, key := range keys {
		plan := model.PlanInfo{Plan: key}
		err := s.GetAsTarget(projectID, "plan."+key, &plan)
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
	models := []any{
		&model.ProjectModel{},
		&model.UserModel{},
		&model.UserQuotaModel{},
		&model.VerifyModel{},
		&model.SessionModel{},
		&model.OrderModel{},
		&model.LlmLogModel{},
		&model.ConfigModel{},
	}
	if cfgPkg.HasModule(cfgPkg.ModuleDevice) {
		models = append(models, &model.DeviceModel{})
	}
	if cfgPkg.HasModule(cfgPkg.ModuleAnalytics) {
		models = append(models, &model.SiteModel{}, &model.EventModel{}, &model.EventDailyModel{})
	}
	if cfgPkg.HasModule(cfgPkg.ModuleCoupon) {
		models = append(models, &model.CouponModel{}, &model.CouponUsageModel{})
	}
	if cfgPkg.HasModule(cfgPkg.ModuleReferral) {
		models = append(models,
			&model.ReferralCodeModel{},
			&model.ReferralModel{},
			&model.CampaignModel{},
			&model.ShareLogModel{},
		)
	}
	if cfgPkg.HasModule(cfgPkg.ModuleSurvey) {
		models = append(models,
			&model.SurveyModel{},
			&model.SurveyQuestionModel{},
			&model.SurveyResponseModel{},
		)
	}
	if cfgPkg.HasModule(cfgPkg.ModuleFeedback) {
		models = append(models, &model.FeedbackModel{})
	}
	return s.db.AutoMigrate(models...)
}

// createDefaultAdmin 创建默认管理员用户
func (s *SetupService) createDefaultAdmin() error {
	username := strings.TrimSpace(s.cfg.Admin.Username)
	password := s.cfg.Admin.Password
	if username == "" || password == "" {
		log.Println("[SETUP] ADMIN_USERNAME / ADMIN_PASSWORD 未配置，跳过创建默认管理员")
		return nil
	}

	var admin model.UserModel
	err := s.db.Where("user_role = ?", model.RoleAdmin).First(&admin).Error
	if err == nil {
		// 修复早期版本在 env 未就绪时创建的空账号
		if strings.TrimSpace(admin.Username) == "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("%w: %v", consts.ErrPasswordEncryptionFailed, err)
			}
			admin.Username = username
			admin.Email = username
			admin.Password = string(hashed)
			if err := s.db.Save(&admin).Error; err != nil {
				return fmt.Errorf("%w: %v", consts.ErrCreateAdminUserFailed, err)
			}
			log.Printf("[SETUP] Repaired default admin account: %s", username)
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// 创建默认管理员
	expire_at := time.Now().AddDate(99, 0, 0)
	admin = model.UserModel{
		IsActive: true, Email: username,
		Verified: true, Username: username,
		UserRole: model.RoleAdmin, ExpireAt: &expire_at,
	}

	// 生成密码哈希
	pass, cost := []byte(password), bcrypt.DefaultCost
	if data, err := bcrypt.GenerateFromPassword(pass, cost); err != nil {
		return fmt.Errorf("%w: %v", consts.ErrPasswordEncryptionFailed, err)
	} else {
		admin.Password = string(data)
	}

	// 生成 API Key（入库存 hash）
	plainKey := ""
	if key, err := support.GenerateAPIKey(); err != nil {
		return fmt.Errorf("failed to generate API key: %v", err)
	} else {
		plainKey = key
		admin.APIKey = support.HashAPIKey(key)
	}

	// build plan limit
	defaultProject, _ := NewProjectService().GetDefault()
	if limit, err := s.GetPlanLimit(defaultProject.ID, model.PlanSuper); err == nil {
		admin.ApiLimit, admin.UserPlan = limit, model.PlanSuper
	} else {
		admin.ApiLimit = relay.DefaultBasicLimit()
		admin.ApiLimit.DailyRequests = 100000
		admin.ApiLimit.MonthlyRequests = 1000000
		log.Println("GetPlanLimit error: ", err)
	}
	admin.ApiUsage = &model.ApiUsage{DayStamp: time.Now().Format("2006-01-02")}

	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrCreateAdminUserFailed, err)
	}
	if err := s.db.Create(&model.UserQuotaModel{
		UserID:   admin.ID,
		ApiUsage: &model.ApiUsage{DayStamp: time.Now().Format("2006-01-02")},
		ApiLimit: admin.ApiLimit,
	}).Error; err != nil {
		return fmt.Errorf("%w: %v", consts.ErrCreateAdminUserFailed, err)
	}
	log.Printf("[SETUP] Default admin created, API Key: %s", plainKey)
	return nil
}

func (s *SetupService) GetPlanLimit(projectID uint64, name model.PayPlan) (*model.ApiLimit, error) {
	plan := &model.PlanInfo{Plan: string(name)}
	if err := s.GetAsTarget(projectID, "plan."+plan.Plan, plan); err != nil {
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
			Brief:      "适合个人用户和小型项目",
			MaxDevices: 1, QuotaType: "tokens",
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
			MaxDevices: 2, QuotaType: "tokens",
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
			MaxDevices: 3, QuotaType: "tokens",
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
			MaxDevices: 0, QuotaType: "tokens",
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

// migrateQuotaFromUser 一次性迁移：为已有用户（llm_user 旧列有配额数据）创建独立配额表记录
// dropIndexes 删除旧版索引，让 AutoMigrate 按新模型重建。
// 包含旧版本的单列全局唯一索引（gorm 默认名 idx_<table>_<column>）以及
// 之前已创建的按项目复合唯一索引；不存在时忽略。
func (s *SetupService) dropIndexes() error {
	drops := []struct {
		model any
		name  string
	}{
		// 旧单列唯一索引
		{&model.ConfigModel{}, "idx_llm_cfg_key"},
		{&model.CouponModel{}, "idx_member_coupon_code"},
		{&model.SiteModel{}, "idx_member_site_site_key"},
		{&model.ReferralCodeModel{}, "idx_member_referral_code_code"},
		{&model.ReferralCodeModel{}, "idx_member_referral_code_user_id"},
		{&model.ReferralModel{}, "idx_member_referral_invitee_id"},
		// 新复合索引（AutoMigrate 会重建）
		{&model.ConfigModel{}, "uidx_cfg_project_key"},
		{&model.CouponModel{}, "uidx_coupon_project_code"},
		{&model.SiteModel{}, "uidx_site_project_key"},
		{&model.ReferralCodeModel{}, "uidx_refcode_project_user"},
		{&model.ReferralCodeModel{}, "uidx_refcode_project_code"},
		{&model.ReferralModel{}, "uidx_ref_project_invitee"},
	}
	for _, d := range drops {
		if err := s.db.Migrator().DropIndex(d.model, d.name); err != nil {
			// 索引不存在（新库）或方言不支持时忽略
			log.Printf("[SETUP] drop index %s ignored: %v", d.name, err)
		}
	}
	return nil
}

// ensureUniqueIndexes 用原生 SQL 创建按项目复合唯一索引（方言兼容）。
// gorm 模型上唯一性已降为普通索引，唯一约束由这里显式保障。
func (s *SetupService) ensureUniqueIndexes() error {
	keyQ := cfgPkg.QuoteIdent(s.db, "key")
	projectQ := cfgPkg.QuoteIdent(s.db, "project_id")
	type idx struct {
		table string
		name  string
		cols  string
	}
	idxs := []idx{
		{"llm_cfg", "uidx_cfg_project_key", projectQ + ", " + keyQ},
		{"member_coupon", "uidx_coupon_project_code", projectQ + ", " + cfgPkg.QuoteIdent(s.db, "code")},
		{"member_site", "uidx_site_project_key", projectQ + ", " + cfgPkg.QuoteIdent(s.db, "site_key")},
		{"member_referral_code", "uidx_refcode_project_user", projectQ + ", " + cfgPkg.QuoteIdent(s.db, "user_id")},
		{"member_referral_code", "uidx_refcode_project_code", projectQ + ", " + cfgPkg.QuoteIdent(s.db, "code")},
		{"member_referral", "uidx_ref_project_invitee", projectQ + ", " + cfgPkg.QuoteIdent(s.db, "invitee_id")},
	}
	for _, i := range idxs {
		sql := "CREATE UNIQUE INDEX IF NOT EXISTS " + cfgPkg.QuoteIdent(s.db, i.name) +
			" ON " + cfgPkg.QuoteIdent(s.db, i.table) + " (" + i.cols + ")"
		if err := s.db.Exec(sql).Error; err != nil {
			log.Printf("[SETUP] create index %s ignored: %v", i.name, err)
		}
	}
	return nil
}

// migrateProjectData 把 project_id=0 的存量数据归入默认项目。
// llm_cfg 仅迁移套餐配置（plan.*），模块开关 app.modules 保持全局（project_id=0）。
func (s *SetupService) migrateProjectData(defaultProjectID uint64) error {
	if defaultProjectID == 0 {
		return nil
	}
	// 通用租户表：全部归入默认项目
	tables := []string{
		"llm_user", "llm_user_quota", "llm_order", "llm_log",
		"member_device", "member_site", "member_event", "member_event_daily",
		"member_coupon", "member_coupon_usage",
		"member_referral_code", "member_referral", "member_campaign", "member_share_log",
		"member_survey", "member_survey_question", "member_survey_response",
		"member_feedback", "llm_verify",
	}
	for _, t := range tables {
		if err := s.db.Exec(
			"UPDATE "+cfgPkg.QuoteIdent(s.db, t)+" SET project_id = ? WHERE project_id = 0 OR project_id IS NULL",
			defaultProjectID,
		).Error; err != nil {
			log.Printf("[SETUP] migrate table %s ignored: %v", t, err)
		}
	}
	// 配置表：仅套餐配置归默认项目，模块开关保持全局
	if err := s.db.Exec(
		"UPDATE "+cfgPkg.QuoteIdent(s.db, "llm_cfg")+" SET project_id = ? WHERE (project_id = 0 OR project_id IS NULL) AND kind = 'plan'",
		defaultProjectID,
	).Error; err != nil {
		log.Printf("[SETUP] migrate llm_cfg plan ignored: %v", err)
	}
	return nil
}

func (s *SetupService) migrateQuotaFromUser() error {
	// 新库无旧列，直接跳过；旧库 llm_user 存在 api_usage/api_limit 时才迁移
	if !s.db.Migrator().HasColumn(&model.UserModel{}, "api_usage") &&
		!s.db.Migrator().HasColumn(&model.UserModel{}, "api_limit") {
		return nil
	}
	// UserModel 的配额字段已是 gorm:"-"，需用匿名结构读取旧列
	var rows []struct {
		ID       uint64
		ApiUsage *model.ApiUsage
		ApiLimit *model.ApiLimit
	}
	if err := s.db.Table("llm_user").
		Select("id", "api_usage", "api_limit").
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, r := range rows {
		var count int64
		if err := s.db.Model(&model.UserQuotaModel{}).
			Where("user_id = ?", r.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		quota := &model.UserQuotaModel{
			UserID:   r.ID,
			ApiUsage: r.ApiUsage,
			ApiLimit: r.ApiLimit,
		}
		if quota.ApiUsage == nil {
			quota.ApiUsage = &model.ApiUsage{DayStamp: time.Now().Format("2006-01-02")}
		}
		if err := s.db.Create(quota).Error; err != nil {
			return err
		}
	}
	return nil
}

// loadModulesConfig 从 DB 加载模块开关；不存在则用 APP_MODULES / 默认值写入
func (s *SetupService) loadModulesConfig() error {
	var row model.ConfigModel
	err := s.db.Where("project_id = 0 AND "+cfgPkg.QuoteIdent(s.db, "key")+" = ?", cfgPkg.ConfigKeyModules).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		seed := ""
		if cfgPkg.AppConfig() != nil {
			seed = cfgPkg.AppConfig().Modules
		}
		if seed == "" {
			seed = strings.Join(cfgPkg.AllModules(), ",")
		}
		cfgPkg.SetModulesCSV(seed)
		row = model.ConfigModel{
			ProjectID: 0,
			Key:       cfgPkg.ConfigKeyModules,
			Kind:      "app",
			Data:      cfgPkg.ModulesJSON(),
		}
		return s.db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	cfgPkg.SetModulesCSV(row.Data)
	return nil
}

// SaveModules 保存模块开关到 DB 并热更新内存
func (s *SetupService) SaveModules(names []string) error {
	cfgPkg.SetModulesList(names)
	return s.SetConfig(0, &model.ConfigModel{
		Key:  cfgPkg.ConfigKeyModules,
		Kind: "app",
		Data: cfgPkg.ModulesJSON(),
	})
}

// initDefaultConfigs 初始化默认配置（为默认项目写入 plans）
func (s *SetupService) initDefaultConfigs() error {
	project, err := NewProjectService().GetDefault()
	if err != nil {
		return err
	}
	return s.InitProjectConfigs(project.ID)
}

// InitProjectConfigs 为指定项目初始化默认 plans（缺失才写入）
func (s *SetupService) InitProjectConfigs(projectID uint64) error {
	var creemConfig = cfgPkg.GetCreemConfig()
	var stripeConfig = cfgPkg.GetStripeConfig()
	for _, plan := range s.GetDefaultPlan() {
		var config model.ConfigModel
		var key = "plan." + plan.Plan
		var query = s.db.Where("project_id = ? AND "+cfgPkg.QuoteIdent(s.db, "key")+" = ?", projectID, key)
		var err = query.First(&config).Error
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
		s.setPlanByCreem(&plan, creemConfig)
		s.setPlanByStripe(&plan, stripeConfig)
		config.ProjectID = projectID
		if err := s.db.Create(&config).Error; err != nil {
			return fmt.Errorf("failed to create config %s: %v", config.Key, err)
		}
	}
	return nil
}
func (s *SetupService) setPlanByCreem(plan *model.PlanInfo, cfg *cfgPkg.Creem) {
	if cfg == nil {
		return
	}
	switch plan.Plan {
	case "basic":
		plan.Enabled = cfg.PlanBasicId != ""
	case "extra":
		plan.Enabled = cfg.PlanExtraId != ""
	case "ultra":
		plan.Enabled = cfg.PlanUltraId != ""
	case "super":
		plan.Enabled = cfg.PlanSuperId != ""
	}
}
func (s *SetupService) setPlanByStripe(plan *model.PlanInfo, cfg *cfgPkg.Stripe) {
	if cfg == nil {
		return
	}
	switch plan.Plan {
	case "basic":
		plan.Enabled = cfg.PlanBasicId != ""
	case "extra":
		plan.Enabled = cfg.PlanExtraId != ""
	case "ultra":
		plan.Enabled = cfg.PlanUltraId != ""
	case "super":
		plan.Enabled = cfg.PlanSuperId != ""
	}
}
