package routes

import (
	"net/http"
	"strconv"

	"lexi/books/models"
	"lexi/books/schemas"
	"lexi/books/service"

	"github.com/gin-gonic/gin"
)

type SerieHandler struct {
	service *service.SerieService
}

func NewSerieHandler(service *service.SerieService) *SerieHandler {
	return &SerieHandler{service: service}
}

// GetSeries godoc
// @Summary      List series
// @Description  Lists series, optionally filtered by universe or position within the universe
// @Tags         series
// @Produce      json
// @Param        universeId           query     string  false  "Filter by universe ID"
// @Param        universePosition     query     int     false  "Filter by exact position in the universe"
// @Param        maxUniversePosition  query     int     false  "Filter by position <= this value"
// @Success      200  {array}   schemas.SerieRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /series [get]
func (h *SerieHandler) GetSeries(c *gin.Context) {
	filter, err := parseSerieFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	series, err := h.service.ListSeries(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch series"})
		return
	}

	c.JSON(http.StatusOK, toSerieReadList(series))
}

// GetSerieByID godoc
// @Summary      Get a serie by ID
// @Tags         series
// @Produce      json
// @Param        id  path      string  true  "Serie ID"
// @Success      200  {object}  schemas.SerieRead
// @Failure      404  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id} [get]
func (h *SerieHandler) GetSerieByID(c *gin.Context) {
	id := c.Param("id")

	serie, err := h.service.GetSerie(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "serie not found"})
		return
	}

	c.JSON(http.StatusOK, toSerieRead(serie))
}

// CreateSerie godoc
// @Summary      Create a serie
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        payload  body  schemas.SerieCreate  true  "Serie to create"
// @Success      201  {object}  schemas.SerieRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /series [post]
func (h *SerieHandler) CreateSerie(c *gin.Context) {
	var payload schemas.SerieCreate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serie := models.Serie{
		Name:             payload.Name,
		UserID:           c.GetString("userId"),
		UniverseID:       payload.UniverseID,
		UniversePosition: payload.UniversePosition,
	}

	created, err := h.service.CreateSerie(serie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create serie"})
		return
	}

	c.JSON(http.StatusCreated, toSerieRead(created))
}

// UpdateSerie godoc
// @Summary      Update a serie
// @Tags         series
// @Accept       json
// @Produce      json
// @Param        id       path  string               true  "Serie ID"
// @Param        payload  body  schemas.SerieUpdate  true  "Fields to update"
// @Success      200  {object}  schemas.SerieRead
// @Failure      400  {object}  schemas.ErrorResponse
// @Failure      404  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id} [put]
func (h *SerieHandler) UpdateSerie(c *gin.Context) {
	id := c.Param("id")

	var payload schemas.SerieUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetSerie(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "serie not found"})
		return
	}

	if payload.Name != nil {
		existing.Name = *payload.Name
	}

	updated, err := h.service.UpdateSerie(id, existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update serie"})
		return
	}

	c.JSON(http.StatusOK, toSerieRead(updated))
}

// DeleteSerie godoc
// @Summary      Delete a serie
// @Tags         series
// @Param        id  path  string  true  "Serie ID"
// @Success      204  "No Content"
// @Failure      500  {object}  schemas.ErrorResponse
// @Security     BearerAuth
// @Router       /series/{id} [delete]
func (h *SerieHandler) DeleteSerie(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteSerie(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete serie"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseSerieFilter(c *gin.Context) (models.SerieFilter, error) {
	filter := models.SerieFilter{
		UniverseID: c.Query("universeId"),
	}

	if raw := c.Query("universePosition"); raw != "" {
		position, err := strconv.Atoi(raw)
		if err != nil {
			return models.SerieFilter{}, err
		}
		filter.UniversePosition = &position
	}

	if raw := c.Query("maxUniversePosition"); raw != "" {
		position, err := strconv.Atoi(raw)
		if err != nil {
			return models.SerieFilter{}, err
		}
		filter.MaxUniversePosition = &position
	}

	return filter, nil
}

func toSerieRead(serie models.Serie) schemas.SerieRead {
	return schemas.SerieRead{
		ID:               serie.ID,
		Name:             serie.Name,
		UniverseID:       serie.UniverseID,
		UniversePosition: serie.UniversePosition,
	}
}

func toSerieReadList(series []models.Serie) []schemas.SerieRead {
	result := make([]schemas.SerieRead, len(series))
	for i, serie := range series {
		result[i] = toSerieRead(serie)
	}
	return result
}
