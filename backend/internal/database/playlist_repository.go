package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"harmony/internal/models"
)

var (
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrTrackNotInPlaylist = errors.New("track not in playlist")
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

type PlaylistFilter struct {
	UserID   string
	IsPublic *bool
	Query    string
}

type PlaylistListOptions struct {
	Filter PlaylistFilter
	Page   int
	Limit  int
	SortBy string
	Order  string
}

func (r *PlaylistRepository) Create(ctx context.Context, playlist *models.Playlist) error {
	if playlist.ID == "" {
		playlist.ID = GenerateID()
	}
	if err := r.db.WithContext(ctx).Create(playlist).Error; err != nil {
		return fmt.Errorf("creating playlist: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) FindByID(ctx context.Context, id string) (*models.Playlist, error) {
	var playlist models.Playlist
	result := r.db.WithContext(ctx).
		Preload("User").
		First(&playlist, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, fmt.Errorf("finding playlist: %w", result.Error)
	}
	return &playlist, nil
}

func (r *PlaylistRepository) FindByIDWithTracks(ctx context.Context, id string) (*models.Playlist, error) {
	var playlist models.Playlist
	result := r.db.WithContext(ctx).
		Preload("User").
		Preload("PlaylistTracks", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("PlaylistTracks.Track").
		Preload("PlaylistTracks.Track.Album").
		Preload("PlaylistTracks.Track.Artist").
		First(&playlist, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPlaylistNotFound
		}
		return nil, fmt.Errorf("finding playlist with tracks: %w", result.Error)
	}

	// Build tracks slice from PlaylistTracks
	playlist.Tracks = make([]models.Track, len(playlist.PlaylistTracks))
	for i, pt := range playlist.PlaylistTracks {
		if pt.Track != nil {
			playlist.Tracks[i] = *pt.Track
		}
	}

	// Calculate totals
	playlist.TrackCount = len(playlist.Tracks)
	for _, track := range playlist.Tracks {
		playlist.Duration += track.Duration
	}

	return &playlist, nil
}

func (r *PlaylistRepository) List(ctx context.Context, opts PlaylistListOptions) ([]models.Playlist, int64, error) {
	var playlists []models.Playlist
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Playlist{})

	// Apply filters
	if opts.Filter.UserID != "" {
		query = query.Where("user_id = ?", opts.Filter.UserID)
	}
	if opts.Filter.IsPublic != nil {
		query = query.Where("is_public = ?", *opts.Filter.IsPublic)
	}
	if opts.Filter.Query != "" {
		searchQuery := "%" + opts.Filter.Query + "%"
		query = query.Where("name LIKE ?", searchQuery)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting playlists: %w", err)
	}

	// Apply sorting
	sortBy := "name"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	order := "ASC"
	if opts.Order == "desc" {
		order = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Page > 0 && opts.Limit > 0 {
		offset := (opts.Page - 1) * opts.Limit
		query = query.Offset(offset)
	}

	if err := query.Preload("User").Find(&playlists).Error; err != nil {
		return nil, 0, fmt.Errorf("listing playlists: %w", err)
	}

	// Get track counts
	for i := range playlists {
		var count int64
		r.db.WithContext(ctx).
			Model(&models.PlaylistTrack{}).
			Where("playlist_id = ?", playlists[i].ID).
			Count(&count)
		playlists[i].TrackCount = int(count)
	}

	return playlists, total, nil
}

func (r *PlaylistRepository) Update(ctx context.Context, playlist *models.Playlist) error {
	if err := r.db.WithContext(ctx).Save(playlist).Error; err != nil {
		return fmt.Errorf("updating playlist: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete playlist tracks first
		if err := tx.Delete(&models.PlaylistTrack{}, "playlist_id = ?", id).Error; err != nil {
			return fmt.Errorf("deleting playlist tracks: %w", err)
		}

		// Delete playlist
		result := tx.Delete(&models.Playlist{}, "id = ?", id)
		if result.Error != nil {
			return fmt.Errorf("deleting playlist: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrPlaylistNotFound
		}
		return nil
	})
}

func (r *PlaylistRepository) AddTrack(ctx context.Context, playlistID, trackID string) error {
	// Get current max position
	var maxPosition int
	r.db.WithContext(ctx).
		Model(&models.PlaylistTrack{}).
		Where("playlist_id = ?", playlistID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPosition)

	playlistTrack := &models.PlaylistTrack{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Position:   maxPosition + 1,
		AddedAt:    time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(playlistTrack).Error; err != nil {
		return fmt.Errorf("adding track to playlist: %w", err)
	}

	// Update playlist's updated_at
	r.db.WithContext(ctx).
		Model(&models.Playlist{}).
		Where("id = ?", playlistID).
		Update("updated_at", time.Now())

	return nil
}

func (r *PlaylistRepository) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	result := r.db.WithContext(ctx).
		Delete(&models.PlaylistTrack{}, "playlist_id = ? AND track_id = ?", playlistID, trackID)

	if result.Error != nil {
		return fmt.Errorf("removing track from playlist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTrackNotInPlaylist
	}

	// Reorder remaining tracks
	if err := r.reorderTracks(ctx, playlistID); err != nil {
		return err
	}

	return nil
}

func (r *PlaylistRepository) ReorderTracks(ctx context.Context, playlistID string, trackIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, trackID := range trackIDs {
			if err := tx.Model(&models.PlaylistTrack{}).
				Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
				Update("position", i+1).Error; err != nil {
				return fmt.Errorf("updating track position: %w", err)
			}
		}
		return nil
	})
}

func (r *PlaylistRepository) reorderTracks(ctx context.Context, playlistID string) error {
	var tracks []models.PlaylistTrack
	if err := r.db.WithContext(ctx).
		Where("playlist_id = ?", playlistID).
		Order("position ASC").
		Find(&tracks).Error; err != nil {
		return fmt.Errorf("getting playlist tracks: %w", err)
	}

	for i, track := range tracks {
		if track.Position != i+1 {
			if err := r.db.WithContext(ctx).
				Model(&models.PlaylistTrack{}).
				Where("playlist_id = ? AND track_id = ?", playlistID, track.TrackID).
				Update("position", i+1).Error; err != nil {
				return fmt.Errorf("updating track position: %w", err)
			}
		}
	}

	return nil
}

func (r *PlaylistRepository) GetByUser(ctx context.Context, userID string) ([]models.Playlist, error) {
	var playlists []models.Playlist
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("name ASC").
		Find(&playlists).Error

	if err != nil {
		return nil, fmt.Errorf("getting playlists by user: %w", err)
	}
	return playlists, nil
}

func (r *PlaylistRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Playlist{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting playlists: %w", err)
	}
	return count, nil
}

func (r *PlaylistRepository) HasTrack(ctx context.Context, playlistID, trackID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PlaylistTrack{}).
		Where("playlist_id = ? AND track_id = ?", playlistID, trackID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("checking track in playlist: %w", err)
	}
	return count > 0, nil
}
