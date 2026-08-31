package service_test

import (
	"errors"
	"testing"

	"lexi/books/models"
	"lexi/books/service"
	"lexi/books/test/testutil"
)

func TestCreateSerie_ReturnsSerieWithID(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	created, err := svc.CreateSerie(models.Serie{Name: "Mistborn Era One"})
	if err != nil {
		t.Fatalf("CreateSerie returned error: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected created serie to have an ID")
	}
}

func TestGetSerie_NotFound(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	if _, err := svc.GetSerie("missing"); err == nil {
		t.Fatalf("expected an error for a missing serie")
	}
}

func TestListSeries_ReturnsAllSeries(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	serieRepo.Create(models.Serie{Name: "Mistborn Era One"})
	serieRepo.Create(models.Serie{Name: "Mistborn Era Two"})

	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	series, err := svc.ListSeries(models.SerieFilter{})
	if err != nil {
		t.Fatalf("ListSeries returned error: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}
}

func TestListSeries_FiltersByUniverseID(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	serieRepo.Create(models.Serie{Name: "Mistborn Era One", UniverseID: "cosmere"})
	serieRepo.Create(models.Serie{Name: "The Stormlight Archive", UniverseID: "cosmere"})
	serieRepo.Create(models.Serie{Name: "The Lord of the Rings", UniverseID: "middle-earth"})

	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	series, err := svc.ListSeries(models.SerieFilter{UniverseID: "middle-earth"})
	if err != nil {
		t.Fatalf("ListSeries returned error: %v", err)
	}
	if len(series) != 1 || series[0].Name != "The Lord of the Rings" {
		t.Fatalf("expected only The Lord of the Rings, got %v", series)
	}
}

func TestListSeries_FiltersByMaxUniversePosition(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	serieRepo.Create(models.Serie{Name: "Serie 1", UniversePosition: 1})
	serieRepo.Create(models.Serie{Name: "Serie 2", UniversePosition: 2})
	serieRepo.Create(models.Serie{Name: "Serie 3", UniversePosition: 3})

	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	maxPosition := 2
	series, err := svc.ListSeries(models.SerieFilter{MaxUniversePosition: &maxPosition})
	if err != nil {
		t.Fatalf("ListSeries returned error: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 series with position <= 2, got %d", len(series))
	}
}

func TestUpdateSerie_ChangesName(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	created, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	updated, err := svc.UpdateSerie(created.ID, models.Serie{Name: "Mistborn: Era One"})
	if err != nil {
		t.Fatalf("UpdateSerie returned error: %v", err)
	}
	if updated.Name != "Mistborn: Era One" {
		t.Fatalf("expected serie name to be updated, got %q", updated.Name)
	}
}

func TestDeleteSerie_RemovesSerie(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	created, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	svc := service.NewSerieService(serieRepo, testutil.NewFakeBookRepository())

	if err := svc.DeleteSerie(created.ID); err != nil {
		t.Fatalf("DeleteSerie returned error: %v", err)
	}
	if _, err := svc.GetSerie(created.ID); err == nil {
		t.Fatalf("expected the serie to be deleted")
	}
}

func TestDeleteSerie_ClearsChildBooks(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	serie, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	bookRepo := testutil.NewFakeBookRepository()
	linkedBook, _ := bookRepo.Create(models.Book{Name: "The Final Empire", SerieID: serie.ID, SeriePosition: 1})
	unrelatedBook, _ := bookRepo.Create(models.Book{Name: "Elantris", SerieID: "other-serie", SeriePosition: 1})

	svc := service.NewSerieService(serieRepo, bookRepo)

	if err := svc.DeleteSerie(serie.ID); err != nil {
		t.Fatalf("DeleteSerie returned error: %v", err)
	}

	cleared := bookRepo.Books[linkedBook.ID]
	if cleared.SerieID != "" || cleared.SeriePosition != 0 {
		t.Fatalf("expected the linked book's serieId/seriePosition to be cleared, got %+v", cleared)
	}

	untouched := bookRepo.Books[unrelatedBook.ID]
	if untouched.SerieID != "other-serie" || untouched.SeriePosition != 1 {
		t.Fatalf("expected the unrelated book to be untouched, got %+v", untouched)
	}
}

func TestDeleteSerie_ClearFailurePropagatesAndSerieSurvives(t *testing.T) {
	serieRepo := testutil.NewFakeSerieRepository()
	serie, _ := serieRepo.Create(models.Serie{Name: "Mistborn Era One"})

	bookRepo := testutil.NewFakeBookRepository()
	bookRepo.ClearSerieErr = errors.New("database unreachable")

	svc := service.NewSerieService(serieRepo, bookRepo)

	if err := svc.DeleteSerie(serie.ID); err == nil {
		t.Fatalf("expected the clear error to propagate")
	}
	if _, err := svc.GetSerie(serie.ID); err != nil {
		t.Fatalf("expected the serie to still exist when clearing its books fails, got %v", err)
	}
}
