package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"oidysts/internal/domain"
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
	mux.HandleFunc("GET /v1/results", s.results)
	mux.HandleFunc("GET /v1/results/latest", s.latestResult)
	mux.HandleFunc("GET /v1/meetings", s.meetings)
	mux.HandleFunc("GET /v1/sessions", s.sessions)
	mux.HandleFunc("GET /v1/car-data", s.carData)
	mux.HandleFunc("GET /v1/laps", s.laps)
	mux.HandleFunc("GET /v1/stints", s.stints)
	mux.HandleFunc("GET /v1/pit-stops", s.pitStops)
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

func (s *Server) pitStops(w http.ResponseWriter, r *http.Request) {
	sessionKey, err := positiveQueryInt(r, "session_key")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	driverNumber, err := optionalPositiveInt(r, "driver_number")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	limit := 500
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 2000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 2000"})
			return
		}
	}
	items, truncated, err := s.store.PitStops(r.Context(), domain.PitStopQuery{SessionKey: sessionKey,
		DriverNumber: driverNumber, Limit: limit})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "pit stops not synced for requested filters"})
		return
	}
	if err != nil {
		s.logger.Error("list pit stops", "session_key", sessionKey, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{
		"session_key": sessionKey, "count": len(items), "truncated": truncated,
	}})
}

func (s *Server) stints(w http.ResponseWriter, r *http.Request) {
	sessionKey, err := positiveQueryInt(r, "session_key")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	driverNumber, err := optionalPositiveInt(r, "driver_number")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	limit := 500
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 2000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 2000"})
			return
		}
	}
	items, truncated, err := s.store.Stints(r.Context(), domain.StintQuery{SessionKey: sessionKey,
		DriverNumber: driverNumber, Limit: limit})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "stints not synced for requested filters"})
		return
	}
	if err != nil {
		s.logger.Error("list stints", "session_key", sessionKey, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{
		"session_key": sessionKey, "count": len(items), "truncated": truncated,
	}})
}

func (s *Server) laps(w http.ResponseWriter, r *http.Request) {
	sessionKey, err := positiveQueryInt(r, "session_key")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	driverNumber, err := optionalPositiveInt(r, "driver_number")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	lapNumber, err := optionalPositiveInt(r, "lap_number")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	limit := 2000
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 5000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 5000"})
			return
		}
	}
	items, truncated, err := s.store.Laps(r.Context(), domain.LapQuery{SessionKey: sessionKey,
		DriverNumber: driverNumber, LapNumber: lapNumber, Limit: limit})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "laps not synced for requested filters"})
		return
	}
	if err != nil {
		s.logger.Error("list laps", "session_key", sessionKey, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{
		"session_key": sessionKey, "count": len(items), "truncated": truncated,
	}})
}

func optionalPositiveInt(r *http.Request, name string) (*int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return nil, errors.New(name + " must be a positive integer")
	}
	return &parsed, nil
}

func (s *Server) carData(w http.ResponseWriter, r *http.Request) {
	sessionKey, err := positiveQueryInt(r, "session_key")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	driverNumber, err := positiveQueryInt(r, "driver_number")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	from, err := optionalQueryTime(r, "from")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	to, err := optionalQueryTime(r, "to")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if from != nil && to != nil && !to.After(*from) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "to must be after from"})
		return
	}
	limit := 1000
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit <= 0 || limit > 5000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 5000"})
			return
		}
	}
	samples, truncated, err := s.store.CarData(r.Context(), domain.CarDataQuery{SessionKey: sessionKey,
		DriverNumber: driverNumber, From: from, To: to, Limit: limit})
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "car data not synced for requested filters"})
		return
	}
	if err != nil {
		s.logger.Error("list car data", "session_key", sessionKey, "driver_number", driverNumber, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": samples, "meta": map[string]any{
		"session_key": sessionKey, "driver_number": driverNumber, "count": len(samples), "truncated": truncated,
	}})
}

func positiveQueryInt(r *http.Request, name string) (int, error) {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || value <= 0 {
		return 0, errors.New(name + " must be a positive integer")
	}
	return value, nil
}

func optionalQueryTime(r *http.Request, name string) (*time.Time, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, errors.New(name + " must be an RFC3339 timestamp")
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func (s *Server) meetings(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	items, err := s.store.Meetings(r.Context(), year)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "meetings not synced for requested season"})
		return
	}
	if err != nil {
		s.logger.Error("list meetings", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{"season": year, "count": len(items)}})
}

func (s *Server) sessions(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var meetingKey *int
	if value := r.URL.Query().Get("meeting_key"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "meeting_key must be a positive integer"})
			return
		}
		meetingKey = &parsed
	}
	items, err := s.store.Sessions(r.Context(), year, meetingKey)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "sessions not synced for requested filters"})
		return
	}
	if err != nil {
		s.logger.Error("list sessions", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{"season": year, "count": len(items)}})
}

func (s *Server) results(w http.ResponseWriter, r *http.Request) {
	year, err := requestedSeason(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var round *int
	if value := r.URL.Query().Get("round"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "round must be a positive integer"})
			return
		}
		round = &parsed
	}
	races, err := s.store.Results(r.Context(), year, round)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "results not synced for requested season or round"})
		return
	}
	if err != nil {
		s.logger.Error("list results", "season", year, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": races, "meta": map[string]any{"season": year, "races": len(races)}})
}

func (s *Server) latestResult(w http.ResponseWriter, r *http.Request) {
	race, err := s.store.LatestResult(r.Context(), time.Now().UTC())
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no completed results have been synced"})
		return
	}
	if err != nil {
		s.logger.Error("get latest result", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": race})
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
