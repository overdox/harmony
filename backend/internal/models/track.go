package models

import (
	"time"
)

type Track struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	Title       string    `gorm:"not null;index" json:"title"`
	Duration    int       `gorm:"not null" json:"duration"`
	TrackNumber int       `gorm:"default:0" json:"trackNumber"`
	DiscNumber  int       `gorm:"default:1" json:"discNumber"`
	FilePath    string    `gorm:"not null;uniqueIndex;type:text" json:"-"`
	FileSize    int64     `gorm:"not null" json:"fileSize"`
	Format      string    `gorm:"not null;type:text" json:"format"`
	Bitrate     int       `gorm:"default:0" json:"bitrate,omitempty"`
	SampleRate  int       `gorm:"default:0" json:"sampleRate,omitempty"`
	Channels    int       `gorm:"default:2" json:"channels,omitempty"`
	AlbumID     string    `gorm:"index;type:text" json:"albumId,omitempty"`
	Album       *Album    `gorm:"foreignKey:AlbumID" json:"album,omitempty"`
	ArtistID    string    `gorm:"index;type:text" json:"artistId,omitempty"`
	Artist      *Artist   `gorm:"foreignKey:ArtistID" json:"artist,omitempty"`
	Genre       string    `gorm:"index;type:text" json:"genre,omitempty"`
	Year        int       `gorm:"index" json:"year,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Track) TableName() string {
	return "tracks"
}
