package db

import "time"

type Post struct {
	ID             string    `gorm:"primaryKey"`
	AuthorID       string    `gorm:"not null"`
	AuthorNickname string    `gorm:"not null"`
	AuthorPhotoURL string
	Body           string    `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type Like struct {
	PostID string `gorm:"primaryKey"`
	UserID string `gorm:"primaryKey"`
}
