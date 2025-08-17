package config

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetSQLiteDB(cfg *Config) (*gorm.DB, error) {
	if err := os.MkdirAll(cfg.Storage, 0755); err != nil {
		return nil, err
	}

	filename := filepath.Join(cfg.Storage, "app.db")
	return gorm.Open(sqlite.Open(filename), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}
