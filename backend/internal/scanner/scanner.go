package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Supported audio formats
var SupportedFormats = map[string]bool{
	".mp3":  true,
	".flac": true,
	".wav":  true,
	".ogg":  true,
	".m4a":  true,
	".aac":  true,
	".opus": true,
	".wma":  true,
}

// FileInfo contains information about a discovered audio file
type FileInfo struct {
	Path         string
	Size         int64
	ModTime      time.Time
	Format       string
	Hash         string
	IsNew        bool
	IsModified   bool
}

// ScanResult contains the result of scanning a single file
type ScanResult struct {
	FileInfo FileInfo
	Metadata *TrackMetadata
	Artwork  *ArtworkInfo
	Error    error
}

// ScanProgress reports scan progress
type ScanProgress struct {
	TotalFiles     int
	ProcessedFiles int
	CurrentFile    string
	NewFiles       int
	ModifiedFiles  int
	ErrorCount     int
}

// Scanner handles file discovery in media directories
type Scanner struct {
	mediaRoot     string
	knownFiles    map[string]time.Time // path -> modTime
	mu            sync.RWMutex
	progressChan  chan ScanProgress
	workerCount   int
}

// NewScanner creates a new Scanner instance
func NewScanner(mediaRoot string, workerCount int) *Scanner {
	if workerCount <= 0 {
		workerCount = 4
	}
	return &Scanner{
		mediaRoot:   mediaRoot,
		knownFiles:  make(map[string]time.Time),
		workerCount: workerCount,
	}
}

// SetKnownFiles sets the map of known files and their modification times
func (s *Scanner) SetKnownFiles(files map[string]time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.knownFiles = files
}

// SetProgressChannel sets the channel for progress updates
func (s *Scanner) SetProgressChannel(ch chan ScanProgress) {
	s.progressChan = ch
}

// DiscoverFiles walks the media directory and returns all audio files
func (s *Scanner) DiscoverFiles(ctx context.Context) ([]FileInfo, error) {
	var files []FileInfo
	var mu sync.Mutex

	slog.Info("starting file discovery", "root", s.mediaRoot)

	err := filepath.WalkDir(s.mediaRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("error accessing path", "path", path, "error", err)
			return nil // Continue walking
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip directories
		if d.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(d.Name(), ".") && path != s.mediaRoot {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is a supported audio format
		ext := strings.ToLower(filepath.Ext(path))
		if !SupportedFormats[ext] {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			slog.Warn("error getting file info", "path", path, "error", err)
			return nil
		}

		fileInfo := FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Format:  ext[1:], // Remove leading dot
		}

		// Check if file is new or modified
		s.mu.RLock()
		knownModTime, exists := s.knownFiles[path]
		s.mu.RUnlock()

		if !exists {
			fileInfo.IsNew = true
		} else if info.ModTime().After(knownModTime) {
			fileInfo.IsModified = true
		}

		mu.Lock()
		files = append(files, fileInfo)
		mu.Unlock()

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	slog.Info("file discovery complete", "totalFiles", len(files))
	return files, nil
}

// DiscoverNewAndModified returns only new or modified files
func (s *Scanner) DiscoverNewAndModified(ctx context.Context) ([]FileInfo, error) {
	allFiles, err := s.DiscoverFiles(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []FileInfo
	for _, f := range allFiles {
		if f.IsNew || f.IsModified {
			filtered = append(filtered, f)
		}
	}

	slog.Info("incremental scan", "newFiles", len(filtered), "totalFiles", len(allFiles))
	return filtered, nil
}

// ComputeFileHash generates a SHA256 hash of the file content
func (s *Scanner) ComputeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// For large files, only hash the first and last 1MB
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("getting file stat: %w", err)
	}

	hasher := sha256.New()

	if info.Size() > 2*1024*1024 {
		// Hash first 1MB
		buf := make([]byte, 1024*1024)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("reading file start: %w", err)
		}
		hasher.Write(buf[:n])

		// Hash last 1MB
		_, err = file.Seek(-1024*1024, io.SeekEnd)
		if err != nil {
			return "", fmt.Errorf("seeking file end: %w", err)
		}
		n, err = file.Read(buf)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("reading file end: %w", err)
		}
		hasher.Write(buf[:n])

		// Include file size in hash
		hasher.Write([]byte(fmt.Sprintf("%d", info.Size())))
	} else {
		// Hash entire file for small files
		if _, err := io.Copy(hasher, file); err != nil {
			return "", fmt.Errorf("hashing file: %w", err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// FindDeletedFiles returns paths of files that exist in knownFiles but not on disk
func (s *Scanner) FindDeletedFiles(ctx context.Context) ([]string, error) {
	var deleted []string

	s.mu.RLock()
	defer s.mu.RUnlock()

	for path := range s.knownFiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			deleted = append(deleted, path)
		}
	}

	slog.Info("found deleted files", "count", len(deleted))
	return deleted, nil
}

// IsSupportedFormat checks if a file extension is a supported audio format
func IsSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return SupportedFormats[ext]
}

// GetFormatFromPath extracts the format from a file path
func GetFormatFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if len(ext) > 0 {
		return ext[1:] // Remove leading dot
	}
	return ""
}
