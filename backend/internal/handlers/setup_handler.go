package handlers

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
	"harmony/internal/services"
)

// SetupHandler handles setup/onboarding endpoints
type SetupHandler struct {
	settingsRepo   *database.SettingsRepository
	libraryService *services.LibraryService
	mediaRoot      string
}

// NewSetupHandler creates a new SetupHandler
func NewSetupHandler(settingsRepo *database.SettingsRepository, libraryService *services.LibraryService, mediaRoot string) *SetupHandler {
	return &SetupHandler{
		settingsRepo:   settingsRepo,
		libraryService: libraryService,
		mediaRoot:      mediaRoot,
	}
}

// FolderInfo represents a folder in the browser
type FolderInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	HasAudio bool   `json:"hasAudio"`
	Children int    `json:"children"`
}

// Status handles GET /api/v1/setup/status
func (h *SetupHandler) Status(c *gin.Context) {
	ctx := c.Request.Context()
	completed := h.settingsRepo.IsSetupCompleted(ctx)

	Success(c, gin.H{
		"completed": completed,
		"mediaRoot": h.mediaRoot,
	})
}

// BrowseFolders handles GET /api/v1/setup/folders
func (h *SetupHandler) BrowseFolders(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = h.mediaRoot
	}

	// Security: ensure path is within media root
	absPath, err := filepath.Abs(path)
	if err != nil {
		BadRequest(c, "invalid path")
		return
	}

	absMediaRoot, _ := filepath.Abs(h.mediaRoot)
	if !strings.HasPrefix(absPath, absMediaRoot) {
		BadRequest(c, "path outside media root")
		return
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			NotFound(c, "path not found")
			return
		}
		InternalError(c, "failed to access path")
		return
	}

	if !info.IsDir() {
		BadRequest(c, "path is not a directory")
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(absPath)
	if err != nil {
		InternalError(c, "failed to read directory")
		return
	}

	var folders []FolderInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		folderPath := filepath.Join(absPath, entry.Name())
		hasAudio := h.hasAudioFiles(folderPath)
		children := h.countSubdirectories(folderPath)

		folders = append(folders, FolderInfo{
			Name:     entry.Name(),
			Path:     folderPath,
			HasAudio: hasAudio,
			Children: children,
		})
	}

	// Also check if current directory has audio files
	currentHasAudio := h.hasAudioFiles(absPath)

	Success(c, gin.H{
		"currentPath":     absPath,
		"parentPath":      filepath.Dir(absPath),
		"isMediaRoot":     absPath == absMediaRoot,
		"hasAudioFiles":   currentHasAudio,
		"folders":         folders,
	})
}

// GetSelectedFolders handles GET /api/v1/setup/selected-folders
func (h *SetupHandler) GetSelectedFolders(c *gin.Context) {
	ctx := c.Request.Context()
	paths, err := h.settingsRepo.GetMediaPaths(ctx)
	if err != nil {
		InternalError(c, "failed to get selected folders")
		return
	}

	Success(c, gin.H{
		"paths": paths,
	})
}

// SetSelectedFolders handles POST /api/v1/setup/selected-folders
func (h *SetupHandler) SetSelectedFolders(c *gin.Context) {
	var req struct {
		Paths []string `json:"paths" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	// Validate all paths are within media root
	absMediaRoot, _ := filepath.Abs(h.mediaRoot)
	for _, path := range req.Paths {
		absPath, err := filepath.Abs(path)
		if err != nil || !strings.HasPrefix(absPath, absMediaRoot) {
			BadRequest(c, "invalid path: "+path)
			return
		}
	}

	ctx := c.Request.Context()
	if err := h.settingsRepo.SetMediaPaths(ctx, req.Paths); err != nil {
		InternalError(c, "failed to save selected folders")
		return
	}

	Success(c, gin.H{
		"message": "folders saved",
		"paths":   req.Paths,
	})
}

// Complete handles POST /api/v1/setup/complete
func (h *SetupHandler) Complete(c *gin.Context) {
	var req struct {
		StartScan bool `json:"startScan"`
	}
	c.ShouldBindJSON(&req)

	ctx := c.Request.Context()

	// Mark setup as completed
	if err := h.settingsRepo.MarkSetupCompleted(ctx); err != nil {
		InternalError(c, "failed to complete setup")
		return
	}

	// Optionally start a scan
	if req.StartScan {
		go h.libraryService.FullScan(ctx)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "setup completed",
		"scanStarted": req.StartScan,
	})
}

// hasAudioFiles checks if a directory contains audio files (non-recursive)
func (h *SetupHandler) hasAudioFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	audioExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".aac":  true,
		".opus": true,
		".wma":  true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if audioExtensions[ext] {
			return true
		}
	}

	return false
}

// countSubdirectories counts immediate subdirectories
func (h *SetupHandler) countSubdirectories(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			count++
		}
	}
	return count
}

// hasAudioFilesRecursive checks if a directory or subdirectories contain audio files
func (h *SetupHandler) hasAudioFilesRecursive(path string) bool {
	audioExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".aac":  true,
		".opus": true,
		".wma":  true,
	}

	found := false
	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			if audioExtensions[ext] {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})

	return found
}
