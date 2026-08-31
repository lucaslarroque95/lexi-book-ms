package models

type Book struct {
	ID            string
	UserID        string
	Name          string
	SerieID       string
	SeriePosition int
	FileKey       string
}

type BookFilter struct {
	SerieID         string
	UniverseID      string
	BookPosition    *int
	MaxBookPosition *int
}
