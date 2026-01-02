package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"harmony/internal/models"
)

var (
	ErrAlbumNotFound = errors.New("album not found")
)

type AlbumRepository struct {
	db *gorm.DB
}

func NewAlbumRepository(db *gorm.DB) *AlbumRepository {
	return &AlbumRepository{db: db}
}

type AlbumFilter struct {
	ArtistID string
	Year     int
	Query    string
}

type AlbumListOptions struct {
	Filter AlbumFilter
	Page   int
	Limit  int
	SortBy string
	Order  string
}

func (r *AlbumRepository) Create(ctx context.Context, album *models.Album) error {
	if err := r.db.WithContext(ctx).Create(album).Error; err != nil {
		return fmt.Errorf("creating album: %w", err)
	}
	return nil
}

func (r *AlbumRepository) FindByID(ctx context.Context, id string) (*models.Album, error) {
	var album models.Album
	result := r.db.WithContext(ctx).
		Preload("Artist").
		First(&album, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAlbumNotFound
		}
		return nil, fmt.Errorf("finding album: %w", result.Error)
	}
	return &album, nil
}

func (r *AlbumRepository) FindByIDWithTracks(ctx context.Context, id string) (*models.Album, error) {
	var album models.Album
	result := r.db.WithContext(ctx).
		Preload("Artist").
		Preload("Tracks", func(db *gorm.DB) *gorm.DB {
			return db.Order("disc_number ASC, track_number ASC")
		}).
		First(&album, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAlbumNotFound
		}
		return nil, fmt.Errorf("finding album with tracks: %w", result.Error)
	}

	// Calculate totals
	album.TrackCount = len(album.Tracks)
	for _, track := range album.Tracks {
		album.Duration += track.Duration
	}

	return &album, nil
}

func (r *AlbumRepository) FindByTitleAndArtist(ctx context.Context, title, artistID string) (*models.Album, error) {
	var album models.Album
	result := r.db.WithContext(ctx).
		Where("title = ? AND artist_id = ?", title, artistID).
		First(&album)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrAlbumNotFound
		}
		return nil, fmt.Errorf("finding album by title and artist: %w", result.Error)
	}
	return &album, nil
}

func (r *AlbumRepository) List(ctx context.Context, opts AlbumListOptions) ([]models.Album, int64, error) {
	var albums []models.Album
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Album{})

	// Apply filters
	if opts.Filter.ArtistID != "" {
		query = query.Where("artist_id = ?", opts.Filter.ArtistID)
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
		return nil, 0, fmt.Errorf("counting albums: %w", err)
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
	if err := query.Preload("Artist").Find(&albums).Error; err != nil {
		return nil, 0, fmt.Errorf("listing albums: %w", err)
	}

	return albums, total, nil
}

func (r *AlbumRepository) Search(ctx context.Context, query string, limit int) ([]models.Album, error) {
	var albums []models.Album
	searchQuery := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Preload("Artist").
		Where("title LIKE ?", searchQuery).
		Limit(limit).
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("searching albums: %w", err)
	}
	return albums, nil
}

func (r *AlbumRepository) Update(ctx context.Context, album *models.Album) error {
	if err := r.db.WithContext(ctx).Save(album).Error; err != nil {
		return fmt.Errorf("updating album: %w", err)
	}
	return nil
}

func (r *AlbumRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Album{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("deleting album: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAlbumNotFound
	}
	return nil
}

func (r *AlbumRepository) GetRecentlyAdded(ctx context.Context, limit int) ([]models.Album, error) {
	var albums []models.Album
	err := r.db.WithContext(ctx).
		Preload("Artist").
		Order("created_at DESC").
		Limit(limit).
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("getting recent albums: %w", err)
	}
	return albums, nil
}

func (r *AlbumRepository) GetRandom(ctx context.Context, limit int) ([]models.Album, error) {
	var albums []models.Album
	err := r.db.WithContext(ctx).
		Preload("Artist").
		Order("RANDOM()").
		Limit(limit).
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("getting random albums: %w", err)
	}
	return albums, nil
}

func (r *AlbumRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Album{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting albums: %w", err)
	}
	return count, nil
}

func (r *AlbumRepository) GetByArtist(ctx context.Context, artistID string) ([]models.Album, error) {
	var albums []models.Album
	err := r.db.WithContext(ctx).
		Where("artist_id = ?", artistID).
		Order("year DESC, title ASC").
		Find(&albums).Error

	if err != nil {
		return nil, fmt.Errorf("getting albums by artist: %w", err)
	}
	return albums, nil
}
