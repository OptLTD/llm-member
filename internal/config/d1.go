package config

import (
	"fmt"

	d1 "github.com/intmian/gorm-d1-adapter"
	"github.com/intmian/gorm-d1-adapter/gormd1"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type D1Config struct {
	AccountID  string
	DatabaseID string
	APIToken   string
	ProxyURL   string
	ProxyToken string
}

func GetD1DB(cfg *Config) (*gorm.DB, error) {
	d1cfg := cfg.D1
	if d1cfg == nil {
		return nil, fmt.Errorf("D1 config is missing")
	}

	var logMode = logger.Error
	switch cfg.AppMode {
	case "debug":
		logMode = logger.Info
	case "test":
		logMode = logger.Warn
	}

	driverCfg := d1.Config{}
	if d1cfg.ProxyURL != "" {
		driverCfg.Mode = d1.ExecutorModeWorker
		driverCfg.WorkerEndpoint = d1cfg.ProxyURL
		driverCfg.WorkerToken = d1cfg.ProxyToken
	} else {
		driverCfg.Mode = d1.ExecutorModeREST
		driverCfg.AccountID = d1cfg.AccountID
		driverCfg.DatabaseID = d1cfg.DatabaseID
		driverCfg.APIToken = d1cfg.APIToken
	}
	if err := driverCfg.Validate(); err != nil {
		return nil, fmt.Errorf("D1 config invalid: %w", err)
	}

	db, err := gorm.Open(gormd1.OpenConfig(driverCfg), &gorm.Config{
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
