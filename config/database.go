package config

import (
	"go-grst-boilerplate/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewDatabase(cfg *Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := database.New(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Name:     cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
		Timezone: cfg.DBTimezone,
		
		// Performance optimizations
		PrepareStmt:            cfg.DBPrepareStmt,
		SkipDefaultTransaction: cfg.DBSkipDefaultTransaction,
	}, log)
	if err != nil {
		return nil, err
	}

	if err := database.AutoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
