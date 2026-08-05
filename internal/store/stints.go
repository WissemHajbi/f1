package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertStints(ctx context.Context, stints []domain.Stint, syncedAt time.Time) error {
	if len(stints) == 0 {
		return fmt.Errorf("stints cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin stints transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO stints
		(session_key, driver_number, stint_number, meeting_key, lap_start, lap_end, compound,
		 tyre_age_at_start, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, stint_number) DO UPDATE SET
			meeting_key=excluded.meeting_key, lap_start=excluded.lap_start, lap_end=excluded.lap_end,
			compound=excluded.compound, tyre_age_at_start=excluded.tyre_age_at_start, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare stint upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, stint := range stints {
		_, err := statement.ExecContext(ctx, stint.SessionKey, stint.DriverNumber, stint.StintNumber,
			stint.MeetingKey, stint.LapStart, stint.LapEnd, stint.Compound, stint.TyreAgeAtStart, syncTime)
		if err != nil {
			return fmt.Errorf("upsert driver %d stint %d: %w", stint.DriverNumber, stint.StintNumber, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stints: %w", err)
	}
	return nil
}

func (s *Store) Stints(ctx context.Context, query domain.StintQuery) ([]domain.Stint, bool, error) {
	statement := `
		SELECT session_key, meeting_key, driver_number, stint_number, lap_start, lap_end, compound, tyre_age_at_start
		FROM stints WHERE session_key=?`
	args := []any{query.SessionKey}
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	statement += ` ORDER BY driver_number, stint_number LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query stints: %w", err)
	}
	defer rows.Close()
	items := []domain.Stint{}
	for rows.Next() {
		var item domain.Stint
		var lapEnd, tyreAge sql.NullInt64
		if err := rows.Scan(&item.SessionKey, &item.MeetingKey, &item.DriverNumber, &item.StintNumber,
			&item.LapStart, &lapEnd, &item.Compound, &tyreAge); err != nil {
			return nil, false, fmt.Errorf("scan stint: %w", err)
		}
		item.LapEnd = nullableInt(lapEnd)
		item.TyreAgeAtStart = nullableInt(tyreAge)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate stints: %w", err)
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
