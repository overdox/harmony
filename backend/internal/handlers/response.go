package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response wrapper
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta contains metadata like pagination
type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination information
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
	HasMore    bool  `json:"hasMore"`
}

// PaginationParams holds pagination request parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// DefaultPagination returns default pagination parameters
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 20,
	}
}

// ParsePagination parses pagination parameters from the request
func ParsePagination(c *gin.Context) PaginationParams {
	params := DefaultPagination()

	if page := c.Query("page"); page != "" {
		if p, err := parseInt(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := parseInt(limit); err == nil && l > 0 && l <= 100 {
			params.Limit = l
		}
	}

	return params
}

// NewPagination creates pagination info from total count
func NewPagination(page, limit int, total int64) *Pagination {
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}
}

// Success sends a successful response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithPagination sends a successful response with pagination
func SuccessWithPagination(c *gin.Context, data interface{}, pagination *Pagination) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Pagination: pagination,
		},
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails sends an error response with details
func ErrorWithDetails(c *gin.Context, status int, code, message, details string) {
	c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// Conflict sends a 409 Conflict error
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{}
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

type parseError struct{}

func (e *parseError) Error() string { return "parse error" }

// Link represents a hypermedia link
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// TrackResponse extends track data with links
type TrackResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Duration    int     `json:"duration"`
	TrackNumber int     `json:"trackNumber"`
	DiscNumber  int     `json:"discNumber"`
	Format      string  `json:"format"`
	Bitrate     int     `json:"bitrate,omitempty"`
	AlbumID     string  `json:"albumId,omitempty"`
	ArtistID    string  `json:"artistId,omitempty"`
	Genre       string  `json:"genre,omitempty"`
	Year        int     `json:"year,omitempty"`
	Links       []Link  `json:"links,omitempty"`
}

// AlbumResponse extends album data with links
type AlbumResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Year        int     `json:"year,omitempty"`
	ArtistID    string  `json:"artistId"`
	ArtistName  string  `json:"artistName,omitempty"`
	TrackCount  int     `json:"trackCount,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	CoverArtURL string  `json:"coverArtUrl,omitempty"`
	Links       []Link  `json:"links,omitempty"`
}

// ArtistResponse extends artist data with links
type ArtistResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Bio        string `json:"bio,omitempty"`
	ImageURL   string `json:"imageUrl,omitempty"`
	AlbumCount int    `json:"albumCount,omitempty"`
	TrackCount int    `json:"trackCount,omitempty"`
	Links      []Link `json:"links,omitempty"`
}

// BuildTrackLinks generates hypermedia links for a track
func BuildTrackLinks(baseURL, trackID, albumID string) []Link {
	links := []Link{
		{Href: baseURL + "/api/v1/tracks/" + trackID, Rel: "self"},
		{Href: baseURL + "/api/v1/tracks/" + trackID + "/stream", Rel: "stream"},
	}
	if albumID != "" {
		links = append(links, Link{Href: baseURL + "/api/v1/albums/" + albumID, Rel: "album"})
	}
	return links
}

// BuildAlbumLinks generates hypermedia links for an album
func BuildAlbumLinks(baseURL, albumID, artistID string) []Link {
	links := []Link{
		{Href: baseURL + "/api/v1/albums/" + albumID, Rel: "self"},
		{Href: baseURL + "/api/v1/artwork/album/" + albumID, Rel: "artwork"},
	}
	if artistID != "" {
		links = append(links, Link{Href: baseURL + "/api/v1/artists/" + artistID, Rel: "artist"})
	}
	return links
}

// BuildArtistLinks generates hypermedia links for an artist
func BuildArtistLinks(baseURL, artistID string) []Link {
	return []Link{
		{Href: baseURL + "/api/v1/artists/" + artistID, Rel: "self"},
		{Href: baseURL + "/api/v1/artists/" + artistID + "/albums", Rel: "albums"},
	}
}
