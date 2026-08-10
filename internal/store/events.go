package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) ReplaceCalendar(ctx context.Context, season int, events []domain.Event) error {
	if len(events) == 0 {
		return fmt.Errorf("calendar events cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin calendar transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM event_sessions WHERE season=?`, season); err != nil {
		return fmt.Errorf("clear calendar sessions: %w", err)
	}
	eventStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO events
		(season, round, name, source_url, race_at, circuit_id, circuit_name, locality, country,
		 latitude, longitude, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'jolpica', ?)
		ON CONFLICT(season, round) DO UPDATE SET
			name=excluded.name, source_url=excluded.source_url, race_at=excluded.race_at,
			circuit_id=excluded.circuit_id, circuit_name=excluded.circuit_name,
			locality=excluded.locality, country=excluded.country, latitude=excluded.latitude,
			longitude=excluded.longitude, source=excluded.source, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare event insert: %w", err)
	}
	defer eventStatement.Close()
	sessionStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO event_sessions (season, round, type, start_at) VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare event session insert: %w", err)
	}
	defer sessionStatement.Close()
	for _, event := range events {
		if event.Season != season || event.RaceAt == nil {
			return fmt.Errorf("invalid event season or race time for round %d", event.Round)
		}
		_, err := eventStatement.ExecContext(ctx, event.Season, event.Round, event.Name, event.SourceURL,
			event.RaceAt.UTC().Format(time.RFC3339Nano), event.Circuit.ID, event.Circuit.Name,
			event.Circuit.Locality, event.Circuit.Country, event.Circuit.Latitude, event.Circuit.Longitude,
			event.SyncedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert event round %d: %w", event.Round, err)
		}
		for _, session := range event.Sessions {
			if session.StartAt == nil {
				continue
			}
			if _, err := sessionStatement.ExecContext(ctx, event.Season, event.Round, session.Type,
				session.StartAt.UTC().Format(time.RFC3339Nano)); err != nil {
				return fmt.Errorf("insert %s for round %d: %w", session.Type, event.Round, err)
			}
		}
	}
	args := make([]any, 0, len(events)+1)
	args = append(args, season)
	for _, event := range events {
		args = append(args, event.Round)
	}
	deleteStale := `DELETE FROM events WHERE season=? AND round NOT IN (` + strings.TrimRight(strings.Repeat("?,", len(events)), ",") + `)`
	if _, err := tx.ExecContext(ctx, deleteStale, args...); err != nil {
		return fmt.Errorf("delete stale calendar events: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit calendar: %w", err)
	}
	return nil
}

func (s *Store) Calendar(ctx context.Context, season int) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT season, round, name, source_url, race_at, circuit_id, circuit_name, locality, country,
		       latitude, longitude, synced_at
		FROM events WHERE season=? ORDER BY round
	`, season)
	if err != nil {
		return nil, fmt.Errorf("query calendar: %w", err)
	}
	defer rows.Close()
	events := []domain.Event{}
	for rows.Next() {
		event, err := scanEvent(rows.Scan)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate calendar: %w", err)
	}
	if len(events) == 0 {
		return nil, ErrNotFound
	}
	if err := s.loadEventSessions(ctx, events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) NextEvent(ctx context.Context, after time.Time) (domain.Event, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT season, round, name, source_url, race_at, circuit_id, circuit_name, locality, country,
		       latitude, longitude, synced_at
		FROM events WHERE race_at>=? ORDER BY race_at LIMIT 1
	`, after.UTC().Format(time.RFC3339Nano))
	event, err := scanEvent(row.Scan)
	if err != nil {
		if err == ErrNotFound {
			return domain.Event{}, err
		}
		return domain.Event{}, fmt.Errorf("query next event: %w", err)
	}
	events := []domain.Event{event}
	if err := s.loadEventSessions(ctx, events); err != nil {
		return domain.Event{}, err
	}
	return events[0], nil
}

type scanFunc func(dest ...any) error

func scanEvent(scan scanFunc) (domain.Event, error) {
	var event domain.Event
	var raceAt, syncedAt string
	err := scan(&event.Season, &event.Round, &event.Name, &event.SourceURL, &raceAt,
		&event.Circuit.ID, &event.Circuit.Name, &event.Circuit.Locality, &event.Circuit.Country,
		&event.Circuit.Latitude, &event.Circuit.Longitude, &syncedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, ErrNotFound
		}
		return domain.Event{}, fmt.Errorf("scan event: %w", err)
	}
	parsedRace, err := time.Parse(time.RFC3339Nano, raceAt)
	if err != nil {
		return domain.Event{}, fmt.Errorf("parse race time: %w", err)
	}
	event.RaceAt = &parsedRace
	event.SyncedAt, err = time.Parse(time.RFC3339Nano, syncedAt)
	if err != nil {
		return domain.Event{}, fmt.Errorf("parse event sync time: %w", err)
	}
	return event, nil
}

func (s *Store) loadEventSessions(ctx context.Context, events []domain.Event) error {
	byEvent := make(map[[2]int]*domain.Event, len(events))
	for index := range events {
		events[index].Sessions = []domain.EventSession{}
		byEvent[[2]int{events[index].Season, events[index].Round}] = &events[index]
	}
	minSeason, maxSeason := events[0].Season, events[0].Season
	for _, event := range events {
		if event.Season < minSeason {
			minSeason = event.Season
		}
		if event.Season > maxSeason {
			maxSeason = event.Season
		}
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT season, round, type, start_at FROM event_sessions
		WHERE season BETWEEN ? AND ? ORDER BY start_at
	`, minSeason, maxSeason)
	if err != nil {
		return fmt.Errorf("query event sessions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var season, round int
		var session domain.EventSession
		var start string
		if err := rows.Scan(&season, &round, &session.Type, &start); err != nil {
			return fmt.Errorf("scan event session: %w", err)
		}
		event := byEvent[[2]int{season, round}]
		if event == nil {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, start)
		if err != nil {
			return fmt.Errorf("parse event session time: %w", err)
		}
		session.StartAt = &parsed
		event.Sessions = append(event.Sessions, session)
	}
	return rows.Err()
}
