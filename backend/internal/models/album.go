package models

import (
	"time"
)

type Album struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	Title        string    `gorm:"not null;index" json:"title"`
	Year         int       `gorm:"index" json:"year,omitempty"`
	CoverArtPath string    `gorm:"type:text" json:"-"`
	CoverArtURL  string    `gorm:"-" json:"coverArtUrl,omitempty"`
	ArtistID     string    `gorm:"index;type:text" json:"artistId"`
	Artist       *Artist   `gorm:"foreignKey:ArtistID" json:"artist,omitempty"`
	Tracks       []Track   `gorm:"foreignKey:AlbumID" json:"tracks,omitempty"`
	TrackCount   int       `gorm:"-" json:"trackCount,omitempty"`
	Duration     int       `gorm:"-" json:"duration,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (Album) TableName() string {
	return "albums"
}
