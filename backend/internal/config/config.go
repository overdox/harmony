package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values for the application
type Config struct {
	// Server settings
	Port     int
	LogLevel string

	// Database settings
	DBPath   string
	RedisURL string

	// Media settings
	MediaPath   string
	ArtworkPath string
	CachePath   string

	// Feature flags
	ScanOnStartup bool
}

// Default values
const (
	DefaultPort        = 8080
	DefaultLogLevel    = "info"
	DefaultDBPath      = "/data/harmony.db"
	DefaultRedisURL    = "redis://localhost:6379"
	DefaultMediaPath   = "/media"
	DefaultArtworkPath = "/app/artwork"
	DefaultCachePath   = "/app/cache"
)

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnvInt("PORT", DefaultPort),
		LogLevel:      getEnv("LOG_LEVEL", DefaultLogLevel),
		DBPath:        getEnv("DB_PATH", DefaultDBPath),
		RedisURL:      getEnv("REDIS_URL", DefaultRedisURL),
		MediaPath:     getEnv("MEDIA_PATH", DefaultMediaPath),
		ArtworkPath:   getEnv("ARTWORK_PATH", DefaultArtworkPath),
		CachePath:     getEnv("CACHE_PATH", DefaultCachePath),
		ScanOnStartup: getEnvBool("SCAN_ON_STARTUP", false),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks all configuration values
func (c *Config) Validate() error {
	var errs []string

	// Validate port
	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Sprintf("invalid port: %d (must be 1-65535)", c.Port))
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[strings.ToLower(c.LogLevel)] {
		errs = append(errs, fmt.Sprintf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel))
	}

	// Validate required paths
	if c.DBPath == "" {
		errs = append(errs, "DB_PATH is required")
	}

	if c.MediaPath == "" {
		errs = append(errs, "MEDIA_PATH is required")
	}

	// Check if media path exists (warning only, might be mounted later in Docker)
	if c.MediaPath != "" {
		if info, err := os.Stat(c.MediaPath); err != nil {
			slog.Warn("media path does not exist", "path", c.MediaPath, "error", err)
		} else if !info.IsDir() {
			errs = append(errs, fmt.Sprintf("MEDIA_PATH is not a directory: %s", c.MediaPath))
		}
	}

	// Validate Redis URL format
	if c.RedisURL != "" && !strings.HasPrefix(c.RedisURL, "redis://") && !strings.HasPrefix(c.RedisURL, "rediss://") {
		errs = append(errs, fmt.Sprintf("invalid REDIS_URL format: %s (must start with redis:// or rediss://)", c.RedisURL))
	}

	if len(errs) > 0 {
		return errors.New("configuration validation failed:\n  - " + strings.Join(errs, "\n  - "))
	}

	return nil
}

// LogLevel returns the slog.Level for the configured log level
func (c *Config) SlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Print logs the current configuration (with sensitive values masked)
func (c *Config) Print() {
	slog.Info("configuration loaded",
		"port", c.Port,
		"log_level", c.LogLevel,
		"db_path", c.DBPath,
		"redis_url", maskRedisURL(c.RedisURL),
		"media_path", c.MediaPath,
		"artwork_path", c.ArtworkPath,
		"cache_path", c.CachePath,
		"scan_on_startup", c.ScanOnStartup,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func maskRedisURL(url string) string {
	// Mask password in Redis URL if present
	// Format: redis://[:password]@host:port/db
	if strings.Contains(url, "@") {
		parts := strings.SplitN(url, "@", 2)
		return "redis://***@" + parts[1]
	}
	return url
}
