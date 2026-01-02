package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
	"harmony/internal/models"
)

// PlaylistHandler handles playlist-related endpoints
type PlaylistHandler struct {
	repo *database.PlaylistRepository
}

// NewPlaylistHandler creates a new PlaylistHandler
func NewPlaylistHandler(repo *database.PlaylistRepository) *PlaylistHandler {
	return &PlaylistHandler{repo: repo}
}

// CreatePlaylistRequest represents a playlist creation request
type CreatePlaylistRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	IsPublic    bool   `json:"isPublic"`
}

// UpdatePlaylistRequest represents a playlist update request
type UpdatePlaylistRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	IsPublic    *bool   `json:"isPublic"`
}

// AddTrackRequest represents a request to add a track to a playlist
type AddTrackRequest struct {
	TrackID string `json:"trackId" binding:"required"`
}

// PlaylistResponse represents a playlist in API responses
type PlaylistResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	IsPublic    bool            `json:"isPublic"`
	TrackCount  int             `json:"trackCount"`
	Duration    int             `json:"duration"`
	UserID      string          `json:"userId"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
	Tracks      []TrackResponse `json:"tracks,omitempty"`
}

// List handles GET /api/v1/playlists
func (h *PlaylistHandler) List(c *gin.Context) {
	pagination := ParsePagination(c)

	// TODO: Get user ID from auth context
	userID := c.Query("userId")

	opts := database.PlaylistListOptions{
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Filter: database.PlaylistFilter{
			UserID: userID,
			Query:  c.Query("q"),
		},
		SortBy: c.DefaultQuery("sortBy", "name"),
		Order:  c.DefaultQuery("order", "asc"),
	}

	playlists, total, err := h.repo.List(c.Request.Context(), opts)
	if err != nil {
		InternalError(c, "failed to list playlists")
		return
	}

	response := make([]PlaylistResponse, len(playlists))
	for i, playlist := range playlists {
		response[i] = PlaylistResponse{
			ID:          playlist.ID,
			Name:        playlist.Name,
			Description: playlist.Description,
			IsPublic:    playlist.IsPublic,
			TrackCount:  playlist.TrackCount,
			Duration:    playlist.Duration,
			UserID:      playlist.UserID,
			CreatedAt:   playlist.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   playlist.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	SuccessWithPagination(c, response, NewPagination(pagination.Page, pagination.Limit, total))
}

// Create handles POST /api/v1/playlists
func (h *PlaylistHandler) Create(c *gin.Context) {
	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	// TODO: Get user ID from auth context
	userID := "default-user"

	playlist := &models.Playlist{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UserID:      userID,
	}

	if err := h.repo.Create(c.Request.Context(), playlist); err != nil {
		InternalError(c, "failed to create playlist")
		return
	}

	response := PlaylistResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsPublic:    playlist.IsPublic,
		TrackCount:  0,
		Duration:    0,
		UserID:      playlist.UserID,
		CreatedAt:   playlist.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   playlist.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	Created(c, response)
}

// Get handles GET /api/v1/playlists/:id
func (h *PlaylistHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "playlist ID required")
		return
	}

	playlist, err := h.repo.FindByIDWithTracks(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrPlaylistNotFound) {
			NotFound(c, "playlist")
			return
		}
		InternalError(c, "failed to get playlist")
		return
	}

	// Build track responses
	tracks := make([]TrackResponse, len(playlist.Tracks))
	for i, track := range playlist.Tracks {
		tracks[i] = TrackResponse{
			ID:          track.ID,
			Title:       track.Title,
			Duration:    track.Duration,
			TrackNumber: track.TrackNumber,
			Format:      track.Format,
			AlbumID:     track.AlbumID,
			ArtistID:    track.ArtistID,
		}
	}

	response := PlaylistResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsPublic:    playlist.IsPublic,
		TrackCount:  playlist.TrackCount,
		Duration:    playlist.Duration,
		UserID:      playlist.UserID,
		CreatedAt:   playlist.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   playlist.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Tracks:      tracks,
	}

	Success(c, response)
}

// Update handles PUT /api/v1/playlists/:id
func (h *PlaylistHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "playlist ID required")
		return
	}

	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	playlist, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrPlaylistNotFound) {
			NotFound(c, "playlist")
			return
		}
		InternalError(c, "failed to get playlist")
		return
	}

	// Apply updates
	if req.Name != nil {
		playlist.Name = *req.Name
	}
	if req.Description != nil {
		playlist.Description = *req.Description
	}
	if req.IsPublic != nil {
		playlist.IsPublic = *req.IsPublic
	}

	if err := h.repo.Update(c.Request.Context(), playlist); err != nil {
		InternalError(c, "failed to update playlist")
		return
	}

	response := PlaylistResponse{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsPublic:    playlist.IsPublic,
		UserID:      playlist.UserID,
		CreatedAt:   playlist.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   playlist.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	Success(c, response)
}

// Delete handles DELETE /api/v1/playlists/:id
func (h *PlaylistHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "playlist ID required")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, database.ErrPlaylistNotFound) {
			NotFound(c, "playlist")
			return
		}
		InternalError(c, "failed to delete playlist")
		return
	}

	NoContent(c)
}

// AddTrack handles POST /api/v1/playlists/:id/tracks
func (h *PlaylistHandler) AddTrack(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "playlist ID required")
		return
	}

	var req AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	// Check if playlist exists
	_, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrPlaylistNotFound) {
			NotFound(c, "playlist")
			return
		}
		InternalError(c, "failed to get playlist")
		return
	}

	// Check if track already in playlist
	exists, _ := h.repo.HasTrack(c.Request.Context(), id, req.TrackID)
	if exists {
		Conflict(c, "track already in playlist")
		return
	}

	if err := h.repo.AddTrack(c.Request.Context(), id, req.TrackID); err != nil {
		InternalError(c, "failed to add track to playlist")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "track added to playlist",
	})
}

// RemoveTrack handles DELETE /api/v1/playlists/:id/tracks/:trackId
func (h *PlaylistHandler) RemoveTrack(c *gin.Context) {
	playlistID := c.Param("id")
	trackID := c.Param("trackId")

	if playlistID == "" || trackID == "" {
		BadRequest(c, "playlist ID and track ID required")
		return
	}

	if err := h.repo.RemoveTrack(c.Request.Context(), playlistID, trackID); err != nil {
		if errors.Is(err, database.ErrTrackNotInPlaylist) {
			NotFound(c, "track in playlist")
			return
		}
		InternalError(c, "failed to remove track from playlist")
		return
	}

	NoContent(c)
}
