package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"swiflow-auth/internal/models"
)

func Initialize(dbPath string) (*gorm.DB, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(&models.User{}, &models.ChatLog{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// 创建默认管理员用户
	err = createDefaultAdmin(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create default admin: %v", err)
	}

	return db, nil
}

// generateAPIKey 生成随机 API Key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin(db *gorm.DB) error {
	// 检查是否已存在管理员
	var count int64
	db.Model(&models.User{}).Where("is_admin = ?", true).Count(&count)
	if count > 0 {
		return nil // 已存在管理员，跳过创建
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// 生成 API Key
	apiKey, err := generateAPIKey()
	if err != nil {
		return fmt.Errorf("failed to generate API key: %v", err)
	}

	// 创建默认管理员
	admin := models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		Password:     string(hashedPassword),
		APIKey:       apiKey,
		IsActive:     true,
		IsAdmin:      true,
		DailyLimit:   10000,
		MonthlyLimit: 100000,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	return nil
}