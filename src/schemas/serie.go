package schemas

type SerieCreate struct {
	Name             string `json:"serie" binding:"required"`
	UniverseID       string `json:"universeId"`
	UniversePosition int    `json:"universePosition"`
}

type SerieUpdate struct {
	Name *string `json:"serie" binding:"omitempty"`
}

type SerieRead struct {
	ID               string `json:"id"`
	Name             string `json:"serie"`
	UniverseID       string `json:"universeId"`
	UniversePosition int    `json:"universePosition"`
}
