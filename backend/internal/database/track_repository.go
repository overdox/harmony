package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"harmony/internal/models"
)

var (
	ErrTrackNotFound = errors.New("track not found")
)

type TrackRepository struct {
	db *gorm.DB
}

func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

type TrackFilter struct {
	AlbumID  string
	ArtistID string
	Genre    string
	Year     int
	Query    string
}

type TrackListOptions struct {
	Filter TrackFilter
	Page   int
	Limit  int
	SortBy string
	Order  string
}

func (r *TrackRepository) Create(ctx context.Context, track *models.Track) error {
	if err := r.db.WithContext(ctx).Create(track).Error; err != nil {
		return fmt.Errorf("creating track: %w", err)
	}
	return nil
}

func (r *TrackRepository) CreateBatch(ctx context.Context, tracks []models.Track) error {
	if len(tracks) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(tracks, 100).Error; err != nil {
		return fmt.Errorf("creating tracks batch: %w", err)
	}
	return nil
}

func (r *TrackRepository) FindByID(ctx context.Context, id string) (*models.Track, error) {
	var track models.Track
	result := r.db.WithContext(ctx).
		Preload("Album").
		Preload("Artist").
		First(&track, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTrackNotFound
		}
		return nil, fmt.Errorf("finding track: %w", result.Error)
	}
	return &track, nil
}

func (r *TrackRepository) FindByFilePath(ctx context.Context, filePath string) (*models.Track, error) {
	var track models.Track
	result := r.db.WithContext(ctx).First(&track, "file_path = ?", filePath)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTrackNotFound
		}
		return nil, fmt.Errorf("finding track by path: %w", result.Error)
	}
	return &track, nil
}

func (r *TrackRepository) List(ctx context.Context, opts TrackListOptions) ([]models.Track, int64, error) {
	var tracks []models.Track
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Track{})

	// Apply filters
	if opts.Filter.AlbumID != "" {
		query = query.Where("album_id = ?", opts.Filter.AlbumID)
	}
	if opts.Filter.ArtistID != "" {
		query = query.Where("artist_id = ?", opts.Filter.ArtistID)
	}
	if opts.Filter.Genre != "" {
		query = query.Where("genre = ?", opts.Filter.Genre)
	}
	if opts.Filter.Year > 0 {
		query = query.Where("year = ?", opts.Filter.Year)
	}
	if opts.Filter.Query != "" {
		searchQuery := "%" + opts.Filter.Query + "%"
		query = query.Where("title LIKE ?", searchQuery)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting tracks: %w", err)
	}

	// Apply sorting
	sortBy := "title"
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

	// Execute query with preloads
	if err := query.Preload("Album").Preload("Artist").Find(&tracks).Error; err != nil {
		return nil, 0, fmt.Errorf("listing tracks: %w", err)
	}

	return tracks, total, nil
}

func (r *TrackRepository) Search(ctx context.Context, query string, limit int) ([]models.Track, error) {
	var tracks []models.Track
	searchQuery := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Preload("Album").
		Preload("Artist").
		Where("title LIKE ?", searchQuery).
		Limit(limit).
		Find(&tracks).Error

	if err != nil {
		return nil, fmt.Errorf("searching tracks: %w", err)
	}
	return tracks, nil
}

func (r *TrackRepository) Update(ctx context.Context, track *models.Track) error {
	if err := r.db.WithContext(ctx).Save(track).Error; err != nil {
		return fmt.Errorf("updating track: %w", err)
	}
	return nil
}

func (r *TrackRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Track{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("deleting track: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTrackNotFound
	}
	return nil
}

func (r *TrackRepository) DeleteByFilePath(ctx context.Context, filePath string) error {
	result := r.db.WithContext(ctx).Delete(&models.Track{}, "file_path = ?", filePath)
	if result.Error != nil {
		return fmt.Errorf("deleting track by path: %w", result.Error)
	}
	return nil
}

func (r *TrackRepository) GetRecentlyAdded(ctx context.Context, limit int) ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.WithContext(ctx).
		Preload("Album").
		Preload("Artist").
		Order("created_at DESC").
		Limit(limit).
		Find(&tracks).Error

	if err != nil {
		return nil, fmt.Errorf("getting recent tracks: %w", err)
	}
	return tracks, nil
}

func (r *TrackRepository) GetRandom(ctx context.Context, limit int) ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.WithContext(ctx).
		Preload("Album").
		Preload("Artist").
		Order("RANDOM()").
		Limit(limit).
		Find(&tracks).Error

	if err != nil {
		return nil, fmt.Errorf("getting random tracks: %w", err)
	}
	return tracks, nil
}

func (r *TrackRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Track{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting tracks: %w", err)
	}
	return count, nil
}

func (r *TrackRepository) GetAllFilePaths(ctx context.Context) ([]string, error) {
	var paths []string
	err := r.db.WithContext(ctx).
		Model(&models.Track{}).
		Pluck("file_path", &paths).Error

	if err != nil {
		return nil, fmt.Errorf("getting file paths: %w", err)
	}
	return paths, nil
}
