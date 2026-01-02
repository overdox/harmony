package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"harmony/internal/models"
)

var (
	ErrArtistNotFound = errors.New("artist not found")
)

type ArtistRepository struct {
	db *gorm.DB
}

func NewArtistRepository(db *gorm.DB) *ArtistRepository {
	return &ArtistRepository{db: db}
}

type ArtistFilter struct {
	Query string
}

type ArtistListOptions struct {
	Filter ArtistFilter
	Page   int
	Limit  int
	SortBy string
	Order  string
}

func (r *ArtistRepository) Create(ctx context.Context, artist *models.Artist) error {
	if err := r.db.WithContext(ctx).Create(artist).Error; err != nil {
		return fmt.Errorf("creating artist: %w", err)
	}
	return nil
}

func (r *ArtistRepository) FindByID(ctx context.Context, id string) (*models.Artist, error) {
	var artist models.Artist
	result := r.db.WithContext(ctx).First(&artist, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrArtistNotFound
		}
		return nil, fmt.Errorf("finding artist: %w", result.Error)
	}
	return &artist, nil
}

func (r *ArtistRepository) FindByIDWithAlbums(ctx context.Context, id string) (*models.Artist, error) {
	var artist models.Artist
	result := r.db.WithContext(ctx).
		Preload("Albums", func(db *gorm.DB) *gorm.DB {
			return db.Order("year DESC, title ASC")
		}).
		First(&artist, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrArtistNotFound
		}
		return nil, fmt.Errorf("finding artist with albums: %w", result.Error)
	}
	return &artist, nil
}

func (r *ArtistRepository) FindByName(ctx context.Context, name string) (*models.Artist, error) {
	var artist models.Artist
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&artist)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrArtistNotFound
		}
		return nil, fmt.Errorf("finding artist by name: %w", result.Error)
	}
	return &artist, nil
}

func (r *ArtistRepository) FindOrCreate(ctx context.Context, name string) (*models.Artist, error) {
	artist, err := r.FindByName(ctx, name)
	if err == nil {
		return artist, nil
	}
	if !errors.Is(err, ErrArtistNotFound) {
		return nil, err
	}

	// Create new artist
	newArtist := &models.Artist{
		ID:   GenerateID(),
		Name: name,
	}
	if err := r.Create(ctx, newArtist); err != nil {
		return nil, err
	}
	return newArtist, nil
}

func (r *ArtistRepository) List(ctx context.Context, opts ArtistListOptions) ([]models.Artist, int64, error) {
	var artists []models.Artist
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Artist{})

	// Apply filters
	if opts.Filter.Query != "" {
		searchQuery := "%" + opts.Filter.Query + "%"
		query = query.Where("name LIKE ?", searchQuery)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting artists: %w", err)
	}

	// Apply sorting - map frontend field names to database columns
	sortBy := "name"
	if opts.SortBy != "" {
		// Map common field names to actual column names
		sortMapping := map[string]string{
			"name":      "name",
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		}
		if mapped, ok := sortMapping[opts.SortBy]; ok {
			sortBy = mapped
		}
		// If not in mapping, ignore invalid sort field for security
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

	if err := query.Find(&artists).Error; err != nil {
		return nil, 0, fmt.Errorf("listing artists: %w", err)
	}

	return artists, total, nil
}

func (r *ArtistRepository) Search(ctx context.Context, query string, limit int) ([]models.Artist, error) {
	var artists []models.Artist
	searchQuery := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("name LIKE ?", searchQuery).
		Limit(limit).
		Find(&artists).Error

	if err != nil {
		return nil, fmt.Errorf("searching artists: %w", err)
	}
	return artists, nil
}

func (r *ArtistRepository) Update(ctx context.Context, artist *models.Artist) error {
	if err := r.db.WithContext(ctx).Save(artist).Error; err != nil {
		return fmt.Errorf("updating artist: %w", err)
	}
	return nil
}

func (r *ArtistRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Artist{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("deleting artist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrArtistNotFound
	}
	return nil
}

func (r *ArtistRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Artist{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting artists: %w", err)
	}
	return count, nil
}

func (r *ArtistRepository) GetPopularTracks(ctx context.Context, artistID string, limit int) ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.WithContext(ctx).
		Preload("Album").
		Where("artist_id = ?", artistID).
		Limit(limit).
		Find(&tracks).Error

	if err != nil {
		return nil, fmt.Errorf("getting popular tracks: %w", err)
	}
	return tracks, nil
}

// DeleteEmpty deletes artists that have no albums
func (r *ArtistRepository) DeleteEmpty(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM artists
		WHERE id NOT IN (SELECT DISTINCT artist_id FROM albums WHERE artist_id IS NOT NULL)
	`)
	if result.Error != nil {
		return 0, fmt.Errorf("deleting empty artists: %w", result.Error)
	}
	return result.RowsAffected, nil
}

