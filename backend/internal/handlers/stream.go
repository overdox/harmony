package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
	"harmony/internal/transcoder"
)

// MIME types for audio formats
var audioMIMETypes = map[string]string{
	"mp3":  "audio/mpeg",
	"flac": "audio/flac",
	"wav":  "audio/wav",
	"ogg":  "audio/ogg",
	"m4a":  "audio/mp4",
	"aac":  "audio/aac",
	"opus": "audio/opus",
	"wma":  "audio/x-ms-wma",
}

// StreamHandler handles audio streaming requests
type StreamHandler struct {
	trackRepo   *database.TrackRepository
	transcoder  *transcoder.Transcoder
	mediaRoot   string
}

// NewStreamHandler creates a new StreamHandler
func NewStreamHandler(
	trackRepo *database.TrackRepository,
	transcoder *transcoder.Transcoder,
	mediaRoot string,
) *StreamHandler {
	return &StreamHandler{
		trackRepo:  trackRepo,
		transcoder: transcoder,
		mediaRoot:  mediaRoot,
	}
}

// Stream handles streaming requests for a track
func (h *StreamHandler) Stream(c *gin.Context) {
	trackID := c.Param("id")
	if trackID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "track ID required"})
		return
	}

	// Get track from database
	track, err := h.trackRepo.FindByID(c.Request.Context(), trackID)
	if err != nil {
		if errors.Is(err, database.ErrTrackNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get track"})
		return
	}

	// Validate file path is within media root (security)
	absPath, err := filepath.Abs(track.FilePath)
	if err != nil || !strings.HasPrefix(absPath, h.mediaRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Check if file exists
	fileInfo, err := os.Stat(track.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to access file"})
		return
	}

	// Get quality parameter
	quality := c.Query("quality")
	if quality == "" {
		quality = h.detectQuality(c)
	}

	// Handle transcoding if requested
	if quality != "" && quality != "original" {
		h.streamTranscoded(c, track.FilePath, track.Format, quality)
		return
	}

	// Stream original file
	h.streamOriginal(c, track.FilePath, track.Format, fileInfo)
}

// streamOriginal streams the original file with range request support
func (h *StreamHandler) streamOriginal(c *gin.Context, filePath, format string, fileInfo os.FileInfo) {
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer file.Close()

	// Get MIME type
	mimeType := getMIMEType(format)

	// Set headers
	c.Header("Content-Type", mimeType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

	// Handle conditional requests
	if h.handleConditional(c, fileInfo) {
		return
	}

	// Handle range requests
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		h.serveRange(c, file, fileInfo, rangeHeader)
		return
	}

	// Serve entire file
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, file)
}

// streamTranscoded streams a transcoded version of the file
func (h *StreamHandler) streamTranscoded(c *gin.Context, filePath, format, quality string) {
	if h.transcoder == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "transcoding not available"})
		return
	}

	profile, err := transcoder.GetProfile(quality)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quality"})
		return
	}

	// Check if cached version exists
	cachedPath := h.transcoder.GetCachedPath(filePath, profile)
	if cachedPath != "" {
		if fileInfo, err := os.Stat(cachedPath); err == nil {
			h.streamOriginal(c, cachedPath, profile.Format, fileInfo)
			return
		}
	}

	// Set headers for streaming transcoded content
	c.Header("Content-Type", getMIMEType(profile.Format))
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")
	c.Status(http.StatusOK)

	// Stream transcoded content
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Minute)
	defer cancel()

	err = h.transcoder.TranscodeToWriter(ctx, filePath, profile, c.Writer)
	if err != nil {
		// Can't send error response after streaming started
		return
	}
}

// serveRange handles HTTP range requests for seeking
func (h *StreamHandler) serveRange(c *gin.Context, file *os.File, fileInfo os.FileInfo, rangeHeader string) {
	fileSize := fileInfo.Size()

	// Parse range header
	start, end, err := parseRangeHeader(rangeHeader, fileSize)
	if err != nil {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Seek to start position
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "seek failed"})
		return
	}

	// Calculate content length
	contentLength := end - start + 1

	// Set headers
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Status(http.StatusPartialContent)

	// Copy the requested range
	io.CopyN(c.Writer, file, contentLength)
}

// handleConditional handles If-Modified-Since and If-Range headers
func (h *StreamHandler) handleConditional(c *gin.Context, fileInfo os.FileInfo) bool {
	modTime := fileInfo.ModTime()

	// Check If-Modified-Since
	ifModSince := c.GetHeader("If-Modified-Since")
	if ifModSince != "" {
		t, err := http.ParseTime(ifModSince)
		if err == nil && !modTime.After(t) {
			c.Status(http.StatusNotModified)
			return true
		}
	}

	// Check If-Range (for range requests)
	ifRange := c.GetHeader("If-Range")
	if ifRange != "" {
		t, err := http.ParseTime(ifRange)
		if err == nil && modTime.After(t) {
			// Resource has been modified, ignore range request
			c.Request.Header.Del("Range")
		}
	}

	return false
}

// detectQuality auto-detects quality based on client hints
func (h *StreamHandler) detectQuality(c *gin.Context) string {
	// Check Save-Data header
	if c.GetHeader("Save-Data") == "on" {
		return "low"
	}

	// Check network quality hints
	ect := c.GetHeader("ECT") // Effective Connection Type
	switch ect {
	case "slow-2g", "2g":
		return "low"
	case "3g":
		return "medium"
	case "4g":
		return "high"
	}

	// Default to original quality
	return "original"
}

// parseRangeHeader parses the Range header and returns start and end positions
func parseRangeHeader(rangeHeader string, fileSize int64) (int64, int64, error) {
	// Format: "bytes=start-end" or "bytes=start-" or "bytes=-suffix"
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeSpec, "-")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	var start, end int64
	var err error

	if parts[0] == "" {
		// Suffix range: "-500" means last 500 bytes
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range suffix")
		}
		start = fileSize - suffix
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range start")
		}

		if parts[1] == "" {
			// Open-ended range: "500-" means from 500 to end
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("invalid range end")
			}
		}
	}

	// Validate range
	if start < 0 || start >= fileSize || end < start || end >= fileSize {
		return 0, 0, fmt.Errorf("range out of bounds")
	}

	return start, end, nil
}

// getMIMEType returns the MIME type for an audio format
func getMIMEType(format string) string {
	format = strings.ToLower(format)
	if mime, ok := audioMIMETypes[format]; ok {
		return mime
	}
	return "application/octet-stream"
}

// GetStreamURL generates a stream URL for a track
func GetStreamURL(baseURL, trackID, quality string) string {
	url := fmt.Sprintf("%s/api/v1/tracks/%s/stream", baseURL, trackID)
	if quality != "" && quality != "original" {
		url += "?quality=" + quality
	}
	return url
}
