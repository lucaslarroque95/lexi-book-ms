package schemas

type UniverseCreate struct {
	Name string `json:"universe" binding:"required"`
}

type UniverseUpdate struct {
	Name *string `json:"universe" binding:"omitempty"`
}

type UniverseRead struct {
	ID   string `json:"id"`
	Name string `json:"universe"`
}
