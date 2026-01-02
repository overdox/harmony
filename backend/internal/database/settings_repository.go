package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"harmony/internal/models"
)

var (
	ErrSettingNotFound = errors.New("setting not found")
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves a setting by key
func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var setting models.Settings
	result := r.db.WithContext(ctx).Where("key = ?", key).First(&setting)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", ErrSettingNotFound
		}
		return "", fmt.Errorf("getting setting: %w", result.Error)
	}
	return setting.Value, nil
}

// Set creates or updates a setting
func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	setting := models.Settings{
		Key:   key,
		Value: value,
	}

	result := r.db.WithContext(ctx).
		Where("key = ?", key).
		Assign(models.Settings{Value: value}).
		FirstOrCreate(&setting)

	if result.Error != nil {
		return fmt.Errorf("setting value: %w", result.Error)
	}
	return nil
}

// GetBool retrieves a boolean setting
func (r *SettingsRepository) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := r.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return value == "true" || value == "1", nil
}

// SetBool sets a boolean setting
func (r *SettingsRepository) SetBool(ctx context.Context, key string, value bool) error {
	strValue := "false"
	if value {
		strValue = "true"
	}
	return r.Set(ctx, key, strValue)
}

// GetJSON retrieves and unmarshals a JSON setting
func (r *SettingsRepository) GetJSON(ctx context.Context, key string, dest interface{}) error {
	value, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// SetJSON marshals and stores a JSON setting
func (r *SettingsRepository) SetJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	return r.Set(ctx, key, string(data))
}

// Delete removes a setting
func (r *SettingsRepository) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Settings{})
	if result.Error != nil {
		return fmt.Errorf("deleting setting: %w", result.Error)
	}
	return nil
}

// GetAll retrieves all settings
func (r *SettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	var settings []models.Settings
	if err := r.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("getting all settings: %w", err)
	}

	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// IsSetupCompleted checks if initial setup has been completed
func (r *SettingsRepository) IsSetupCompleted(ctx context.Context) bool {
	completed, err := r.GetBool(ctx, models.SettingSetupCompleted)
	if err != nil {
		return false
	}
	return completed
}

// MarkSetupCompleted marks the setup as completed
func (r *SettingsRepository) MarkSetupCompleted(ctx context.Context) error {
	return r.SetBool(ctx, models.SettingSetupCompleted, true)
}

// GetMediaPaths retrieves the configured media paths
func (r *SettingsRepository) GetMediaPaths(ctx context.Context) ([]string, error) {
	var paths []string
	err := r.GetJSON(ctx, models.SettingMediaPaths, &paths)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return []string{}, nil
		}
		return nil, err
	}
	return paths, nil
}

// SetMediaPaths saves the configured media paths
func (r *SettingsRepository) SetMediaPaths(ctx context.Context, paths []string) error {
	return r.SetJSON(ctx, models.SettingMediaPaths, paths)
}
