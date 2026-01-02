package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
)

// AlbumHandler handles album-related endpoints
type AlbumHandler struct {
	repo    *database.AlbumRepository
	baseURL string
}

// NewAlbumHandler creates a new AlbumHandler
func NewAlbumHandler(repo *database.AlbumRepository, baseURL string) *AlbumHandler {
	return &AlbumHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

// List handles GET /api/v1/albums
func (h *AlbumHandler) List(c *gin.Context) {
	pagination := ParsePagination(c)

	opts := database.AlbumListOptions{
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Filter: database.AlbumFilter{
			ArtistID: c.Query("artistId"),
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

	albums, total, err := h.repo.List(c.Request.Context(), opts)
	if err != nil {
		InternalError(c, "failed to list albums")
		return
	}

	// Build response with links
	response := make([]AlbumResponse, len(albums))
	for i, album := range albums {
		response[i] = AlbumResponse{
			ID:          album.ID,
			Title:       album.Title,
			Year:        album.Year,
			ArtistID:    album.ArtistID,
			TrackCount:  album.TrackCount,
			Duration:    album.Duration,
			CoverArtURL: h.baseURL + "/api/v1/artwork/album/" + album.ID,
			Links:       BuildAlbumLinks(h.baseURL, album.ID, album.ArtistID),
		}

		// Include artist name if preloaded
		if album.Artist != nil {
			response[i].ArtistName = album.Artist.Name
		}
	}

	SuccessWithPagination(c, response, NewPagination(pagination.Page, pagination.Limit, total))
}

// Get handles GET /api/v1/albums/:id
func (h *AlbumHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "album ID required")
		return
	}

	// Get album with tracks
	album, err := h.repo.FindByIDWithTracks(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrAlbumNotFound) {
			NotFound(c, "album")
			return
		}
		InternalError(c, "failed to get album")
		return
	}

	// Build track responses
	tracks := make([]TrackResponse, len(album.Tracks))
	for i, track := range album.Tracks {
		tracks[i] = TrackResponse{
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

	response := struct {
		AlbumResponse
		Tracks []TrackResponse `json:"tracks"`
	}{
		AlbumResponse: AlbumResponse{
			ID:          album.ID,
			Title:       album.Title,
			Year:        album.Year,
			ArtistID:    album.ArtistID,
			TrackCount:  album.TrackCount,
			Duration:    album.Duration,
			CoverArtURL: h.baseURL + "/api/v1/artwork/album/" + album.ID,
			Links:       BuildAlbumLinks(h.baseURL, album.ID, album.ArtistID),
		},
		Tracks: tracks,
	}

	// Include artist name if preloaded
	if album.Artist != nil {
		response.ArtistName = album.Artist.Name
	}

	Success(c, response)
}
