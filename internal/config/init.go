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
	Version string
	AppHost string
	AppMode string
	Storage string

	DBDriver string
	Modules  string

	Admin    *AdminInfo
	Postgres *PostgresConfig
	D1       *D1Config
	Redis    *RedisConfig
}

type AdminInfo struct {
	Username string
	Password string
}

func Load() *Config {
	cfg := &Config{
		AppMode: getEnv("APP_MODE", "test"),
		AppPort: getEnv("APP_PORT", "8080"),
		AppName: getEnv("APP_NAME", "Demo"),
		AppDesc: getEnv("APP_DESC", "Desc"),
		AppHost: getEnv("APP_HOST", "localhost"),
		Storage:  getEnv("DATA_PATH", "./data"),
		Version:  getEnv("APP_VERSION", "v1.0.0"),
		DBDriver: getEnv("DB_DRIVER", "sqlite"),
		Modules: getEnv("APP_MODULES",
			"member,device,analytics,coupon,referral,survey,feedback,llm-relay"),

		// PostgreSQL配置
		Postgres: &PostgresConfig{
			Host: getEnv("POSTGRES_HOST", "localhost"),
			Port: getEnv("POSTGRES_PORT", "5432"),
			User: getEnv("POSTGRES_USER", ""),
			Pass: getEnv("POSTGRES_PASS", ""),
			Name: getEnv("POSTGRES_NAME", ""),
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
	if url := getEnv("REDIS_URL", ""); url != "" {
		cfg.Redis = &RedisConfig{Url: url}
	}
	if dsn := getEnv("POSTGRES_DSN", ""); dsn != "" {
		cfg.Postgres = &PostgresConfig{DSN: dsn}
	}
	if dsn := getEnv("DATABASE_URL", ""); dsn != "" {
		cfg.Postgres = &PostgresConfig{DSN: dsn}
	}
	cfg.D1 = &D1Config{
		AccountID:  getEnv("D1_ACCOUNT_ID", ""),
		DatabaseID: getEnv("D1_DATABASE_ID", ""),
		APIToken:   getEnv("D1_API_TOKEN", ""),
		ProxyURL:   getEnv("D1_PROXY_ENDPOINT", ""),
		ProxyToken: getEnv("D1_PROXY_TOKEN", ""),
	}
	setGlobalConfig(cfg)
	return cfg
}

var globalDB *gorm.DB

func InitDB(cfg *Config) error {
	var db *gorm.DB
	var err error
	switch cfg.DBDriver {
	case "postgres":
		if cfg.Postgres.Pass != "" || cfg.Postgres.DSN != "" {
			db, err = GetPostgresDB(cfg.Postgres)
		} else {
			return fmt.Errorf("DB_DRIVER=postgres but PostgreSQL is not configured")
		}
	case "d1":
		db, err = GetD1DB(cfg)
	default:
		db, err = GetSQLiteDB(cfg)
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
	if cfg == nil {
		return nil
	}
	if cfg.Host == "" && cfg.Url == "" {
		return nil
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
