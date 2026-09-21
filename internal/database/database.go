package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"boilerplate/internal/config"
	"boilerplate/internal/domain"

	"github.com/glebarez/sqlite" // pure-Go sqlite driver, no CGO required
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Open connects to the configured database, pings it and runs migrations.
func Open(cfg config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBDriver {
	case "sqlite":
		if dir := filepath.Dir(cfg.DBDSN); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create sqlite directory: %w", err)
			}
		}
		dialector = sqlite.Open(cfg.DBDSN)
	case "postgres":
		if cfg.DBDSN == "" {
			return nil, fmt.Errorf("BARA_DB_DSN is required when BARA_DB_DRIVER=postgres (example: \"postgres://user:pass@localhost:5432/bara?sslmode=disable\")")
		}
		dialector = postgres.Open(cfg.DBDSN)
	default:
		return nil, fmt.Errorf("unsupported BARA_DB_DRIVER %q (use \"sqlite\" or \"postgres\")", cfg.DBDriver)
	}

	logLevel := gormlogger.Warn
	if cfg.IsDev() {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.DBDriver == "postgres" {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Session{},
		&domain.PasswordResetToken{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return db, nil
}
