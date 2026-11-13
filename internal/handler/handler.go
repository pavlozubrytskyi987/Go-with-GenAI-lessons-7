package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"homework/internal/model"
	"homework/internal/repository"
)

type BookHandler struct {
	repo *repository.BookRepository
}

func New(repo *repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, ok := parseBody(w, r)
	if !ok {
		return
	}

	if err := validateRequired(body); err != "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err)
		return
	}

	now := time.Now().UTC()
	b := model.Book{
		Title:     body["title"].(string),
		ISBN:      body["isbn"].(string),
		Author:    body["author"].(string),
		Category:  body["category"].(string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if v, ok := body["published_year"]; ok && v != nil {
		f, fOk := v.(float64)
		if !fOk || f != float64(int(f)) {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "published_year must be an integer")
			return
		}
		i := int(f)
		b.PublishedYear = &i
	}

	if v, ok := body["description"]; ok && v != nil {
		s, sOk := v.(string)
		if !sOk {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "description must be a string")
			return
		}
		b.Description = &s
	}

	created := h.repo.Create(b)
	writeJSON(w, http.StatusCreated, created)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	b, err := h.repo.FindByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Book not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	body, ok := parseBody(w, r)
	if !ok {
		return
	}

	if err := validateRequired(body); err != "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err)
		return
	}

	now := time.Now().UTC()
	b := model.Book{
		Title:     body["title"].(string),
		ISBN:      body["isbn"].(string),
		Author:    body["author"].(string),
		Category:  body["category"].(string),
		UpdatedAt: now,
	}

	if v, ok := body["published_year"]; ok && v != nil {
		f, fOk := v.(float64)
		if !fOk || f != float64(int(f)) {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "published_year must be an integer")
			return
		}
		i := int(f)
		b.PublishedYear = &i
	}

	if v, ok := body["description"]; ok && v != nil {
		s, sOk := v.(string)
		if !sOk {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "description must be a string")
			return
		}
		b.Description = &s
	}

	updated, err := h.repo.Update(id, b)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_server_error", "Internal error")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	err := h.repo.Delete(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_server_error", "Internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := 1
	limit := 10

	if v := q.Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			writeError(w, http.StatusBadRequest, "invalid_pagination", "page must be a positive integer")
			return
		}
		page = p
	}

	if v := q.Get("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err != nil || l < 1 {
			writeError(w, http.StatusBadRequest, "invalid_pagination", "limit must be a positive integer")
			return
		}
		limit = l
	}

	if limit > 100 {
		limit = 100
	}

	category := q.Get("category")
	books := h.repo.List(category, page, limit)

	writeJSON(w, http.StatusOK, books)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]

	if strings.Contains(idStr, ".") {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a valid integer")
		return 0, false
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "ID must be a valid integer")
		return 0, false
	}
	return int(id), true
}

func parseBody(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	var body map[string]any
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "Invalid JSON body")
		return nil, false
	}

	if body == nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "Request body must be a JSON object")
		return nil, false
	}
	return body, true
}

func validateRequired(body map[string]any) string {
	for _, field := range []string{"title", "isbn", "author", "category"} {
		v, ok := body[field]
		if !ok || v == nil {
			return field + " is required"
		}
		s, sOk := v.(string)
		if !sOk {
			return field + " must be a string"
		}
		if strings.TrimSpace(s) == "" {
			return field + " must not be empty"
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
