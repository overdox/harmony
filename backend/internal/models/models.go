package models

// AllModels returns all models for GORM auto-migration
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Artist{},
		&Album{},
		&Track{},
		&Playlist{},
		&PlaylistTrack{},
		&Settings{},
	}
}
