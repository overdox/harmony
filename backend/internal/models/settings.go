package models

import (
	"time"
)

// Settings stores application configuration
type Settings struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Settings) TableName() string {
	return "settings"
}

// Common settings keys
const (
	SettingSetupCompleted = "setup_completed"
	SettingMediaPaths     = "media_paths"
	SettingAppName        = "app_name"
	SettingTheme          = "theme"
)
