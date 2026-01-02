package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"harmony/internal/database"
	"harmony/internal/models"
	"harmony/internal/scanner"
)

var (
	ErrScanInProgress = errors.New("scan already in progress")
	ErrScanNotRunning = errors.New("no scan is running")
)

// ScanStatus represents the current scan status
type ScanStatus string

const (
	ScanStatusIdle       ScanStatus = "idle"
	ScanStatusScanning   ScanStatus = "scanning"
	ScanStatusProcessing ScanStatus = "processing"
	ScanStatusCompleted  ScanStatus = "completed"
	ScanStatusFailed     ScanStatus = "failed"
	ScanStatusCancelled  ScanStatus = "cancelled"
)

// ScanProgress represents scan progress information
type ScanProgress struct {
	Status         ScanStatus `json:"status"`
	TotalFiles     int        `json:"totalFiles"`
	ProcessedFiles int        `json:"processedFiles"`
	NewTracks      int        `json:"newTracks"`
	UpdatedTracks  int        `json:"updatedTracks"`
	DeletedTracks  int        `json:"deletedTracks"`
	ErrorCount     int        `json:"errorCount"`
	CurrentFile    string     `json:"currentFile,omitempty"`
	StartedAt      time.Time  `json:"startedAt,omitempty"`
	CompletedAt    time.Time  `json:"completedAt,omitempty"`
	Duration       string     `json:"duration,omitempty"`
}

// ScanEvent represents a scan event for WebSocket updates
type ScanEvent struct {
	Type     string       `json:"type"`
	Progress ScanProgress `json:"progress"`
}

// LibraryStats contains library statistics
type LibraryStats struct {
	TotalTracks   int64  `json:"totalTracks"`
	TotalAlbums   int64  `json:"totalAlbums"`
	TotalArtists  int64  `json:"totalArtists"`
	TotalDuration int64  `json:"totalDuration"`
	TotalSize     int64  `json:"totalSize"`
	LastScanAt    string `json:"lastScanAt,omitempty"`
}

// LibraryService handles library scanning and management
type LibraryService struct {
	mediaRoot        string
	cacheDir         string
	trackRepo        *database.TrackRepository
	albumRepo        *database.AlbumRepository
	artistRepo       *database.ArtistRepository
	scanner          *scanner.Scanner
	metadataExtractor *scanner.MetadataExtractor
	artworkProcessor *scanner.ArtworkProcessor

	// Scan state
	mu            sync.RWMutex
	scanning      bool
	cancelFunc    context.CancelFunc
	progress      ScanProgress
	progressChan  chan ScanProgress
	eventHandlers []func(ScanEvent)
}

// NewLibraryService creates a new LibraryService
func NewLibraryService(
	mediaRoot string,
	cacheDir string,
	trackRepo *database.TrackRepository,
	albumRepo *database.AlbumRepository,
	artistRepo *database.ArtistRepository,
) *LibraryService {
	workerCount := runtime.NumCPU()
	if workerCount > 8 {
		workerCount = 8
	}

	return &LibraryService{
		mediaRoot:         mediaRoot,
		cacheDir:          cacheDir,
		trackRepo:         trackRepo,
		albumRepo:         albumRepo,
		artistRepo:        artistRepo,
		scanner:           scanner.NewScanner(mediaRoot, workerCount),
		metadataExtractor: scanner.NewMetadataExtractor(),
		artworkProcessor:  scanner.NewArtworkProcessor(cacheDir),
		progress:          ScanProgress{Status: ScanStatusIdle},
	}
}

// OnScanEvent registers a handler for scan events
func (s *LibraryService) OnScanEvent(handler func(ScanEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventHandlers = append(s.eventHandlers, handler)
}

// emitEvent sends an event to all registered handlers
func (s *LibraryService) emitEvent(eventType string) {
	s.mu.RLock()
	handlers := s.eventHandlers
	progress := s.progress
	s.mu.RUnlock()

	event := ScanEvent{
		Type:     eventType,
		Progress: progress,
	}

	for _, handler := range handlers {
		go handler(event)
	}
}

// GetProgress returns the current scan progress
func (s *LibraryService) GetProgress() ScanProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress
}

// IsScanning returns whether a scan is in progress
func (s *LibraryService) IsScanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scanning
}

// FullScan performs a full library scan
func (s *LibraryService) FullScan(ctx context.Context) error {
	return s.scan(ctx, false)
}

// IncrementalScan performs an incremental library scan
func (s *LibraryService) IncrementalScan(ctx context.Context) error {
	return s.scan(ctx, true)
}

