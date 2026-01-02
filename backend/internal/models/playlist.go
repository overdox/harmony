package models

import (
	"time"
)

type Playlist struct {
	ID             string          `gorm:"primaryKey;type:text" json:"id"`
	Name           string          `gorm:"not null;index;type:text" json:"name"`
	Description    string          `gorm:"type:text" json:"description,omitempty"`
	CoverImagePath string          `gorm:"type:text" json:"-"`
	CoverImageURL  string          `gorm:"-" json:"coverImageUrl,omitempty"`
	UserID         string          `gorm:"index;type:text" json:"userId"`
	User           *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IsPublic       bool            `gorm:"default:false" json:"isPublic"`
	PlaylistTracks []PlaylistTrack `gorm:"foreignKey:PlaylistID" json:"-"`
	Tracks         []Track         `gorm:"-" json:"tracks,omitempty"`
	TrackCount     int             `gorm:"-" json:"trackCount,omitempty"`
	Duration       int             `gorm:"-" json:"duration,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func (Playlist) TableName() string {
	return "playlists"
}

type PlaylistTrack struct {
	PlaylistID string    `gorm:"primaryKey;type:text" json:"playlistId"`
	TrackID    string    `gorm:"primaryKey;type:text" json:"trackId"`
	Position   int       `gorm:"not null;index" json:"position"`
	Track      *Track    `gorm:"foreignKey:TrackID" json:"track,omitempty"`
	AddedAt    time.Time `json:"addedAt"`
}

func (PlaylistTrack) TableName() string {
	return "playlist_tracks"
}
