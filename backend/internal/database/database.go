package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"harmony/internal/models"
)

type Database struct {
	DB *gorm.DB
}

type Config struct {
	Path        string
	MaxOpenConn int
	MaxIdleConn int
	MaxLifetime time.Duration
}

func DefaultConfig() Config {
	return Config{
		Path:        "./data/harmony.db",
		MaxOpenConn: 10,
		MaxIdleConn: 5,
		MaxLifetime: time.Hour,
	}
}

func New(cfg Config) (*Database, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	slog.Info("database connection established", "path", cfg.Path)

	return &Database{DB: db}, nil
}

func (d *Database) Migrate() error {
	slog.Info("running database migrations")

	if err := d.DB.AutoMigrate(models.AllModels()...); err != nil {
		return fmt.Errorf("auto-migrating models: %w", err)
	}

	slog.Info("database migrations completed")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("getting underlying db: %w", err)
	}
	return sqlDB.Close()
}

func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("getting underlying db: %w", err)
	}
	return sqlDB.Ping()
}
