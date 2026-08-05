package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertPitStops(ctx context.Context, items []domain.PitStop, syncedAt time.Time) error {
	if len(items) == 0 {
		return fmt.Errorf("pit stops cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin pit-stop transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO pit_stops
		(session_key, driver_number, stopped_at, meeting_key, lap_number, duration, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, stopped_at) DO UPDATE SET
			meeting_key=excluded.meeting_key, lap_number=excluded.lap_number,
			duration=excluded.duration, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare pit-stop upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		_, err := statement.ExecContext(ctx, item.SessionKey, item.DriverNumber,
			item.Timestamp.UTC().Format(telemetryTimeLayout), item.MeetingKey, item.LapNumber, item.Duration, syncTime)
		if err != nil {
			return fmt.Errorf("upsert driver %d lap %d pit stop: %w", item.DriverNumber, item.LapNumber, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pit stops: %w", err)
	}
	return nil
}

func (s *Store) PitStops(ctx context.Context, query domain.PitStopQuery) ([]domain.PitStop, bool, error) {
	statement := `
		SELECT session_key, meeting_key, driver_number, lap_number, stopped_at, duration
		FROM pit_stops WHERE session_key=?`
	args := []any{query.SessionKey}
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	statement += ` ORDER BY stopped_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query pit stops: %w", err)
	}
	defer rows.Close()
	items := []domain.PitStop{}
	for rows.Next() {
		var item domain.PitStop
		var stoppedAt string
		var duration sql.NullFloat64
		if err := rows.Scan(&item.SessionKey, &item.MeetingKey, &item.DriverNumber, &item.LapNumber,
			&stoppedAt, &duration); err != nil {
			return nil, false, fmt.Errorf("scan pit stop: %w", err)
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, stoppedAt)
		if err != nil {
			return nil, false, fmt.Errorf("parse pit-stop time: %w", err)
		}
		item.Duration = nullableFloat(duration)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate pit stops: %w", err)
	}
	truncated := len(items) > query.Limit
	if truncated {
		items = items[:query.Limit]
	}
	if len(items) == 0 {
		return nil, false, ErrNotFound
	}
	return items, truncated, nil
}
