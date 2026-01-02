package models

import (
	"time"
)

type Artist struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	Name      string    `gorm:"not null;index" json:"name"`
	Bio       string    `gorm:"type:text" json:"bio,omitempty"`
	ImagePath string    `gorm:"type:text" json:"-"`
	ImageURL  string    `gorm:"-" json:"imageUrl,omitempty"`
	Albums    []Album   `gorm:"foreignKey:ArtistID" json:"albums,omitempty"`
	Tracks    []Track   `gorm:"foreignKey:ArtistID" json:"tracks,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Artist) TableName() string {
	return "artists"
}
