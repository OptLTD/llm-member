package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Config struct {
	AppPort string
	AppName string
	AppDesc string
	AppHost string
	GinMode string
	DBPath  string
	LogFile string

	Admin *AdminInfo
	MySQL *MySQLConfig
	Redis *RedisConfig
}

type AdminInfo struct {
	Username string
	Password string
}

// MySQLConfig MySQL数据库配置

// MailConfig 邮件配置

func Load() *Config {
	cfg := &Config{
		GinMode: getEnv("GIN_MODE", "test"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "Demo"),
		AppDesc: getEnv("APP_DESC", "Desc"),
		AppHost: getEnv("APP_HOST", "localhost"),
		DBPath:  getEnv("DB_PATH", "storage/app.db"),
		LogFile: getEnv("LOG_FILE", "storage/app.log"),

		// MySQL配置
		MySQL: &MySQLConfig{
			Host: getEnv("MYSQL_HOST", "localhost"),
			Port: getEnv("MYSQL_PORT", "3306"),
			User: getEnv("MYSQL_USER", ""),
			Pass: getEnv("MYSQL_PASS", ""),
			Name: getEnv("MYSQL_NAME", ""),
		},
		Redis: &RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			Pass: getEnv("REDIS_PASS", ""),
			DB:   getEnv("REDIS_DB", "0"),
		},
		Admin: &AdminInfo{
			Username: getEnv("ADMIN_USERNAME", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
		},
	}
	return cfg
}

var globalDB *gorm.DB

func InitDB(cfg *Config) error {
	var db *gorm.DB
	var err error
	// 检查是否配置了MySQL
	if cfg.MySQL.User == "" || cfg.MySQL.Name == "" {
		db, err = GetSQLiteDB(cfg)
	} else {
		db, err = GetMySQLDB(cfg.MySQL)
	}
	if err != nil {
		return err
	}
	globalDB = db
	return nil
}

var globalRedis *redis.Client

// GetRedis 获取全局Redis实例
func GetRedis() *redis.Client {
	return globalRedis
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *RedisConfig) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.DB == "" {
		return fmt.Errorf("redis config is incomplete")
	}
	client, err := GetRedisClient(cfg)
	if err != nil {
		return err
	}
	globalRedis = client
	return nil
}

// InitLogger 配置日志系统
func InitLogger(cfg *Config) io.Writer {
	// 创建必要的目录（基于日志文件和数据库文件路径）
	logDir := filepath.Dir(cfg.LogFile)
	if logDir != "." && logDir != "" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Fatal("Failed to create log directory:", err)
		}
	}

	// 配置日志文件
	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	return multiWriter
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
