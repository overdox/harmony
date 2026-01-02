package transcoder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrFFmpegNotFound   = errors.New("ffmpeg not found")
	ErrInvalidProfile   = errors.New("invalid transcoding profile")
	ErrTranscodeFailed  = errors.New("transcoding failed")
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// Profile represents a transcoding profile
type Profile struct {
	Name    string
	Format  string
	Codec   string
	Bitrate int    // kbps
	Ext     string // file extension
}

// Predefined transcoding profiles
var (
	ProfileOriginal = Profile{Name: "original", Format: "", Codec: "", Bitrate: 0, Ext: ""}
	ProfileHigh     = Profile{Name: "high", Format: "mp3", Codec: "libmp3lame", Bitrate: 320, Ext: "mp3"}
	ProfileMedium   = Profile{Name: "medium", Format: "mp3", Codec: "libmp3lame", Bitrate: 192, Ext: "mp3"}
	ProfileLow      = Profile{Name: "low", Format: "mp3", Codec: "libmp3lame", Bitrate: 128, Ext: "mp3"}

	// OGG alternatives
	ProfileHighOGG   = Profile{Name: "high-ogg", Format: "ogg", Codec: "libvorbis", Bitrate: 320, Ext: "ogg"}
	ProfileMediumOGG = Profile{Name: "medium-ogg", Format: "ogg", Codec: "libvorbis", Bitrate: 192, Ext: "ogg"}
	ProfileLowOGG    = Profile{Name: "low-ogg", Format: "ogg", Codec: "libvorbis", Bitrate: 128, Ext: "ogg"}

	// All profiles map
	profiles = map[string]Profile{
		"original":   ProfileOriginal,
		"high":       ProfileHigh,
		"medium":     ProfileMedium,
		"low":        ProfileLow,
		"high-ogg":   ProfileHighOGG,
		"medium-ogg": ProfileMediumOGG,
		"low-ogg":    ProfileLowOGG,
	}
)

// GetProfile returns a profile by name
func GetProfile(name string) (Profile, error) {
	name = strings.ToLower(name)
	if profile, ok := profiles[name]; ok {
		return profile, nil
	}
	return Profile{}, ErrInvalidProfile
}

// GetAllProfiles returns all available profiles
func GetAllProfiles() []Profile {
	return []Profile{
		ProfileOriginal,
		ProfileHigh,
		ProfileMedium,
		ProfileLow,
		ProfileHighOGG,
		ProfileMediumOGG,
		ProfileLowOGG,
	}
}

// Transcoder handles audio transcoding using ffmpeg
type Transcoder struct {
	ffmpegPath string
	cacheDir   string
	maxCacheGB float64
	mu         sync.RWMutex
	cacheSize  int64
}

// Config holds transcoder configuration
type Config struct {
	FFmpegPath string
	CacheDir   string
	MaxCacheGB float64
}

// DefaultConfig returns default transcoder configuration
func DefaultConfig() Config {
	return Config{
		FFmpegPath: "ffmpeg",
		CacheDir:   "./data/transcode_cache",
		MaxCacheGB: 10.0,
	}
}

// New creates a new Transcoder
func New(cfg Config) (*Transcoder, error) {
	// Find ffmpeg
	ffmpegPath := cfg.FFmpegPath
	if ffmpegPath == "" || ffmpegPath == "ffmpeg" {
		path, err := exec.LookPath("ffmpeg")
		if err != nil {
			return nil, ErrFFmpegNotFound
		}
		ffmpegPath = path
	}

	// Verify ffmpeg works
	cmd := exec.Command(ffmpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg check failed: %w", err)
	}

	// Create cache directory
	if err := os.MkdirAll(cfg.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	t := &Transcoder{
		ffmpegPath: ffmpegPath,
		cacheDir:   cfg.CacheDir,
		maxCacheGB: cfg.MaxCacheGB,
	}

	// Calculate initial cache size
	go t.calculateCacheSize()

	slog.Info("transcoder initialized", "ffmpeg", ffmpegPath, "cacheDir", cfg.CacheDir)
	return t, nil
}

// TranscodeToFile transcodes an audio file to a new file
func (t *Transcoder) TranscodeToFile(ctx context.Context, inputPath string, profile Profile, outputPath string) error {
	args := t.buildFFmpegArgs(inputPath, profile, outputPath)

	cmd := exec.CommandContext(ctx, t.ffmpegPath, args...)
	cmd.Stderr = io.Discard // Suppress ffmpeg output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %v", ErrTranscodeFailed, err)
	}

	return nil
}

// TranscodeToWriter transcodes an audio file and writes to a writer (for streaming)
func (t *Transcoder) TranscodeToWriter(ctx context.Context, inputPath string, profile Profile, w io.Writer) error {
	args := t.buildFFmpegArgs(inputPath, profile, "pipe:1")

	cmd := exec.CommandContext(ctx, t.ffmpegPath, args...)
	cmd.Stdout = w
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	// Wait for completion or context cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("%w: %v", ErrTranscodeFailed, err)
		}
		return nil
	}
}

// TranscodeAndCache transcodes and caches the result
func (t *Transcoder) TranscodeAndCache(ctx context.Context, inputPath string, profile Profile) (string, error) {
	cacheKey := t.getCacheKey(inputPath, profile)
	cachedPath := filepath.Join(t.cacheDir, cacheKey+"."+profile.Ext)

	// Check if already cached
	if _, err := os.Stat(cachedPath); err == nil {
		return cachedPath, nil
	}

	// Create temp file for transcoding
	tempPath := cachedPath + ".tmp"
	defer os.Remove(tempPath)

	// Transcode to temp file
	if err := t.TranscodeToFile(ctx, inputPath, profile, tempPath); err != nil {
		return "", err
	}

	// Rename to final path
	if err := os.Rename(tempPath, cachedPath); err != nil {
		return "", fmt.Errorf("moving transcoded file: %w", err)
	}

	// Update cache size
	go t.updateCacheSize(cachedPath)

	return cachedPath, nil
}

