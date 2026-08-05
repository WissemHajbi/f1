package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertLaps(ctx context.Context, laps []domain.Lap, syncedAt time.Time) error {
	if len(laps) == 0 {
		return fmt.Errorf("laps cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin laps transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO laps
		(session_key, driver_number, lap_number, meeting_key, date_start, lap_duration,
		 sector_1_duration, sector_2_duration, sector_3_duration, i1_speed, i2_speed, speed_trap,
		 sector_1_segments, sector_2_segments, sector_3_segments, is_pit_out_lap, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, lap_number) DO UPDATE SET
			meeting_key=excluded.meeting_key, date_start=excluded.date_start, lap_duration=excluded.lap_duration,
			sector_1_duration=excluded.sector_1_duration, sector_2_duration=excluded.sector_2_duration,
			sector_3_duration=excluded.sector_3_duration, i1_speed=excluded.i1_speed, i2_speed=excluded.i2_speed,
			speed_trap=excluded.speed_trap, sector_1_segments=excluded.sector_1_segments,
			sector_2_segments=excluded.sector_2_segments, sector_3_segments=excluded.sector_3_segments,
			is_pit_out_lap=excluded.is_pit_out_lap, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare lap upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, lap := range laps {
		segments1, _ := json.Marshal(lap.Sector1Segments)
		segments2, _ := json.Marshal(lap.Sector2Segments)
		segments3, _ := json.Marshal(lap.Sector3Segments)
		var dateStart any
		if lap.DateStart != nil {
			dateStart = lap.DateStart.UTC().Format(telemetryTimeLayout)
		}
		_, err := statement.ExecContext(ctx, lap.SessionKey, lap.DriverNumber, lap.LapNumber, lap.MeetingKey,
			dateStart, lap.LapDuration, lap.Sector1Duration, lap.Sector2Duration, lap.Sector3Duration,
			lap.I1Speed, lap.I2Speed, lap.SpeedTrap, segments1, segments2, segments3, lap.IsPitOutLap, syncTime)
		if err != nil {
			return fmt.Errorf("upsert driver %d lap %d: %w", lap.DriverNumber, lap.LapNumber, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit laps: %w", err)
	}
	return nil
}

func (s *Store) Laps(ctx context.Context, query domain.LapQuery) ([]domain.Lap, bool, error) {
	statement := `
		SELECT session_key, meeting_key, driver_number, lap_number, date_start, lap_duration,
		       sector_1_duration, sector_2_duration, sector_3_duration, i1_speed, i2_speed, speed_trap,
		       sector_1_segments, sector_2_segments, sector_3_segments, is_pit_out_lap
		FROM laps WHERE session_key=?`
	args := []any{query.SessionKey}
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	if query.LapNumber != nil {
		statement += ` AND lap_number=?`
		args = append(args, *query.LapNumber)
	}
	statement += ` ORDER BY driver_number, lap_number LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query laps: %w", err)
	}
	defer rows.Close()
	laps := []domain.Lap{}
	for rows.Next() {
		var lap domain.Lap
		var dateStart sql.NullString
		var lapDuration, sector1, sector2, sector3 sql.NullFloat64
		var i1, i2, speedTrap sql.NullInt64
		var segments1, segments2, segments3 []byte
		if err := rows.Scan(&lap.SessionKey, &lap.MeetingKey, &lap.DriverNumber, &lap.LapNumber, &dateStart,
			&lapDuration, &sector1, &sector2, &sector3, &i1, &i2, &speedTrap,
			&segments1, &segments2, &segments3, &lap.IsPitOutLap); err != nil {
			return nil, false, fmt.Errorf("scan lap: %w", err)
		}
		if dateStart.Valid {
			parsed, err := time.Parse(telemetryTimeLayout, dateStart.String)
			if err != nil {
				return nil, false, fmt.Errorf("parse lap start: %w", err)
			}
			lap.DateStart = &parsed
		}
		lap.LapDuration = nullableFloat(lapDuration)
		lap.Sector1Duration = nullableFloat(sector1)
		lap.Sector2Duration = nullableFloat(sector2)
		lap.Sector3Duration = nullableFloat(sector3)
		lap.I1Speed = nullableInt(i1)
		lap.I2Speed = nullableInt(i2)
		lap.SpeedTrap = nullableInt(speedTrap)
		if err := json.Unmarshal(segments1, &lap.Sector1Segments); err != nil {
			return nil, false, fmt.Errorf("decode sector 1 segments: %w", err)
		}
		if err := json.Unmarshal(segments2, &lap.Sector2Segments); err != nil {
			return nil, false, fmt.Errorf("decode sector 2 segments: %w", err)
		}
		if err := json.Unmarshal(segments3, &lap.Sector3Segments); err != nil {
			return nil, false, fmt.Errorf("decode sector 3 segments: %w", err)
		}
		laps = append(laps, lap)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate laps: %w", err)
	}
	truncated := len(laps) > query.Limit
	if truncated {
		laps = laps[:query.Limit]
	}
	if len(laps) == 0 {
		return nil, false, ErrNotFound
	}
	return laps, truncated, nil
}

func nullableFloat(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	return &value.Float64
}

func nullableInt(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	converted := int(value.Int64)
	return &converted
}
