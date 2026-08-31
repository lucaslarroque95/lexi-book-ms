package testutil

import (
	"errors"
	"strconv"

	"lexi/books/models"
	"lexi/books/repositories"
)

var (
	_ repositories.BookRepository     = (*FakeBookRepository)(nil)
	_ repositories.UniverseRepository = (*FakeUniverseRepository)(nil)
	_ repositories.SerieRepository    = (*FakeSerieRepository)(nil)
)

// FakeBookRepository is an in-memory repositories.BookRepository for tests, no database required.
type FakeBookRepository struct {
	Books  map[string]models.Book
	NextID int

	CreateErr     error
	GetErr        error
	GetByNameErr  error
	GetAllErr     error
	UpdateErr     error
	DeleteErr     error
	ClearSerieErr error

	CreateCalls int
}

func NewFakeBookRepository() *FakeBookRepository {
	return &FakeBookRepository{Books: make(map[string]models.Book)}
}

func (f *FakeBookRepository) Create(book models.Book) (models.Book, error) {
	f.CreateCalls++
	if f.CreateErr != nil {
		return models.Book{}, f.CreateErr
	}

	if book.ID == "" {
		f.NextID++
		book.ID = strconv.Itoa(f.NextID)
	}
	f.Books[book.ID] = book
	return book, nil
}

func (f *FakeBookRepository) Get(id string) (models.Book, error) {
	if f.GetErr != nil {
		return models.Book{}, f.GetErr
	}

	book, ok := f.Books[id]
	if !ok {
		return models.Book{}, errors.New("book not found")
	}
	return book, nil
}

func (f *FakeBookRepository) GetByName(name string) (models.Book, error) {
	if f.GetByNameErr != nil {
		return models.Book{}, f.GetByNameErr
	}

	for _, book := range f.Books {
		if book.Name == name {
			return book, nil
		}
	}
	return models.Book{}, errors.New("book not found")
}

func (f *FakeBookRepository) GetAll(filter models.BookFilter) ([]models.Book, error) {
	if f.GetAllErr != nil {
		return nil, f.GetAllErr
	}

	// filter.UniverseID is not applied here: a book has no universe of its
	// own (only its serie does), so the postgres repository resolves it via
	// a join this in-memory fake can't emulate.
	var result []models.Book
	for _, book := range f.Books {
		if filter.SerieID != "" && book.SerieID != filter.SerieID {
			continue
		}
		if filter.BookPosition != nil && book.SeriePosition != *filter.BookPosition {
			continue
		}
		if filter.MaxBookPosition != nil && book.SeriePosition > *filter.MaxBookPosition {
			continue
		}
		result = append(result, book)
	}
	return result, nil
}

func (f *FakeBookRepository) Update(id string, book models.Book) (models.Book, error) {
	if f.UpdateErr != nil {
		return models.Book{}, f.UpdateErr
	}
	if _, ok := f.Books[id]; !ok {
		return models.Book{}, errors.New("book not found")
	}

	book.ID = id
	f.Books[id] = book
	return book, nil
}

func (f *FakeBookRepository) Delete(id string) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	delete(f.Books, id)
	return nil
}

func (f *FakeBookRepository) ClearSerie(serieID string) error {
	if f.ClearSerieErr != nil {
		return f.ClearSerieErr
	}
	for id, book := range f.Books {
		if book.SerieID == serieID {
			book.SerieID = ""
			book.SeriePosition = 0
			f.Books[id] = book
		}
	}
	return nil
}

// FakeUniverseRepository is an in-memory repositories.UniverseRepository for tests, no database required.
type FakeUniverseRepository struct {
	Universes map[string]models.Universe
	NextID    int

	CreateErr error
	GetErr    error
	GetAllErr error
	UpdateErr error
	DeleteErr error

	CreateCalls int
}

func NewFakeUniverseRepository() *FakeUniverseRepository {
	return &FakeUniverseRepository{Universes: make(map[string]models.Universe)}
}

