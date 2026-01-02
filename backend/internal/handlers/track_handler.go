package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
)

// TrackHandler handles track-related endpoints
type TrackHandler struct {
	repo    *database.TrackRepository
	baseURL string
}

// NewTrackHandler creates a new TrackHandler
func NewTrackHandler(repo *database.TrackRepository, baseURL string) *TrackHandler {
	return &TrackHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

// List handles GET /api/v1/tracks
func (h *TrackHandler) List(c *gin.Context) {
	pagination := ParsePagination(c)

	opts := database.TrackListOptions{
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Filter: database.TrackFilter{
			AlbumID:  c.Query("albumId"),
			ArtistID: c.Query("artistId"),
			Genre:    c.Query("genre"),
			Query:    c.Query("q"),
		},
		SortBy: c.DefaultQuery("sortBy", "title"),
		Order:  c.DefaultQuery("order", "asc"),
	}

	// Parse year filter
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := parseInt(yearStr); err == nil {
			opts.Filter.Year = year
		}
	}

	tracks, total, err := h.repo.List(c.Request.Context(), opts)
	if err != nil {
		InternalError(c, "failed to list tracks")
		return
	}

	// Build response with links
	response := make([]TrackResponse, len(tracks))
	for i, track := range tracks {
		response[i] = TrackResponse{
			ID:          track.ID,
			Title:       track.Title,
			Duration:    track.Duration,
			TrackNumber: track.TrackNumber,
			DiscNumber:  track.DiscNumber,
			Format:      track.Format,
			Bitrate:     track.Bitrate,
			AlbumID:     track.AlbumID,
			ArtistID:    track.ArtistID,
			Genre:       track.Genre,
			Year:        track.Year,
			Links:       BuildTrackLinks(h.baseURL, track.ID, track.AlbumID),
		}
	}

	SuccessWithPagination(c, response, NewPagination(pagination.Page, pagination.Limit, total))
}

// Get handles GET /api/v1/tracks/:id
func (h *TrackHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "track ID required")
		return
	}

	track, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrTrackNotFound) {
			NotFound(c, "track")
			return
		}
		InternalError(c, "failed to get track")
		return
	}

	response := TrackResponse{
		ID:          track.ID,
		Title:       track.Title,
		Duration:    track.Duration,
		TrackNumber: track.TrackNumber,
		DiscNumber:  track.DiscNumber,
		Format:      track.Format,
		Bitrate:     track.Bitrate,
		AlbumID:     track.AlbumID,
		ArtistID:    track.ArtistID,
		Genre:       track.Genre,
		Year:        track.Year,
		Links:       BuildTrackLinks(h.baseURL, track.ID, track.AlbumID),
	}

	// Include album info if preloaded
	if track.Album != nil {
		response.Links = append(response.Links, Link{
			Href: h.baseURL + "/api/v1/albums/" + track.Album.ID,
			Rel:  "album",
		})
	}

	// Include artist info if preloaded
	if track.Artist != nil {
		response.Links = append(response.Links, Link{
			Href: h.baseURL + "/api/v1/artists/" + track.Artist.ID,
			Rel:  "artist",
		})
	}

	Success(c, response)
}
