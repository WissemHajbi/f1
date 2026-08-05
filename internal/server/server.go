package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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
	mux.HandleFunc("GET /v1/drivers", s.drivers)
	mux.HandleFunc("GET /v1/calendar", s.calendar)
	mux.HandleFunc("GET /v1/calendar/next", s.nextEvent)
	mux.HandleFunc("GET /v1/standings/drivers", s.driverStandings)
	mux.HandleFunc("GET /v1/standings/constructors", s.constructorStandings)
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

func (s *Server) drivers(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	roster, err := s.store.LatestDriverRoster(r.Context(), year)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "drivers not synced for requested season"})
		return
	}
	if err != nil {
		s.logger.Error("list drivers", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": roster.Drivers, "meta": map[string]any{
		"season": roster.Session.Year, "session": roster.Session, "synced_at": roster.SyncedAt,
	}})
}

func (s *Server) driverStandings(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	items, err := s.store.DriverStandings(r.Context(), year)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "driver standings not synced for requested season"})
		return
	}
	if err != nil {
		s.logger.Error("list driver standings", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{
		"season": year, "round": items[0].Round, "count": len(items), "synced_at": items[0].SyncedAt,
	}})
}

func (s *Server) constructorStandings(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	items, err := s.store.ConstructorStandings(r.Context(), year)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "constructor standings not synced for requested season"})
		return
	}
	if err != nil {
		s.logger.Error("list constructor standings", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{
		"season": year, "round": items[0].Round, "count": len(items), "synced_at": items[0].SyncedAt,
	}})
}

func (s *Server) calendar(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	events, err := s.store.Calendar(r.Context(), year)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "calendar not synced for requested season"})
		return
	}
	if err != nil {
		s.logger.Error("list calendar", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": events, "meta": map[string]any{"season": year, "count": len(events)}})
}

func (s *Server) nextEvent(w http.ResponseWriter, r *http.Request) {
	event, err := s.store.NextEvent(r.Context(), time.Now().UTC())
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no upcoming race in synced calendars"})
		return
	}
	if err != nil {
		s.logger.Error("get next event", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": event})
}

func requestedSeason(r *http.Request) (int, error) {
	year, err := strconv.Atoi(r.URL.Query().Get("season"))
	if err != nil || year < 1950 || year > 2100 {
		return 0, errors.New("season must be a year between 1950 and 2100")
	}
	return year, nil
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
