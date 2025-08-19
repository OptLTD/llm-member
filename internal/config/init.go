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
	AppMode string
	Storage string

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
		AppMode: getEnv("APP_MODE", "test"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "Demo"),
		AppDesc: getEnv("APP_DESC", "Desc"),
		AppHost: getEnv("APP_HOST", "localhost"),
		Storage: getEnv("DATA_PATH", "./storage"),

		// MySQL配置
		MySQL: &MySQLConfig{
			Host: getEnv("MYSQL_HOST", "localhost"),
			Port: getEnv("MYSQL_PORT", "3306"),
			User: getEnv("MYSQL_USER", ""),
			Pass: getEnv("MYSQL_PASS", ""),
			Name: getEnv("MYSQL_NAME", ""),
		},
		Admin: &AdminInfo{
			Username: getEnv("ADMIN_USERNAME", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
		},
	}
	if getEnv("REDIS_HOST", "") != "" {
		cfg.Redis = &RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
			Pass: getEnv("REDIS_PASS", ""),
			DB:   getEnv("REDIS_DB", "0"),
		}
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
	if cfg == nil || cfg.Host == "" || cfg.Port == "" {
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
	if err := os.MkdirAll(cfg.Storage, 0755); err != nil {
		log.Fatal("Failed to create log directory:", err)
	}

	flag := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	filename := filepath.Join(cfg.Storage, "app.log")
	logFile, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	return multiWriter
}

// GetAuthCallback 获取AUTH_CALLBACK配置
func GetAuthCallback() string {
	return getEnv("AUTH_CALLBACK", "")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
