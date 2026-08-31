package models

type Serie struct {
	ID               string
	UserID           string
	Name             string
	UniverseID       string
	UniversePosition int
}

type SerieFilter struct {
	UniverseID          string
	UniversePosition    *int
	MaxUniversePosition *int
}
