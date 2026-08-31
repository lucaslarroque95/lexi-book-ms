package routes

import (
	"net/http"

	"lexi/books/models"
	"lexi/books/schemas"
	"lexi/books/service"

	"github.com/gin-gonic/gin"
)

type UniverseHandler struct {
	service *service.UniverseService
}

func NewUniverseHandler(service *service.UniverseService) *UniverseHandler {
	return &UniverseHandler{service: service}
}

// GetUniverses godoc
// @Summary      List universes
// @Tags         universes
// @Produce      json
// @Success      200  {array}   schemas.UniverseRead
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /universes [get]
func (h *UniverseHandler) GetUniverses(c *gin.Context) {
	universes, err := h.service.ListUniverses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch universes"})
		return
	}

	c.JSON(http.StatusOK, toUniverseReadList(universes))
}

// GetUniverseByID godoc
// @Summary      Get a universe by ID
// @Tags         universes
// @Produce      json
// @Param        id  path      string  true  "Universe ID"
// @Success      200  {object}  schemas.UniverseRead
// @Failure      404  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /universes/{id} [get]
func (h *UniverseHandler) GetUniverseByID(c *gin.Context) {
	id := c.Param("id")

	universe, err := h.service.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "universe not found"})
		return
	}

	c.JSON(http.StatusOK, toUniverseRead(universe))
}

// CreateUniverse godoc
// @Summary      Create a universe
// @Tags         universes
// @Accept       json
// @Produce      json
// @Param        payload  body  schemas.UniverseCreate  true  "Universe to create"
// @Success      201  {object}  schemas.UniverseRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /universes [post]
func (h *UniverseHandler) CreateUniverse(c *gin.Context) {
	var payload schemas.UniverseCreate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	universe := models.Universe{
		Name:   payload.Name,
		UserID: c.GetString("userId"),
	}

	created, err := h.service.CreateUniverse(universe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create universe"})
		return
	}

	c.JSON(http.StatusCreated, toUniverseRead(created))
}

// UpdateUniverse godoc
// @Summary      Update a universe
// @Tags         universes
// @Accept       json
// @Produce      json
// @Param        id       path  string                  true  "Universe ID"
// @Param        payload  body  schemas.UniverseUpdate  true  "Fields to update"
// @Success      200  {object}  schemas.UniverseRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      404  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /universes/{id} [put]
func (h *UniverseHandler) UpdateUniverse(c *gin.Context) {
	id := c.Param("id")

	var payload schemas.UniverseUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "universe not found"})
		return
	}

	if payload.Name != nil {
		existing.Name = *payload.Name
	}

	updated, err := h.service.UpdateUniverse(id, existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update universe"})
		return
	}

	c.JSON(http.StatusOK, toUniverseRead(updated))
}

// DeleteUniverse godoc
// @Summary      Delete a universe
// @Tags         universes
// @Param        id  path  string  true  "Universe ID"
// @Success      204  "No Content"
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /universes/{id} [delete]
func (h *UniverseHandler) DeleteUniverse(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteUniverse(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete universe"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toUniverseRead(universe models.Universe) schemas.UniverseRead {
	return schemas.UniverseRead{ID: universe.ID, Name: universe.Name}
}

func toUniverseReadList(universes []models.Universe) []schemas.UniverseRead {
	result := make([]schemas.UniverseRead, len(universes))
	for i, universe := range universes {
		result[i] = toUniverseRead(universe)
	}
	return result
}
