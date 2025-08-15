package config

import (
	"fmt"
	"os"
	"path/filepath"

	"llm-member/internal/model"
	"llm-member/internal/support"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var globalDB *gorm.DB

// GetDB 获取全局数据库实例
func GetDB() *gorm.DB {
	return globalDB
}

func InitDB(path string) error {
	// 确保数据库目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&model.UserModel{},
		&model.OrderModel{},
		&model.LlmLogModel{},
		&model.ConfigModel{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 创建默认管理员用户
	err = createDefaultAdmin(db)
	if err != nil {
		return fmt.Errorf("failed to create default admin: %v", err)
	}

	// 初始化默认配置
	err = initDefaultConfigs(db)
	if err != nil {
		return fmt.Errorf("failed to init default configs: %v", err)
	}

	globalDB = db
	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin(db *gorm.DB) error {
	var count int64
	if db.Model(&model.UserModel{}).
		Where("curr_role = ?", "admin").
		Count(&count); count > 0 {
		return nil // 已存在管理员，跳过创建
	}

	// 创建默认管理员
	admin := model.UserModel{
		Email: "admin@demo.com", Username: "admin",
		IsActive: true, CurrRole: model.RoleAdmin,
		DailyLimit: 10000, MonthlyLimit: 100000,
	}

	// 生成密码哈希
	pass, cost := []byte("admin123"), bcrypt.DefaultCost
	if pass, err := bcrypt.GenerateFromPassword(pass, cost); err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	} else {
		admin.Password = string(pass)
	}

	// 生成 API Key
	if key, err := support.GenerateAPIKey(); err != nil {
		return fmt.Errorf("failed to generate API key: %v", err)
	} else {
		admin.APIKey = key
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}
	return nil
}

// initDefaultConfigs 初始化默认配置
func initDefaultConfigs(db *gorm.DB) error {
	// 默认配置项
	configs := []model.ConfigModel{
		// 付费方案配置 - 作为JSON对象存储
		{Key: "plan.basic", Data: `{"name":"Basic 基础版","price":99,"desc":"适合个人用户和小型项目","features":["100万 tokens/月","基础模型访问","邮件支持","API文档"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.extra", Data: `{"name":"Extra 增强版","price":299,"desc":"适合中小企业和开发团队","features":["500万 tokens/月","所有模型访问","优先支持","使用统计分析","自定义限制"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.ultra", Data: `{"name":"Ultra 专业版","price":699,"desc":"适合大型企业和高频使用","features":["2000万 tokens/月","所有模型访问","24/7 专属支持","高级分析报告","SLA保障"],"enabled":true}`, Kind: "plan"},
		{Key: "plan.super", Data: `{"name":"Super 旗舰版","price":1999,"desc":"适合超大型企业和极高频使用","features":["无限 tokens","所有模型访问","专属客户经理","定制化服务","最高优先级支持","企业级SLA"],"enabled":true}`, Kind: "plan"},
	}

	// 批量创建配置，检查是否已存在
	for _, config := range configs {
		var existing model.ConfigModel
		err := db.Where("key = ?", config.Key).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err = db.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create config %s: %v", config.Key, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check config %s: %v", config.Key, err)
		}
		// 如果配置已存在，跳过创建
	}

	return nil
}
