package store

import (
	"context"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertLocation(ctx context.Context, samples []domain.LocationSample, syncedAt time.Time) error {
	if len(samples) == 0 {
		return fmt.Errorf("location samples cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin location transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `INSERT INTO location_samples
		(session_key, driver_number, sampled_at, meeting_key, x, y, z, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, sampled_at) DO UPDATE SET
			meeting_key=excluded.meeting_key, x=excluded.x, y=excluded.y, z=excluded.z, synced_at=excluded.synced_at`)
	if err != nil {
		return fmt.Errorf("prepare location upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, sample := range samples {
		if _, err := statement.ExecContext(ctx, sample.SessionKey, sample.DriverNumber,
			sample.Timestamp.UTC().Format(telemetryTimeLayout), sample.MeetingKey,
			sample.X, sample.Y, sample.Z, syncTime); err != nil {
			return fmt.Errorf("upsert location at %s: %w", sample.Timestamp, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit location: %w", err)
	}
	return nil
}

func (s *Store) Location(ctx context.Context, query domain.LocationQuery) ([]domain.LocationSample, bool, error) {
	statement := `SELECT sampled_at, session_key, meeting_key, driver_number, x, y, z
	              FROM location_samples WHERE session_key=? AND driver_number=?`
	args := []any{query.SessionKey, query.DriverNumber}
	if query.From != nil {
		statement += ` AND sampled_at>=?`
		args = append(args, query.From.UTC().Format(telemetryTimeLayout))
	}
	if query.To != nil {
		statement += ` AND sampled_at<=?`
		args = append(args, query.To.UTC().Format(telemetryTimeLayout))
	}
	statement += ` ORDER BY sampled_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query location: %w", err)
	}
	defer rows.Close()
	samples := []domain.LocationSample{}
	for rows.Next() {
		var sample domain.LocationSample
		var timestamp string
		if err := rows.Scan(&timestamp, &sample.SessionKey, &sample.MeetingKey, &sample.DriverNumber,
			&sample.X, &sample.Y, &sample.Z); err != nil {
			return nil, false, fmt.Errorf("scan location: %w", err)
		}
		sample.Timestamp, err = time.Parse(telemetryTimeLayout, timestamp)
		if err != nil {
			return nil, false, fmt.Errorf("parse location time: %w", err)
		}
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate location: %w", err)
	}
	truncated := len(samples) > query.Limit
	if truncated {
		samples = samples[:query.Limit]
	}
	if len(samples) == 0 {
		return nil, false, ErrNotFound
	}
	return samples, truncated, nil
}