// scan performs the actual scan operation
func (s *LibraryService) scan(ctx context.Context, incremental bool) error {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return ErrScanInProgress
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel
	s.scanning = true
	s.progress = ScanProgress{
		Status:    ScanStatusScanning,
		StartedAt: time.Now(),
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.scanning = false
		s.cancelFunc = nil
		s.progress.CompletedAt = time.Now()
		s.progress.Duration = s.progress.CompletedAt.Sub(s.progress.StartedAt).String()
		s.mu.Unlock()
	}()

	scanType := "full"
	if incremental {
		scanType = "incremental"
	}
	slog.Info("starting library scan", "type", scanType, "mediaRoot", s.mediaRoot)
	s.emitEvent("scan_started")

	// Load known files for incremental scan
	if incremental {
		if err := s.loadKnownFiles(ctx); err != nil {
			s.setStatus(ScanStatusFailed)
			return fmt.Errorf("loading known files: %w", err)
		}
	}

	// Discover files
	var files []scanner.FileInfo
	var err error
	if incremental {
		files, err = s.scanner.DiscoverNewAndModified(ctx)
	} else {
		files, err = s.scanner.DiscoverFiles(ctx)
	}
	if err != nil {
		s.setStatus(ScanStatusFailed)
		return fmt.Errorf("discovering files: %w", err)
	}

	s.mu.Lock()
	s.progress.TotalFiles = len(files)
	s.progress.Status = ScanStatusProcessing
	s.mu.Unlock()
	s.emitEvent("scan_progress")

	// Process files concurrently
	if err := s.processFiles(ctx, files); err != nil {
		if errors.Is(err, context.Canceled) {
			s.setStatus(ScanStatusCancelled)
			return err
		}
		s.setStatus(ScanStatusFailed)
		return fmt.Errorf("processing files: %w", err)
	}

	// Cleanup deleted files (only on full scan)
	if !incremental {
		if err := s.cleanupDeletedFiles(ctx); err != nil {
			slog.Warn("cleanup failed", "error", err)
		}
	}

	s.setStatus(ScanStatusCompleted)
	slog.Info("library scan completed",
		"newTracks", s.progress.NewTracks,
		"updatedTracks", s.progress.UpdatedTracks,
		"deletedTracks", s.progress.DeletedTracks,
		"errors", s.progress.ErrorCount,
	)
	s.emitEvent("scan_completed")

	return nil
}

// processFiles processes discovered files concurrently
func (s *LibraryService) processFiles(ctx context.Context, files []scanner.FileInfo) error {
	if len(files) == 0 {
		return nil
	}

	workerCount := runtime.NumCPU()
	if workerCount > 8 {
		workerCount = 8
	}

	fileChan := make(chan scanner.FileInfo, workerCount*2)
	var wg sync.WaitGroup
	var processedCount int64
	var newCount, updatedCount, errorCount int64

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fileInfo := range fileChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				isNew, err := s.processFile(ctx, fileInfo)
				if err != nil {
					slog.Warn("failed to process file", "path", fileInfo.Path, "error", err)
					atomic.AddInt64(&errorCount, 1)
				} else if isNew {
					atomic.AddInt64(&newCount, 1)
				} else {
					atomic.AddInt64(&updatedCount, 1)
				}

				processed := atomic.AddInt64(&processedCount, 1)

				// Update progress periodically
				if processed%10 == 0 || processed == int64(len(files)) {
					s.mu.Lock()
					s.progress.ProcessedFiles = int(processed)
					s.progress.NewTracks = int(atomic.LoadInt64(&newCount))
					s.progress.UpdatedTracks = int(atomic.LoadInt64(&updatedCount))
					s.progress.ErrorCount = int(atomic.LoadInt64(&errorCount))
					s.progress.CurrentFile = fileInfo.Path
					s.mu.Unlock()
					s.emitEvent("scan_progress")
				}
			}
		}()
	}

	// Send files to workers
	for _, file := range files {
		select {
		case <-ctx.Done():
			close(fileChan)
			wg.Wait()
			return ctx.Err()
		case fileChan <- file:
		}
	}
	close(fileChan)
	wg.Wait()

	return nil
}

// processFile processes a single audio file
func (s *LibraryService) processFile(ctx context.Context, fileInfo scanner.FileInfo) (bool, error) {
	// Extract metadata
	metadata, err := s.metadataExtractor.Extract(fileInfo.Path)
	if err != nil {
		return false, fmt.Errorf("extracting metadata: %w", err)
	}

	// Find or create artist
	artist, err := s.artistRepo.FindOrCreate(ctx, metadata.Artist)
	if err != nil {
		return false, fmt.Errorf("finding/creating artist: %w", err)
	}

	// Find or create album
	album, err := s.findOrCreateAlbum(ctx, metadata, artist.ID, fileInfo.Path)
	if err != nil {
		return false, fmt.Errorf("finding/creating album: %w", err)
	}

	// Check if track exists
	existingTrack, err := s.trackRepo.FindByFilePath(ctx, fileInfo.Path)
	isNew := errors.Is(err, database.ErrTrackNotFound)

	// Create or update track
	track := &models.Track{
		Title:       metadata.Title,
		Duration:    metadata.Duration,
		TrackNumber: metadata.TrackNumber,
		DiscNumber:  metadata.DiscNumber,
		FilePath:    fileInfo.Path,
		FileSize:    fileInfo.Size,
		Format:      metadata.Format,
		Bitrate:     metadata.Bitrate,
		SampleRate:  metadata.SampleRate,
		Channels:    metadata.Channels,
		AlbumID:     album.ID,
		ArtistID:    artist.ID,
		Genre:       metadata.Genre,
		Year:        metadata.Year,
	}

	if isNew {
		track.ID = database.GenerateID()
		if err := s.trackRepo.Create(ctx, track); err != nil {
			return false, fmt.Errorf("creating track: %w", err)
		}
	} else {
		track.ID = existingTrack.ID
		track.CreatedAt = existingTrack.CreatedAt
		if err := s.trackRepo.Update(ctx, track); err != nil {
			return false, fmt.Errorf("updating track: %w", err)
		}
	}

	return isNew, nil
}

