package handlers

import (
	"github.com/gin-gonic/gin"

	"harmony/internal/database"
)

// SearchHandler handles search and discovery endpoints
type SearchHandler struct {
	trackRepo  *database.TrackRepository
	albumRepo  *database.AlbumRepository
	artistRepo *database.ArtistRepository
	redis      *database.RedisClient
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(
	trackRepo *database.TrackRepository,
	albumRepo *database.AlbumRepository,
	artistRepo *database.ArtistRepository,
	redis *database.RedisClient,
) *SearchHandler {
	return &SearchHandler{
		trackRepo:  trackRepo,
		albumRepo:  albumRepo,
		artistRepo: artistRepo,
		redis:      redis,
	}
}

// SearchResponse represents global search results
type SearchResponse struct {
	Query   string           `json:"query"`
	Tracks  []TrackResponse  `json:"tracks"`
	Albums  []AlbumResponse  `json:"albums"`
	Artists []ArtistResponse `json:"artists"`
}

// Search handles GET /api/v1/search
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		BadRequest(c, "search query required")
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	ctx := c.Request.Context()

	// Try to get cached results
	if h.redis != nil {
		var cached SearchResponse
		if err := h.redis.GetCachedSearchResults(ctx, query, &cached); err == nil {
			Success(c, cached)
			return
		}
	}

	// Search tracks
	tracks, _ := h.trackRepo.Search(ctx, query, limit)
	trackResponses := make([]TrackResponse, len(tracks))
	for i, track := range tracks {
		trackResponses[i] = TrackResponse{
			ID:       track.ID,
			Title:    track.Title,
			Duration: track.Duration,
			Format:   track.Format,
			AlbumID:  track.AlbumID,
			ArtistID: track.ArtistID,
		}
	}

	// Search albums
	albums, _ := h.albumRepo.Search(ctx, query, limit)
	albumResponses := make([]AlbumResponse, len(albums))
	for i, album := range albums {
		albumResponses[i] = AlbumResponse{
			ID:       album.ID,
			Title:    album.Title,
			Year:     album.Year,
			ArtistID: album.ArtistID,
		}
		if album.Artist != nil {
			albumResponses[i].ArtistName = album.Artist.Name
		}
	}

	// Search artists
	artists, _ := h.artistRepo.Search(ctx, query, limit)
	artistResponses := make([]ArtistResponse, len(artists))
	for i, artist := range artists {
		artistResponses[i] = ArtistResponse{
			ID:   artist.ID,
			Name: artist.Name,
		}
	}

	response := SearchResponse{
		Query:   query,
		Tracks:  trackResponses,
		Albums:  albumResponses,
		Artists: artistResponses,
	}

	// Cache results
	if h.redis != nil {
		h.redis.CacheSearchResults(ctx, query, response)
	}

	Success(c, response)
}

// Recent handles GET /api/v1/recent
func (h *SearchHandler) Recent(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ctx := c.Request.Context()
	resourceType := c.DefaultQuery("type", "tracks")

	switch resourceType {
	case "albums":
		albums, err := h.albumRepo.GetRecentlyAdded(ctx, limit)
		if err != nil {
			InternalError(c, "failed to get recent albums")
			return
		}

		response := make([]AlbumResponse, len(albums))
		for i, album := range albums {
			response[i] = AlbumResponse{
				ID:       album.ID,
				Title:    album.Title,
				Year:     album.Year,
				ArtistID: album.ArtistID,
			}
			if album.Artist != nil {
				response[i].ArtistName = album.Artist.Name
			}
		}
		Success(c, response)

	default: // tracks
		tracks, err := h.trackRepo.GetRecentlyAdded(ctx, limit)
		if err != nil {
			InternalError(c, "failed to get recent tracks")
			return
		}

		response := make([]TrackResponse, len(tracks))
		for i, track := range tracks {
			response[i] = TrackResponse{
				ID:       track.ID,
				Title:    track.Title,
				Duration: track.Duration,
				Format:   track.Format,
				AlbumID:  track.AlbumID,
				ArtistID: track.ArtistID,
			}
		}
		Success(c, response)
	}
}

// Random handles GET /api/v1/random
func (h *SearchHandler) Random(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseInt(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ctx := c.Request.Context()
	resourceType := c.DefaultQuery("type", "tracks")

	switch resourceType {
	case "albums":
		albums, err := h.albumRepo.GetRandom(ctx, limit)
		if err != nil {
			InternalError(c, "failed to get random albums")
			return
		}

		response := make([]AlbumResponse, len(albums))
		for i, album := range albums {
			response[i] = AlbumResponse{
				ID:       album.ID,
				Title:    album.Title,
				Year:     album.Year,
				ArtistID: album.ArtistID,
			}
			if album.Artist != nil {
				response[i].ArtistName = album.Artist.Name
			}
		}
		Success(c, response)

	default: // tracks
		tracks, err := h.trackRepo.GetRandom(ctx, limit)
		if err != nil {
			InternalError(c, "failed to get random tracks")
			return
		}

		response := make([]TrackResponse, len(tracks))
		for i, track := range tracks {
			response[i] = TrackResponse{
				ID:       track.ID,
				Title:    track.Title,
				Duration: track.Duration,
				Format:   track.Format,
				AlbumID:  track.AlbumID,
				ArtistID: track.ArtistID,
			}
		}
		Success(c, response)
	}
}
