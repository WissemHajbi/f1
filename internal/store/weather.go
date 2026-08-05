package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertWeather(ctx context.Context, samples []domain.WeatherSample, syncedAt time.Time) error {
	if len(samples) == 0 {
		return fmt.Errorf("weather samples cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin weather transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO weather_samples
		(session_key, sampled_at, meeting_key, air_temperature, track_temperature, humidity, pressure,
		 rainfall, wind_direction, wind_speed, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, sampled_at) DO UPDATE SET
			meeting_key=excluded.meeting_key, air_temperature=excluded.air_temperature,
			track_temperature=excluded.track_temperature, humidity=excluded.humidity, pressure=excluded.pressure,
			rainfall=excluded.rainfall, wind_direction=excluded.wind_direction, wind_speed=excluded.wind_speed,
			synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare weather upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, sample := range samples {
		_, err := statement.ExecContext(ctx, sample.SessionKey, sample.Timestamp.UTC().Format(telemetryTimeLayout),
			sample.MeetingKey, sample.AirTemperature, sample.TrackTemperature, sample.Humidity, sample.Pressure,
			sample.Rainfall, sample.WindDirection, sample.WindSpeed, syncTime)
		if err != nil {
			return fmt.Errorf("upsert weather at %s: %w", sample.Timestamp, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit weather: %w", err)
	}
	return nil
}

func (s *Store) Weather(ctx context.Context, query domain.WeatherQuery) ([]domain.WeatherSample, bool, error) {
	statement := `SELECT sampled_at, session_key, meeting_key, air_temperature, track_temperature,
	                     humidity, pressure, rainfall, wind_direction, wind_speed
	              FROM weather_samples WHERE session_key=?`
	args := []any{query.SessionKey}
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
		return nil, false, fmt.Errorf("query weather: %w", err)
	}
	defer rows.Close()
	items := []domain.WeatherSample{}
	for rows.Next() {
		var item domain.WeatherSample
		var sampledAt string
		var air, track, humidity, pressure, rainfall, direction, speed sql.NullFloat64
		if err := rows.Scan(&sampledAt, &item.SessionKey, &item.MeetingKey, &air, &track, &humidity,
			&pressure, &rainfall, &direction, &speed); err != nil {
			return nil, false, fmt.Errorf("scan weather: %w", err)
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, sampledAt)
		if err != nil {
			return nil, false, fmt.Errorf("parse weather time: %w", err)
		}
		item.AirTemperature, item.TrackTemperature = nullableFloat(air), nullableFloat(track)
		item.Humidity, item.Pressure, item.Rainfall = nullableFloat(humidity), nullableFloat(pressure), nullableFloat(rainfall)
		item.WindDirection, item.WindSpeed = nullableFloat(direction), nullableFloat(speed)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate weather: %w", err)
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