// findOrCreateAlbum finds or creates an album
func (s *LibraryService) findOrCreateAlbum(ctx context.Context, metadata *scanner.TrackMetadata, artistID string, audioPath string) (*models.Album, error) {
	// Try to find existing album
	album, err := s.albumRepo.FindByTitleAndArtist(ctx, metadata.Album, artistID)
	if err == nil {
		return album, nil
	}
	if !errors.Is(err, database.ErrAlbumNotFound) {
		return nil, err
	}

	// Create new album
	album = &models.Album{
		ID:       database.GenerateID(),
		Title:    metadata.Album,
		Year:     metadata.Year,
		ArtistID: artistID,
	}

	if err := s.albumRepo.Create(ctx, album); err != nil {
		return nil, fmt.Errorf("creating album: %w", err)
	}

	// Process artwork for new album
	go func() {
		artwork, err := s.artworkProcessor.FindArtwork(audioPath)
		if err != nil {
			slog.Debug("no artwork found", "album", album.Title, "error", err)
			return
		}
		if artwork == nil {
			return
		}

		paths, err := s.artworkProcessor.ProcessAndCache(artwork, album.ID)
		if err != nil {
			slog.Warn("failed to process artwork", "album", album.Title, "error", err)
			return
		}

		if originalPath, ok := paths["original"]; ok {
			album.CoverArtPath = originalPath
			s.albumRepo.Update(context.Background(), album)
		}
	}()

	return album, nil
}

// loadKnownFiles loads existing file paths and mod times from the database
func (s *LibraryService) loadKnownFiles(ctx context.Context) error {
	paths, err := s.trackRepo.GetAllFilePaths(ctx)
	if err != nil {
		return err
	}

	knownFiles := make(map[string]time.Time)
	for _, path := range paths {
		// We don't have mod times stored, so use zero time
		// This means all files will be considered "modified"
		knownFiles[path] = time.Time{}
	}

	s.scanner.SetKnownFiles(knownFiles)
	return nil
}

// cleanupDeletedFiles removes database entries for files that no longer exist
func (s *LibraryService) cleanupDeletedFiles(ctx context.Context) error {
	deleted, err := s.scanner.FindDeletedFiles(ctx)
	if err != nil {
		return err
	}

	var deletedCount int
	for _, path := range deleted {
		if err := s.trackRepo.DeleteByFilePath(ctx, path); err != nil {
			slog.Warn("failed to delete track", "path", path, "error", err)
			continue
		}
		deletedCount++
	}

	s.mu.Lock()
	s.progress.DeletedTracks = deletedCount
	s.mu.Unlock()

	// Clean up empty albums and artists
	if deletedCount > 0 {
		albumsDeleted, err := s.albumRepo.DeleteEmpty(ctx)
		if err != nil {
			slog.Warn("failed to clean up empty albums", "error", err)
		} else if albumsDeleted > 0 {
			slog.Info("cleaned up empty albums", "count", albumsDeleted)
		}

		artistsDeleted, err := s.artistRepo.DeleteEmpty(ctx)
		if err != nil {
			slog.Warn("failed to clean up empty artists", "error", err)
		} else if artistsDeleted > 0 {
			slog.Info("cleaned up empty artists", "count", artistsDeleted)
		}
	}

	return nil
}

// CancelScan cancels the current scan
func (s *LibraryService) CancelScan() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.scanning || s.cancelFunc == nil {
		return ErrScanNotRunning
	}

	s.cancelFunc()
	return nil
}

// setStatus updates the scan status
func (s *LibraryService) setStatus(status ScanStatus) {
	s.mu.Lock()
	s.progress.Status = status
	s.mu.Unlock()
}

// GetStats returns library statistics
func (s *LibraryService) GetStats(ctx context.Context) (*LibraryStats, error) {
	trackCount, err := s.trackRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting tracks: %w", err)
	}

	albumCount, err := s.albumRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting albums: %w", err)
	}

	artistCount, err := s.artistRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("counting artists: %w", err)
	}

	return &LibraryStats{
		TotalTracks:  trackCount,
		TotalAlbums:  albumCount,
		TotalArtists: artistCount,
	}, nil
}
