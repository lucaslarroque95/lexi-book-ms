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

func newUniverseHandler(t *testing.T) (*routes.UniverseHandler, *testutil.FakeUniverseRepository) {
	t.Helper()

	universeRepo := testutil.NewFakeUniverseRepository()
	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())
	return routes.NewUniverseHandler(svc), universeRepo
}

func TestCreateUniverse_Success(t *testing.T) {
	handler, universeRepo := newUniverseHandler(t)

	router := gin.New()
	router.POST("/universes", handler.CreateUniverse)

	body, _ := json.Marshal(map[string]string{"universe": "Cosmere"})
	req := httptest.NewRequest(http.MethodPost, "/universes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(universeRepo.Universes) != 1 {
		t.Fatalf("expected 1 universe to be persisted, got %d", len(universeRepo.Universes))
	}
}

func TestCreateUniverse_MissingNameReturns400(t *testing.T) {
	handler, _ := newUniverseHandler(t)

	router := gin.New()
	router.POST("/universes", handler.CreateUniverse)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/universes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetUniverseByID_NotFoundReturns404(t *testing.T) {
	handler, _ := newUniverseHandler(t)

	router := gin.New()
	router.GET("/universes/:id", handler.GetUniverseByID)

	req := httptest.NewRequest(http.MethodGet, "/universes/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetUniverses_ReturnsAll(t *testing.T) {
	handler, universeRepo := newUniverseHandler(t)
	universeRepo.Create(models.Universe{Name: "Cosmere"})
	universeRepo.Create(models.Universe{Name: "Middle-earth"})

	router := gin.New()
	router.GET("/universes", handler.GetUniverses)

	req := httptest.NewRequest(http.MethodGet, "/universes", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 universes, got %d", len(resp))
	}
}

func TestUpdateUniverse_ChangesName(t *testing.T) {
	handler, universeRepo := newUniverseHandler(t)
	created, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	router := gin.New()
	router.PUT("/universes/:id", handler.UpdateUniverse)

	body, _ := json.Marshal(map[string]string{"universe": "The Cosmere"})
	req := httptest.NewRequest(http.MethodPut, "/universes/"+created.ID, bytes.NewReader(body))
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
	if resp["universe"] != "The Cosmere" {
		t.Fatalf("expected universe name to be updated, got %v", resp)
	}
}

func TestDeleteUniverse_Success(t *testing.T) {
	handler, universeRepo := newUniverseHandler(t)
	created, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	router := gin.New()
	router.DELETE("/universes/:id", handler.DeleteUniverse)

	req := httptest.NewRequest(http.MethodDelete, "/universes/"+created.ID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
