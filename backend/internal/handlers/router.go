package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"harmony/internal/database"
	"harmony/internal/services"
	"harmony/internal/transcoder"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	AllowedOrigins []string
	MediaRoot      string
	CacheDir       string
	BaseURL        string
}

// DefaultRouterConfig returns default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
		MediaRoot:      "./media",
		CacheDir:       "./data/cache",
		BaseURL:        "http://localhost:8080",
	}
}

// Handlers holds all handler instances
type Handlers struct {
	Track    *TrackHandler
	Album    *AlbumHandler
	Artist   *ArtistHandler
	Playlist *PlaylistHandler
	Search   *SearchHandler
	Library  *LibraryHandler
	Stream   *StreamHandler
	Artwork  *ArtworkHandler
}

// NewRouter creates and configures the Gin router
func NewRouter(
	cfg RouterConfig,
	db *database.Database,
	redis *database.RedisClient,
	trans *transcoder.Transcoder,
	libService *services.LibraryService,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(configureCORS(cfg.AllowedOrigins))

	// Create repositories
	trackRepo := database.NewTrackRepository(db.DB)
	albumRepo := database.NewAlbumRepository(db.DB)
	artistRepo := database.NewArtistRepository(db.DB)
	playlistRepo := database.NewPlaylistRepository(db.DB)

	// Create handlers
	handlers := &Handlers{
		Track:    NewTrackHandler(trackRepo, cfg.BaseURL),
		Album:    NewAlbumHandler(albumRepo, cfg.BaseURL),
		Artist:   NewArtistHandler(artistRepo, cfg.BaseURL),
		Playlist: NewPlaylistHandler(playlistRepo),
		Search:   NewSearchHandler(trackRepo, albumRepo, artistRepo, redis),
		Library:  NewLibraryHandler(libService),
		Stream:   NewStreamHandler(trackRepo, trans, cfg.MediaRoot),
		Artwork:  NewArtworkHandler(cfg.CacheDir),
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Track routes
		tracks := v1.Group("/tracks")
		{
			tracks.GET("", handlers.Track.List)
			tracks.GET("/:id", handlers.Track.Get)
			tracks.GET("/:id/stream", handlers.Stream.Stream)
		}

		// Album routes
		albums := v1.Group("/albums")
		{
			albums.GET("", handlers.Album.List)
			albums.GET("/:id", handlers.Album.Get)
		}

		// Artist routes
		artists := v1.Group("/artists")
		{
			artists.GET("", handlers.Artist.List)
			artists.GET("/:id", handlers.Artist.Get)
		}

		// Playlist routes
		playlists := v1.Group("/playlists")
		{
			playlists.GET("", handlers.Playlist.List)
			playlists.POST("", handlers.Playlist.Create)
			playlists.GET("/:id", handlers.Playlist.Get)
			playlists.PUT("/:id", handlers.Playlist.Update)
			playlists.DELETE("/:id", handlers.Playlist.Delete)
			playlists.POST("/:id/tracks", handlers.Playlist.AddTrack)
			playlists.PUT("/:id/tracks/reorder", handlers.Playlist.ReorderTracks)
			playlists.DELETE("/:id/tracks/:trackId", handlers.Playlist.RemoveTrack)
		}

		// Search & Discovery routes
		v1.GET("/search", handlers.Search.Search)
		v1.GET("/recent", handlers.Search.Recent)
		v1.GET("/random", handlers.Search.Random)

		// Library management routes
		library := v1.Group("/library")
		{
			library.POST("/scan", handlers.Library.Scan)
			library.GET("/scan/status", handlers.Library.ScanStatus)
			library.POST("/scan/cancel", handlers.Library.CancelScan)
			library.GET("/stats", handlers.Library.Stats)
		}

		// Artwork routes
		v1.GET("/artwork/:type/:id", handlers.Artwork.Get)
	}

	return router
}

// requestLogger returns a middleware that logs requests
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		slog.Info("request",
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"latency", latency.String(),
			"ip", c.ClientIP(),
		)
	}
}

// configureCORS returns CORS middleware configuration
func configureCORS(allowedOrigins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range", "Accept-Ranges"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// If no origins specified, allow all (development mode)
	if len(allowedOrigins) == 0 {
		config.AllowAllOrigins = true
	}

	return cors.New(config)
}

// RateLimiter is a simple rate limiter middleware (optional)
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Middleware returns the rate limiter middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Clean old requests
		var valid []time.Time
		for _, t := range rl.requests[ip] {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		rl.requests[ip] = valid

		// Check limit
		if len(rl.requests[ip]) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[ip] = append(rl.requests[ip], now)

		c.Next()
	}
}
