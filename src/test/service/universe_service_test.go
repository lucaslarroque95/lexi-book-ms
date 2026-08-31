package service_test

import (
	"errors"
	"testing"

	"lexi/books/models"
	"lexi/books/service"
	"lexi/books/test/testutil"
)

func TestCreateUniverse_ReturnsUniverseWithID(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())

	created, err := svc.CreateUniverse(models.Universe{Name: "Cosmere"})
	if err != nil {
		t.Fatalf("CreateUniverse returned error: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected created universe to have an ID")
	}
}

func TestGetUniverse_NotFound(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())

	if _, err := svc.GetUniverse("missing"); err == nil {
		t.Fatalf("expected an error for a missing universe")
	}
}

func TestListUniverses_ReturnsAllUniverses(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	universeRepo.Create(models.Universe{Name: "Cosmere"})
	universeRepo.Create(models.Universe{Name: "Middle-earth"})

	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())

	universes, err := svc.ListUniverses()
	if err != nil {
		t.Fatalf("ListUniverses returned error: %v", err)
	}
	if len(universes) != 2 {
		t.Fatalf("expected 2 universes, got %d", len(universes))
	}
}

func TestUpdateUniverse_ChangesName(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	created, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())

	updated, err := svc.UpdateUniverse(created.ID, models.Universe{Name: "The Cosmere"})
	if err != nil {
		t.Fatalf("UpdateUniverse returned error: %v", err)
	}
	if updated.Name != "The Cosmere" {
		t.Fatalf("expected universe name to be updated, got %q", updated.Name)
	}
}

func TestDeleteUniverse_RemovesUniverse(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	created, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	svc := service.NewUniverseService(universeRepo, testutil.NewFakeSerieRepository())

	if err := svc.DeleteUniverse(created.ID); err != nil {
		t.Fatalf("DeleteUniverse returned error: %v", err)
	}
	if _, err := svc.GetUniverse(created.ID); err == nil {
		t.Fatalf("expected the universe to be deleted")
	}
}

func TestDeleteUniverse_ClearsChildSeries(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	universe, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	serieRepo := testutil.NewFakeSerieRepository()
	linkedSerie, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One", UniverseID: universe.ID, UniversePosition: 1})
	unrelatedSerie, _ := serieRepo.Create(models.Serie{Name: "The Lord of the Rings", UniverseID: "middle-earth", UniversePosition: 1})

	svc := service.NewUniverseService(universeRepo, serieRepo)

	if err := svc.DeleteUniverse(universe.ID); err != nil {
		t.Fatalf("DeleteUniverse returned error: %v", err)
	}

	cleared := serieRepo.Series[linkedSerie.ID]
	if cleared.UniverseID != "" || cleared.UniversePosition != 0 {
		t.Fatalf("expected the linked serie's universeId/universePosition to be cleared, got %+v", cleared)
	}

	untouched := serieRepo.Series[unrelatedSerie.ID]
	if untouched.UniverseID != "middle-earth" || untouched.UniversePosition != 1 {
		t.Fatalf("expected the unrelated serie to be untouched, got %+v", untouched)
	}
}

func TestDeleteUniverse_ClearFailurePropagatesAndUniverseSurvives(t *testing.T) {
	universeRepo := testutil.NewFakeUniverseRepository()
	universe, _ := universeRepo.Create(models.Universe{Name: "Cosmere"})

	serieRepo := testutil.NewFakeSerieRepository()
	serieRepo.ClearUniverseErr = errors.New("database unreachable")

	svc := service.NewUniverseService(universeRepo, serieRepo)

	if err := svc.DeleteUniverse(universe.ID); err == nil {
		t.Fatalf("expected the clear error to propagate")
	}
	if _, err := svc.GetUniverse(universe.ID); err != nil {
		t.Fatalf("expected the universe to still exist when clearing its series fails, got %v", err)
	}
}
