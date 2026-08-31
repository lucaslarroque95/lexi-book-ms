package db

import "time"

type Book struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID        string    `gorm:"type:uuid"`
	Name          string    `gorm:"type:text"`
	SerieID       *string   `gorm:"type:uuid"`
	SeriePosition *int      `gorm:"type:integer"`
	FileKey       string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
}

func (Book) TableName() string {
	return "books"
}

type Universe struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid"`
	Name      string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (Universe) TableName() string {
	return "universes"
}

type Serie struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID           string    `gorm:"type:uuid"`
	Name             string    `gorm:"type:text"`
	UniverseID       *string   `gorm:"type:uuid"`
	UniversePosition *int      `gorm:"type:integer"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

func (Serie) TableName() string {
	return "series"
}
