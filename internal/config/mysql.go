package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
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

func GetMySQLDB(cfg *MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name, "utf8mb4",
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}
