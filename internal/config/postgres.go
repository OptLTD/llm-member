package config

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresConfig struct {
	DSN string // DSN / DATABASE_URL

	Host string // 主机地址
	Port string // 端口
	User string // 用户名
	Pass string // 密码
	Name string // 数据库名
}

// GetDB 获取全局数据库实例
func GetDB() *gorm.DB {
	return globalDB
}

func configureDBPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return nil
}

func GetPostgresDB(cfg *PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Pass, cfg.Name, cfg.Port,
	)
	if cfg.DSN != "" {
		dsn = cfg.DSN
	}
	var logMode = logger.Error
	switch getEnv("APP_MODE", "test") {
	case "debug":
		logMode = logger.Info
	case "test":
		logMode = logger.Warn
	default:
		logMode = logger.Error
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return nil, err
	}
	if err := configureDBPool(db); err != nil {
		return nil, err
	}
	return db, nil
}
