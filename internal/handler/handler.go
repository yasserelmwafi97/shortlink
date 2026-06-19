package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Encoder interface {
	Encode(rawURL string) (string, error)
	Decode(shortURL string) (string, error)
}

type Handler struct {
	svc          Encoder
	maxBodyBytes int64
}

func New(svc Encoder, maxBodyBytes int64, rateLimitPerMin int) http.Handler {
	h := &Handler{svc: svc, maxBodyBytes: maxBodyBytes}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(h.limitBody)

	if rateLimitPerMin > 0 {
		r.Use(newRateLimiter(rateLimitPerMin).middleware)
	}

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/encode", h.encode)
	r.Post("/decode", h.decode)

	return r
}

func (h *Handler) limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}
