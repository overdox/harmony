package models

import (
	"time"
)

type User struct {
	ID           string     `gorm:"primaryKey;type:text" json:"id"`
	Username     string     `gorm:"not null;uniqueIndex;type:text" json:"username"`
	Email        string     `gorm:"not null;uniqueIndex;type:text" json:"email"`
	PasswordHash string     `gorm:"not null;type:text" json:"-"`
	Playlists    []Playlist `gorm:"foreignKey:UserID" json:"playlists,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
