package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lexi/books/models"
	"lexi/books/routes"
	"lexi/books/service"
	"lexi/books/test/testutil"

	"github.com/gin-gonic/gin"
)

func newSerieHandler(t *testing.T) (*routes.SerieHandler, *testutil.FakeSerieRepository) {
	t.Helper()

	serieRepo := testutil.NewFakeSerieRepository()
	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())
	return routes.NewSerieHandler(svc), serieRepo
}

func TestCreateSerie_Success(t *testing.T) {
	handler, serieRepo := newSerieHandler(t)

	router := gin.New()
	router.POST("/series", handler.CreateSerie)

	body, _ := json.Marshal(map[string]string{"serie": "Mistborn Era One"})
	req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(serieRepo.Series) != 1 {
		t.Fatalf("expected 1 serie to be persisted, got %d", len(serieRepo.Series))
	}
}

func TestCreateSerie_MissingNameReturns400(t *testing.T) {
	handler, _ := newSerieHandler(t)

	router := gin.New()
	router.POST("/series", handler.CreateSerie)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/series", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetSerieByID_NotFoundReturns404(t *testing.T) {
	handler, _ := newSerieHandler(t)

	router := gin.New()
	router.GET("/series/:id", handler.GetSerieByID)

	req := httptest.NewRequest(http.MethodGet, "/series/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetSeries_FiltersByUniverseID(t *testing.T) {
	handler, serieRepo := newSerieHandler(t)
	serieRepo.Create(models.Serie{Name: "Mistborn Era One", UniverseID: "cosmere"})
	serieRepo.Create(models.Serie{Name: "The Lord of the Rings", UniverseID: "middle-earth"})

	router := gin.New()
	router.GET("/series", handler.GetSeries)

	req := httptest.NewRequest(http.MethodGet, "/series?universeId=cosmere", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 || resp[0]["serie"] != "Mistborn Era One" {
		t.Fatalf("expected only Mistborn Era One, got %v", resp)
	}
}

func TestGetSeries_InvalidUniversePositionReturns400(t *testing.T) {
	handler, _ := newSerieHandler(t)

	router := gin.New()
	router.GET("/series", handler.GetSeries)

	req := httptest.NewRequest(http.MethodGet, "/series?universePosition=not-a-number", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateSerie_ChangesName(t *testing.T) {
	handler, serieRepo := newSerieHandler(t)
	created, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	router := gin.New()
	router.PUT("/series/:id", handler.UpdateSerie)

	body, _ := json.Marshal(map[string]string{"serie": "Mistborn: Era One"})
	req := httptest.NewRequest(http.MethodPut, "/series/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["serie"] != "Mistborn: Era One" {
		t.Fatalf("expected serie name to be updated, got %v", resp)
	}
}

func TestDeleteSerie_Success(t *testing.T) {
	handler, serieRepo := newSerieHandler(t)
	created, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	router := gin.New()
	router.DELETE("/series/:id", handler.DeleteSerie)

	req := httptest.NewRequest(http.MethodDelete, "/series/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