func (f *FakeUniverseRepository) Create(universe models.Universe) (models.Universe, error) {
	f.CreateCalls++
	if f.CreateErr != nil {
		return models.Universe{}, f.CreateErr
	}

	if universe.ID == "" {
		f.NextID++
		universe.ID = strconv.Itoa(f.NextID)
	}
	f.Universes[universe.ID] = universe
	return universe, nil
}

func (f *FakeUniverseRepository) Get(id string) (models.Universe, error) {
	if f.GetErr != nil {
		return models.Universe{}, f.GetErr
	}

	universe, ok := f.Universes[id]
	if !ok {
		return models.Universe{}, errors.New("universe not found")
	}
	return universe, nil
}

func (f *FakeUniverseRepository) GetAll() ([]models.Universe, error) {
	if f.GetAllErr != nil {
		return nil, f.GetAllErr
	}

	result := make([]models.Universe, 0, len(f.Universes))
	for _, universe := range f.Universes {
		result = append(result, universe)
	}
	return result, nil
}

func (f *FakeUniverseRepository) Update(id string, universe models.Universe) (models.Universe, error) {
	if f.UpdateErr != nil {
		return models.Universe{}, f.UpdateErr
	}
	if _, ok := f.Universes[id]; !ok {
		return models.Universe{}, errors.New("universe not found")
	}

	universe.ID = id
	f.Universes[id] = universe
	return universe, nil
}

func (f *FakeUniverseRepository) Delete(id string) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	delete(f.Universes, id)
	return nil
}

// FakeSerieRepository is an in-memory repositories.SerieRepository for tests, no database required.
type FakeSerieRepository struct {
	Series map[string]models.Serie
	NextID int

	CreateErr        error
	GetErr           error
	GetAllErr        error
	UpdateErr        error
	DeleteErr        error
	ClearUniverseErr error

	CreateCalls int
}

func NewFakeSerieRepository() *FakeSerieRepository {
	return &FakeSerieRepository{Series: make(map[string]models.Serie)}
}

func (f *FakeSerieRepository) Create(serie models.Serie) (models.Serie, error) {
	f.CreateCalls++
	if f.CreateErr != nil {
		return models.Serie{}, f.CreateErr
	}

	if serie.ID == "" {
		f.NextID++
		serie.ID = strconv.Itoa(f.NextID)
	}
	f.Series[serie.ID] = serie
	return serie, nil
}

func (f *FakeSerieRepository) Get(id string) (models.Serie, error) {
	if f.GetErr != nil {
		return models.Serie{}, f.GetErr
	}

	serie, ok := f.Series[id]
	if !ok {
		return models.Serie{}, errors.New("serie not found")
	}
	return serie, nil
}

func (f *FakeSerieRepository) GetAll(filter models.SerieFilter) ([]models.Serie, error) {
	if f.GetAllErr != nil {
		return nil, f.GetAllErr
	}

	var result []models.Serie
	for _, serie := range f.Series {
		if filter.UniverseID != "" && serie.UniverseID != filter.UniverseID {
			continue
		}
		if filter.UniversePosition != nil && serie.UniversePosition != *filter.UniversePosition {
			continue
		}
		if filter.MaxUniversePosition != nil && serie.UniversePosition > *filter.MaxUniversePosition {
			continue
		}
		result = append(result, serie)
	}
	return result, nil
}

func (f *FakeSerieRepository) Update(id string, serie models.Serie) (models.Serie, error) {
	if f.UpdateErr != nil {
		return models.Serie{}, f.UpdateErr
	}
	if _, ok := f.Series[id]; !ok {
		return models.Serie{}, errors.New("serie not found")
	}

	serie.ID = id
	f.Series[id] = serie
	return serie, nil
}

func (f *FakeSerieRepository) Delete(id string) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	delete(f.Series, id)
	return nil
}

func (f *FakeSerieRepository) ClearUniverse(universeID string) error {
	if f.ClearUniverseErr != nil {
		return f.ClearUniverseErr
	}
	for id, serie := range f.Series {
		if serie.UniverseID == universeID {
			serie.UniverseID = ""
			serie.UniversePosition = 0
			f.Series[id] = serie
		}
	}
	return nil
}
