package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"harmony/internal/services"
)

// LibraryHandler handles library management endpoints
type LibraryHandler struct {
	service *services.LibraryService
}

// NewLibraryHandler creates a new LibraryHandler
func NewLibraryHandler(service *services.LibraryService) *LibraryHandler {
	return &LibraryHandler{service: service}
}

// ScanRequest represents a scan request
type ScanRequest struct {
	Incremental bool `json:"incremental"`
}

// Scan handles POST /api/v1/library/scan
func (h *LibraryHandler) Scan(c *gin.Context) {
	var req ScanRequest
	c.ShouldBindJSON(&req) // Optional body

	// Also check query parameter for scan type
	if c.Query("type") == "incremental" {
		req.Incremental = true
	}

	// Check if scan is already in progress
	if h.service.IsScanning() {
		Conflict(c, "scan already in progress")
		return
	}

	// Start scan in background
	// Use background context since the HTTP request context will be cancelled
	// when the response is sent, but we want the scan to continue
	go func() {
		ctx := context.Background()
		if req.Incremental {
			h.service.IncrementalScan(ctx)
		} else {
			h.service.FullScan(ctx)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"message": "scan started",
		"type":    map[bool]string{true: "incremental", false: "full"}[req.Incremental],
	})
}

// ScanStatus handles GET /api/v1/library/scan/status
func (h *LibraryHandler) ScanStatus(c *gin.Context) {
	progress := h.service.GetProgress()

	Success(c, gin.H{
		"status":         progress.Status,
		"totalFiles":     progress.TotalFiles,
		"processedFiles": progress.ProcessedFiles,
		"newTracks":      progress.NewTracks,
		"updatedTracks":  progress.UpdatedTracks,
		"deletedTracks":  progress.DeletedTracks,
		"errorCount":     progress.ErrorCount,
		"currentFile":    progress.CurrentFile,
		"startedAt":      progress.StartedAt,
		"completedAt":    progress.CompletedAt,
		"duration":       progress.Duration,
	})
}

// CancelScan handles POST /api/v1/library/scan/cancel
func (h *LibraryHandler) CancelScan(c *gin.Context) {
	if err := h.service.CancelScan(); err != nil {
		if errors.Is(err, services.ErrScanNotRunning) {
			BadRequest(c, "no scan is running")
			return
		}
		InternalError(c, "failed to cancel scan")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "scan cancellation requested",
	})
}

// Stats handles GET /api/v1/library/stats
func (h *LibraryHandler) Stats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		InternalError(c, "failed to get library stats")
		return
	}

	Success(c, gin.H{
		"totalTracks":   stats.TotalTracks,
		"totalAlbums":   stats.TotalAlbums,
		"totalArtists":  stats.TotalArtists,
		"totalDuration": stats.TotalDuration,
		"totalSize":     stats.TotalSize,
		"lastScanAt":    stats.LastScanAt,
	})
}
