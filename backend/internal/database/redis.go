package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		URL: "redis://localhost:6379",
		DB:  0,
	}
}

func NewRedis(cfg RedisConfig) (*RedisClient, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	if cfg.Password != "" {
		opt.Password = cfg.Password
	}
	opt.DB = cfg.DB

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	slog.Info("redis connection established", "url", cfg.URL)

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Cache key prefixes
const (
	KeyPrefixTrack       = "track:"
	KeyPrefixAlbum       = "album:"
	KeyPrefixArtist      = "artist:"
	KeyPrefixAlbumArt    = "art:"
	KeyPrefixSearch      = "search:"
	KeyPrefixLibraryStats = "library:stats"
)

// TTL durations
const (
	TTLTrackMetadata = 30 * time.Minute
	TTLAlbumArt      = 1 * time.Hour
	TTLSearchResults = 5 * time.Minute
	TTLLibraryStats  = 5 * time.Minute
)

// Get retrieves a value from cache
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stores a value in cache with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from cache
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// GetJSON retrieves and unmarshals a JSON value
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals and stores a JSON value
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// CacheTrack caches track metadata
func (r *RedisClient) CacheTrack(ctx context.Context, trackID string, track interface{}) error {
	key := KeyPrefixTrack + trackID
	return r.SetJSON(ctx, key, track, TTLTrackMetadata)
}

// GetCachedTrack retrieves cached track metadata
func (r *RedisClient) GetCachedTrack(ctx context.Context, trackID string, dest interface{}) error {
	key := KeyPrefixTrack + trackID
	return r.GetJSON(ctx, key, dest)
}

// CacheAlbumArt caches album artwork path
func (r *RedisClient) CacheAlbumArt(ctx context.Context, albumID string, artPath string) error {
	key := KeyPrefixAlbumArt + albumID
	return r.Set(ctx, key, artPath, TTLAlbumArt)
}

// GetCachedAlbumArt retrieves cached album artwork path
func (r *RedisClient) GetCachedAlbumArt(ctx context.Context, albumID string) (string, error) {
	key := KeyPrefixAlbumArt + albumID
	return r.Get(ctx, key)
}

// CacheSearchResults caches search results
func (r *RedisClient) CacheSearchResults(ctx context.Context, query string, results interface{}) error {
	key := KeyPrefixSearch + query
	return r.SetJSON(ctx, key, results, TTLSearchResults)
}

// GetCachedSearchResults retrieves cached search results
func (r *RedisClient) GetCachedSearchResults(ctx context.Context, query string, dest interface{}) error {
	key := KeyPrefixSearch + query
	return r.GetJSON(ctx, key, dest)
}

// InvalidateTrack removes a track from cache
func (r *RedisClient) InvalidateTrack(ctx context.Context, trackID string) error {
	return r.Delete(ctx, KeyPrefixTrack+trackID)
}

// InvalidateSearchCache clears all search cache
func (r *RedisClient) InvalidateSearchCache(ctx context.Context) error {
	iter := r.client.Scan(ctx, 0, KeyPrefixSearch+"*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.Delete(ctx, keys...)
	}
	return nil
}
