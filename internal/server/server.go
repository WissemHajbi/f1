package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"oidysts/internal/store"
)

type Server struct {
	store  *store.Store
	logger *slog.Logger
}

func New(db *store.Store, logger *slog.Logger) http.Handler {
	s := &Server{store: db, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.health)
	mux.HandleFunc("GET /v1/sources", s.sources)
	return requestLog(logger, mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) sources(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.Latest(r.Context())
	if err != nil {
		s.logger.Error("list source snapshots", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if items == nil {
		items = []store.Snapshot{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func requestLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}
