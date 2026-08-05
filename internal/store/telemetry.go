package store

import (
	"context"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

const telemetryTimeLayout = "2006-01-02T15:04:05.000000000Z"

func (s *Store) UpsertCarData(ctx context.Context, samples []domain.CarDataSample, syncedAt time.Time) error {
	if len(samples) == 0 {
		return fmt.Errorf("car data samples cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin car data transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO car_data_samples
		(session_key, driver_number, sampled_at, meeting_key, speed, rpm, gear, throttle, brake, drs, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, sampled_at) DO UPDATE SET
			meeting_key=excluded.meeting_key, speed=excluded.speed, rpm=excluded.rpm, gear=excluded.gear,
			throttle=excluded.throttle, brake=excluded.brake, drs=excluded.drs, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare car data upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, sample := range samples {
		_, err := statement.ExecContext(ctx, sample.SessionKey, sample.DriverNumber,
			sample.Timestamp.UTC().Format(telemetryTimeLayout), sample.MeetingKey, sample.Speed, sample.RPM,
			sample.Gear, sample.Throttle, sample.Brake, sample.DRS, syncTime)
		if err != nil {
			return fmt.Errorf("upsert car data at %s: %w", sample.Timestamp, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit car data: %w", err)
	}
	return nil
}

func (s *Store) CarData(ctx context.Context, query domain.CarDataQuery) ([]domain.CarDataSample, bool, error) {
	statement := `
		SELECT sampled_at, session_key, meeting_key, driver_number, speed, rpm, gear, throttle, brake, drs
		FROM car_data_samples WHERE session_key=? AND driver_number=?`
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
		return nil, false, fmt.Errorf("query car data: %w", err)
	}
	defer rows.Close()
	samples := []domain.CarDataSample{}
	for rows.Next() {
		var sample domain.CarDataSample
		var timestamp string
		if err := rows.Scan(&timestamp, &sample.SessionKey, &sample.MeetingKey, &sample.DriverNumber,
			&sample.Speed, &sample.RPM, &sample.Gear, &sample.Throttle, &sample.Brake, &sample.DRS); err != nil {
			return nil, false, fmt.Errorf("scan car data: %w", err)
		}
		sample.Timestamp, err = time.Parse(telemetryTimeLayout, timestamp)
		if err != nil {
			return nil, false, fmt.Errorf("parse car data time: %w", err)
		}
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate car data: %w", err)
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
