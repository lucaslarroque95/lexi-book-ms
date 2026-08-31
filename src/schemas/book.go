package schemas

import "time"

type BookCreate struct {
	Name          string `json:"book" binding:"required"`
	SerieID       string `json:"serieId"`
	SeriePosition int    `json:"seriePosition"`
}

type BookUpdate struct {
	Name    *string `json:"book" binding:"omitempty"`
	FileKey *string `json:"fileKey" binding:"omitempty"`
}

type BookRead struct {
	ID            string `json:"id"`
	Name          string `json:"book"`
	SerieID       string `json:"serieId"`
	SeriePosition int    `json:"seriePosition"`
	FileKey       string `json:"fileKey"`
}

type BookDownloadURL struct {
	DownloadURL string    `json:"downloadUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
