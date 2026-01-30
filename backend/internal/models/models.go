package models

type User struct {
	ID           string `gorm:"primaryKey;type:varchar(36)"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
}

type Track struct {
	ID string `gorm:"primaryKey;type:varchar(36)"`
}

type Playlist struct {
	ID string `gorm:"primaryKey;type:varchar(36)"`
}
