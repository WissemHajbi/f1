package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertRaceControl(ctx context.Context, events []domain.RaceControlEvent, syncedAt time.Time) error {
	if len(events) == 0 {
		return fmt.Errorf("race-control events cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin race-control transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO race_control_events
		(event_key, session_key, occurred_at, meeting_key, category, message, flag, scope,
		 driver_number, lap_number, sector, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(event_key) DO UPDATE SET category=excluded.category, message=excluded.message,
			flag=excluded.flag, scope=excluded.scope, driver_number=excluded.driver_number,
			lap_number=excluded.lap_number, sector=excluded.sector, synced_at=excluded.synced_at
	`)
	if err != nil {
		return fmt.Errorf("prepare race-control upsert: %w", err)
	}
	defer statement.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, event := range events {
		key := raceControlKey(event)
		_, err := statement.ExecContext(ctx, key, event.SessionKey, event.Timestamp.UTC().Format(telemetryTimeLayout),
			event.MeetingKey, event.Category, event.Message, event.Flag, event.Scope, event.DriverNumber,
			event.LapNumber, event.Sector, syncTime)
		if err != nil {
			return fmt.Errorf("upsert race-control event: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit race-control events: %w", err)
	}
	return nil
}

func raceControlKey(event domain.RaceControlEvent) string {
	values := []string{strconv.Itoa(event.SessionKey), event.Timestamp.UTC().Format(telemetryTimeLayout),
		event.Category, event.Message, event.Flag, event.Scope, optionalIntString(event.DriverNumber),
		optionalIntString(event.LapNumber), optionalIntString(event.Sector)}
	digest := sha256.Sum256([]byte(strings.Join(values, "\x1f")))
	return hex.EncodeToString(digest[:])
}

func optionalIntString(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func (s *Store) RaceControl(ctx context.Context, query domain.RaceControlQuery) ([]domain.RaceControlEvent, bool, error) {
	statement := `SELECT occurred_at, session_key, meeting_key, category, message, flag, scope,
	                     driver_number, lap_number, sector
	              FROM race_control_events WHERE session_key=?`
	args := []any{query.SessionKey}
	if query.Category != "" {
		statement += ` AND category=? COLLATE NOCASE`
		args = append(args, query.Category)
	}
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	if query.From != nil {
		statement += ` AND occurred_at>=?`
		args = append(args, query.From.UTC().Format(telemetryTimeLayout))
	}
	if query.To != nil {
		statement += ` AND occurred_at<=?`
		args = append(args, query.To.UTC().Format(telemetryTimeLayout))
	}
	statement += ` ORDER BY occurred_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query race control: %w", err)
	}
	defer rows.Close()
	items := []domain.RaceControlEvent{}
	for rows.Next() {
		var item domain.RaceControlEvent
		var occurredAt string
		var driver, lap, sector sql.NullInt64
		if err := rows.Scan(&occurredAt, &item.SessionKey, &item.MeetingKey, &item.Category, &item.Message,
			&item.Flag, &item.Scope, &driver, &lap, &sector); err != nil {
			return nil, false, fmt.Errorf("scan race control: %w", err)
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, occurredAt)
		if err != nil {
			return nil, false, fmt.Errorf("parse race-control time: %w", err)
		}
		item.DriverNumber, item.LapNumber, item.Sector = nullableInt(driver), nullableInt(lap), nullableInt(sector)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate race control: %w", err)
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
