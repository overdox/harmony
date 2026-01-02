package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"harmony/internal/scanner"
)

// ArtworkHandler handles artwork serving endpoints
type ArtworkHandler struct {
	processor *scanner.ArtworkProcessor
	cacheDir  string
}

// NewArtworkHandler creates a new ArtworkHandler
func NewArtworkHandler(cacheDir string) *ArtworkHandler {
	return &ArtworkHandler{
		processor: scanner.NewArtworkProcessor(cacheDir),
		cacheDir:  cacheDir,
	}
}

// Get handles GET /api/v1/artwork/:type/:id
func (h *ArtworkHandler) Get(c *gin.Context) {
	artType := c.Param("type")
	id := c.Param("id")

	if artType == "" || id == "" {
		BadRequest(c, "type and ID required")
		return
	}

	// Get size parameter (default to medium)
	size := c.DefaultQuery("size", "medium")
	validSizes := map[string]bool{
		"thumbnail": true,
		"small":     true,
		"medium":    true,
		"large":     true,
		"original":  true,
	}
	if !validSizes[size] {
		size = "medium"
	}

	var artworkPath string

	switch artType {
	case "album":
		artworkPath = h.processor.GetArtworkPath(id, size)
	case "artist":
		// Artist images stored differently
		artworkPath = filepath.Join(h.cacheDir, "artists", id, size+".jpg")
	case "playlist":
		// Playlist cover images
		artworkPath = filepath.Join(h.cacheDir, "playlists", id, size+".jpg")
	default:
		BadRequest(c, "invalid artwork type")
		return
	}

	// Check if file exists
	if _, err := os.Stat(artworkPath); os.IsNotExist(err) {
		// Return a 1x1 transparent placeholder to avoid 404 spam
		// The frontend should handle this gracefully with CSS fallback
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("Content-Type", "image/svg+xml")
		c.String(200, `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="300" viewBox="0 0 300 300"><rect fill="#1a1a2e" width="300" height="300"/><text x="150" y="160" font-family="Arial" font-size="48" fill="#4a4a6a" text-anchor="middle">♪</text></svg>`)
		return
	}

	// Set cache headers
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Type", "image/jpeg")

	// Serve the file
	c.File(artworkPath)
}

// GetAlbumArtwork is a convenience method for album artwork
func (h *ArtworkHandler) GetAlbumArtwork(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "album ID required")
		return
	}

	size := c.DefaultQuery("size", "medium")
	artworkPath := h.processor.GetArtworkPath(id, size)

	if _, err := os.Stat(artworkPath); os.IsNotExist(err) {
		// Return SVG placeholder for missing artwork
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("Content-Type", "image/svg+xml")
		c.String(200, `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="300" viewBox="0 0 300 300"><rect fill="#1a1a2e" width="300" height="300"/><text x="150" y="160" font-family="Arial" font-size="48" fill="#4a4a6a" text-anchor="middle">♪</text></svg>`)
		return
	}

	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(artworkPath)
}

// Upload handles artwork upload (for playlists, etc.)
func (h *ArtworkHandler) Upload(c *gin.Context) {
	artType := c.Param("type")
	id := c.Param("id")

	if artType == "" || id == "" {
		BadRequest(c, "type and ID required")
		return
	}

	// Only allow playlist artwork uploads for now
	if artType != "playlist" {
		Forbidden(c, "cannot upload artwork for this type")
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("artwork")
	if err != nil {
		BadRequest(c, "artwork file required")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !validTypes[contentType] {
		BadRequest(c, "invalid image type (jpeg, png, or webp required)")
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5*1024*1024 {
		BadRequest(c, "image too large (max 5MB)")
		return
	}

	// Save and process artwork
	if err := h.processor.SaveArtworkFromReader(id, file, contentType); err != nil {
		InternalError(c, "failed to save artwork")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "artwork uploaded successfully",
	})
}

// Delete handles artwork deletion
func (h *ArtworkHandler) Delete(c *gin.Context) {
	artType := c.Param("type")
	id := c.Param("id")

	if artType == "" || id == "" {
		BadRequest(c, "type and ID required")
		return
	}

	// Only allow playlist artwork deletion for now
	if artType != "playlist" {
		Forbidden(c, "cannot delete artwork for this type")
		return
	}

	if err := h.processor.DeleteArtwork(id); err != nil {
		InternalError(c, "failed to delete artwork")
		return
	}

	NoContent(c)
}