// GetCachedPath returns the cached file path if it exists
func (t *Transcoder) GetCachedPath(inputPath string, profile Profile) string {
	if profile.Name == "original" {
		return inputPath
	}

	cacheKey := t.getCacheKey(inputPath, profile)
	cachedPath := filepath.Join(t.cacheDir, cacheKey+"."+profile.Ext)

	if _, err := os.Stat(cachedPath); err == nil {
		return cachedPath
	}
	return ""
}

// buildFFmpegArgs builds ffmpeg command arguments
func (t *Transcoder) buildFFmpegArgs(inputPath string, profile Profile, outputPath string) []string {
	args := []string{
		"-i", inputPath,
		"-y", // Overwrite output
		"-vn", // No video
	}

	if profile.Codec != "" {
		args = append(args, "-acodec", profile.Codec)
	}

	if profile.Bitrate > 0 {
		args = append(args, "-b:a", fmt.Sprintf("%dk", profile.Bitrate))
	}

	if profile.Format != "" {
		args = append(args, "-f", profile.Format)
	}

	// Add quality settings
	switch profile.Codec {
	case "libmp3lame":
		args = append(args, "-q:a", "2") // VBR quality
	case "libvorbis":
		args = append(args, "-q:a", "6") // VBR quality
	}

	args = append(args, outputPath)
	return args
}

// getCacheKey generates a unique cache key for a file and profile
func (t *Transcoder) getCacheKey(inputPath string, profile Profile) string {
	// Include file path, profile name, and file modification time
	info, _ := os.Stat(inputPath)
	modTime := ""
	if info != nil {
		modTime = info.ModTime().Format(time.RFC3339)
	}

	data := fmt.Sprintf("%s|%s|%s", inputPath, profile.Name, modTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// calculateCacheSize calculates the total cache size
func (t *Transcoder) calculateCacheSize() {
	var size int64
	filepath.Walk(t.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		size += info.Size()
		return nil
	})

	t.mu.Lock()
	t.cacheSize = size
	t.mu.Unlock()

	slog.Debug("cache size calculated", "size", size, "sizeGB", float64(size)/(1024*1024*1024))
}

// updateCacheSize updates cache size after adding a file
func (t *Transcoder) updateCacheSize(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	t.mu.Lock()
	t.cacheSize += info.Size()
	currentSize := t.cacheSize
	t.mu.Unlock()

	// Check if we need to clean up
	maxSize := int64(t.maxCacheGB * 1024 * 1024 * 1024)
	if currentSize > maxSize {
		go t.cleanupCache(maxSize * 80 / 100) // Clean to 80% of max
	}
}

// cleanupCache removes old cached files to stay under the size limit
func (t *Transcoder) cleanupCache(targetSize int64) {
	type fileEntry struct {
		path    string
		size    int64
		modTime time.Time
	}

	var files []fileEntry
	filepath.Walk(t.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		files = append(files, fileEntry{
			path:    path,
			size:    info.Size(),
			modTime: info.ModTime(),
		})
		return nil
	})

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].modTime.Before(files[i].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files until we're under target size
	t.mu.Lock()
	currentSize := t.cacheSize
	t.mu.Unlock()

	var removed int
	for _, f := range files {
		if currentSize <= targetSize {
			break
		}
		if err := os.Remove(f.path); err == nil {
			currentSize -= f.size
			removed++
		}
	}

	t.mu.Lock()
	t.cacheSize = currentSize
	t.mu.Unlock()

	if removed > 0 {
		slog.Info("cache cleanup completed", "filesRemoved", removed, "newSizeGB", float64(currentSize)/(1024*1024*1024))
	}
}

// ClearCache removes all cached files
func (t *Transcoder) ClearCache() error {
	err := os.RemoveAll(t.cacheDir)
	if err != nil {
		return fmt.Errorf("clearing cache: %w", err)
	}

	// Recreate directory
	if err := os.MkdirAll(t.cacheDir, 0755); err != nil {
		return fmt.Errorf("recreating cache directory: %w", err)
	}

	t.mu.Lock()
	t.cacheSize = 0
	t.mu.Unlock()

	return nil
}

// GetCacheStats returns cache statistics
func (t *Transcoder) GetCacheStats() (int64, int, error) {
	t.mu.RLock()
	size := t.cacheSize
	t.mu.RUnlock()

	var count int
	filepath.Walk(t.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		count++
		return nil
	})

	return size, count, nil
}

// IsAvailable checks if the transcoder is available
func (t *Transcoder) IsAvailable() bool {
	return t != nil && t.ffmpegPath != ""
}

// GetFFmpegPath returns the path to ffmpeg
func (t *Transcoder) GetFFmpegPath() string {
	if t == nil {
		return ""
	}
	return t.ffmpegPath
}

// ProbeAudio gets audio information using ffprobe
func (t *Transcoder) ProbeAudio(ctx context.Context, inputPath string) (*AudioInfo, error) {
	ffprobePath := strings.Replace(t.ffmpegPath, "ffmpeg", "ffprobe", 1)

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse JSON output (simplified - in production use encoding/json)
	_ = output
	return &AudioInfo{}, nil
}

// AudioInfo contains audio file information
type AudioInfo struct {
	Duration   float64
	Bitrate    int
	SampleRate int
	Channels   int
	Codec      string
	Format     string
}
