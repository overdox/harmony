package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"harmony/internal/database"
)

// ArtistHandler handles artist-related endpoints
type ArtistHandler struct {
	repo    *database.ArtistRepository
	baseURL string
}

// NewArtistHandler creates a new ArtistHandler
func NewArtistHandler(repo *database.ArtistRepository, baseURL string) *ArtistHandler {
	return &ArtistHandler{
		repo:    repo,
		baseURL: baseURL,
	}
}

// List handles GET /api/v1/artists
func (h *ArtistHandler) List(c *gin.Context) {
	pagination := ParsePagination(c)

	opts := database.ArtistListOptions{
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Filter: database.ArtistFilter{
			Query: c.Query("q"),
		},
		SortBy: c.DefaultQuery("sortBy", "name"),
		Order:  c.DefaultQuery("order", "asc"),
	}

	artists, total, err := h.repo.List(c.Request.Context(), opts)
	if err != nil {
		InternalError(c, "failed to list artists")
		return
	}

	// Build response with links
	response := make([]ArtistResponse, len(artists))
	for i, artist := range artists {
		response[i] = ArtistResponse{
			ID:       artist.ID,
			Name:     artist.Name,
			Bio:      artist.Bio,
			ImageURL: artist.ImageURL,
			Links:    BuildArtistLinks(h.baseURL, artist.ID),
		}
	}

	SuccessWithPagination(c, response, NewPagination(pagination.Page, pagination.Limit, total))
}

// Get handles GET /api/v1/artists/:id
func (h *ArtistHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "artist ID required")
		return
	}

	// Get artist with albums
	artist, err := h.repo.FindByIDWithAlbums(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, database.ErrArtistNotFound) {
			NotFound(c, "artist")
			return
		}
		InternalError(c, "failed to get artist")
		return
	}

	// Build album responses
	albums := make([]AlbumResponse, len(artist.Albums))
	for i, album := range artist.Albums {
		albums[i] = AlbumResponse{
			ID:          album.ID,
			Title:       album.Title,
			Year:        album.Year,
			ArtistID:    album.ArtistID,
			ArtistName:  artist.Name,
			CoverArtURL: h.baseURL + "/api/v1/artwork/album/" + album.ID,
			Links:       BuildAlbumLinks(h.baseURL, album.ID, album.ArtistID),
		}
	}

	// Get popular tracks
	popularTracks, _ := h.repo.GetPopularTracks(c.Request.Context(), id, 10)
	tracks := make([]TrackResponse, len(popularTracks))
	for i, track := range popularTracks {
		tracks[i] = TrackResponse{
			ID:          track.ID,
			Title:       track.Title,
			Duration:    track.Duration,
			TrackNumber: track.TrackNumber,
			Format:      track.Format,
			AlbumID:     track.AlbumID,
			Links:       BuildTrackLinks(h.baseURL, track.ID, track.AlbumID),
		}
	}

	response := struct {
		ArtistResponse
		Albums        []AlbumResponse `json:"albums"`
		PopularTracks []TrackResponse `json:"popularTracks"`
	}{
		ArtistResponse: ArtistResponse{
			ID:         artist.ID,
			Name:       artist.Name,
			Bio:        artist.Bio,
			ImageURL:   artist.ImageURL,
			AlbumCount: len(artist.Albums),
			Links:      BuildArtistLinks(h.baseURL, artist.ID),
		},
		Albums:        albums,
		PopularTracks: tracks,
	}

	Success(c, response)
}
