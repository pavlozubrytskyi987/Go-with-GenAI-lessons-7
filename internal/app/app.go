package app

import (
	"net/http"
	"strings"

	"homework/internal/handler"
	"homework/internal/middleware"
	"homework/internal/repository"
)

func NewRouter() http.Handler {
	repo := repository.New()
	h := handler.New(repo)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	mux.HandleFunc("/api/v1/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		case http.MethodPost:
			h.Create(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":{"code":"method_not_allowed","message":"Method not allowed"}}`))
		}
	})

	mux.HandleFunc("/api/v1/books/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/books/")
		if path == "" || strings.Contains(path, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.Get(w, r)
		case http.MethodPut:
			h.Update(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":{"code":"method_not_allowed","message":"Method not allowed"}}`))
		}
	})

	var chain http.Handler = mux
	chain = middleware.Recovery(chain)
	chain = middleware.Logging(chain)
	chain = middleware.CORS(chain)

	return chain
}
