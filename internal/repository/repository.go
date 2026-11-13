package repository

import (
	"errors"
	"sort"
	"sync"

	"homework/internal/model"
)

var ErrNotFound = errors.New("not found")

type BookRepository struct {
	mu     sync.RWMutex
	books  map[int]model.Book
	nextID int
}

func New() *BookRepository {
	return &BookRepository{
		books:  make(map[int]model.Book),
		nextID: 1,
	}
}

func (r *BookRepository) Create(b model.Book) model.Book {
	r.mu.Lock()
	defer r.mu.Unlock()

	b.ID = r.nextID
	r.nextID++
	r.books[b.ID] = b
	return b
}

func (r *BookRepository) FindByID(id int) (model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}
	return b, nil
}

func (r *BookRepository) Update(id int, b model.Book) (model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}
	b.ID = id
	b.CreatedAt = existing.CreatedAt
	r.books[id] = b
	return b, nil
}

func (r *BookRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}

func (r *BookRepository) List(category string, page, limit int) []model.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []model.Book
	for _, b := range r.books {
		if category != "" && b.Category != category {
			continue
		}
		all = append(all, b)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	start := (page - 1) * limit
	if start >= len(all) {
		return []model.Book{}
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end]
}
