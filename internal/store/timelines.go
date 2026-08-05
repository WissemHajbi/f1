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

func (s *Store) UpsertOvertakes(ctx context.Context, items []domain.Overtake, syncedAt time.Time) error {
	if len(items) == 0 {
		return fmt.Errorf("overtakes cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO overtakes
		(event_key, session_key, occurred_at, meeting_key, driver_number, overtaken_driver_number, position, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(event_key) DO UPDATE SET position=excluded.position, synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		values := []string{strconv.Itoa(item.SessionKey), item.Timestamp.UTC().Format(telemetryTimeLayout),
			strconv.Itoa(item.DriverNumber), strconv.Itoa(item.OvertakenDriverNumber), strconv.Itoa(item.Position)}
		digest := sha256.Sum256([]byte(strings.Join(values, "\x1f")))
		_, err := stmt.ExecContext(ctx, hex.EncodeToString(digest[:]), item.SessionKey,
			item.Timestamp.UTC().Format(telemetryTimeLayout), item.MeetingKey, item.DriverNumber,
			item.OvertakenDriverNumber, item.Position, syncTime)
		if err != nil {
			return fmt.Errorf("upsert overtake: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) UpsertPositions(ctx context.Context, items []domain.PositionSample, syncedAt time.Time) error {
	if len(items) == 0 {
		return fmt.Errorf("positions cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO position_samples
		(session_key, driver_number, sampled_at, meeting_key, position, source, synced_at)
		VALUES (?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, sampled_at) DO UPDATE SET
			meeting_key=excluded.meeting_key, position=excluded.position, synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, item.SessionKey, item.DriverNumber, item.Timestamp.UTC().Format(telemetryTimeLayout),
			item.MeetingKey, item.Position, syncTime); err != nil {
			return fmt.Errorf("upsert position: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) UpsertIntervals(ctx context.Context, items []domain.IntervalSample, syncedAt time.Time) error {
	if len(items) == 0 {
		return fmt.Errorf("intervals cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO interval_samples
		(session_key, driver_number, sampled_at, meeting_key, gap_to_leader, gap_to_leader_seconds,
		 interval_value, interval_seconds, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key, driver_number, sampled_at) DO UPDATE SET meeting_key=excluded.meeting_key,
			gap_to_leader=excluded.gap_to_leader, gap_to_leader_seconds=excluded.gap_to_leader_seconds,
			interval_value=excluded.interval_value, interval_seconds=excluded.interval_seconds, synced_at=excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	syncTime := syncedAt.UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, item.SessionKey, item.DriverNumber, item.Timestamp.UTC().Format(telemetryTimeLayout),
			item.MeetingKey, item.GapToLeader, item.GapToLeaderSeconds, item.Interval, item.IntervalSeconds,
			syncTime); err != nil {
			return fmt.Errorf("upsert interval: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) Overtakes(ctx context.Context, query domain.OvertakeQuery) ([]domain.Overtake, bool, error) {
	statement := `SELECT occurred_at, session_key, meeting_key, driver_number, overtaken_driver_number, position
	              FROM overtakes WHERE session_key=?`
	args := []any{query.SessionKey}
	statement, args = timelineFilters(statement, args, query.TimelineQuery, "occurred_at")
	if query.OvertakenDriverNumber != nil {
		statement += ` AND overtaken_driver_number=?`
		args = append(args, *query.OvertakenDriverNumber)
	}
	statement += ` ORDER BY occurred_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := []domain.Overtake{}
	for rows.Next() {
		var item domain.Overtake
		var at string
		if err := rows.Scan(&at, &item.SessionKey, &item.MeetingKey, &item.DriverNumber,
			&item.OvertakenDriverNumber, &item.Position); err != nil {
			return nil, false, err
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, at)
		if err != nil {
			return nil, false, err
		}
		items = append(items, item)
	}
	return limitedOvertakes(items, query.Limit, rows.Err())
}

func (s *Store) Positions(ctx context.Context, query domain.TimelineQuery) ([]domain.PositionSample, bool, error) {
	statement := `SELECT sampled_at, session_key, meeting_key, driver_number, position FROM position_samples WHERE session_key=?`
	args := []any{query.SessionKey}
	statement, args = timelineFilters(statement, args, query, "sampled_at")
	statement += ` ORDER BY sampled_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := []domain.PositionSample{}
	for rows.Next() {
		var item domain.PositionSample
		var at string
		if err := rows.Scan(&at, &item.SessionKey, &item.MeetingKey, &item.DriverNumber, &item.Position); err != nil {
			return nil, false, err
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, at)
		if err != nil {
			return nil, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
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

func (s *Store) Intervals(ctx context.Context, query domain.TimelineQuery) ([]domain.IntervalSample, bool, error) {
	statement := `SELECT sampled_at, session_key, meeting_key, driver_number, gap_to_leader,
	                     gap_to_leader_seconds, interval_value, interval_seconds
	              FROM interval_samples WHERE session_key=?`
	args := []any{query.SessionKey}
	statement, args = timelineFilters(statement, args, query, "sampled_at")
	statement += ` ORDER BY sampled_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := []domain.IntervalSample{}
	for rows.Next() {
		var item domain.IntervalSample
		var at string
		var gap, interval sql.NullFloat64
		if err := rows.Scan(&at, &item.SessionKey, &item.MeetingKey, &item.DriverNumber, &item.GapToLeader,
			&gap, &item.Interval, &interval); err != nil {
			return nil, false, err
		}
		item.Timestamp, err = time.Parse(telemetryTimeLayout, at)
		if err != nil {
			return nil, false, err
		}
		item.GapToLeaderSeconds, item.IntervalSeconds = nullableFloat(gap), nullableFloat(interval)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
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

func timelineFilters(statement string, args []any, query domain.TimelineQuery, timeColumn string) (string, []any) {
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	if query.From != nil {
		statement += ` AND ` + timeColumn + `>=?`
		args = append(args, query.From.UTC().Format(telemetryTimeLayout))
	}
	if query.To != nil {
		statement += ` AND ` + timeColumn + `<=?`
		args = append(args, query.To.UTC().Format(telemetryTimeLayout))
	}
	return statement, args
}

func limitedOvertakes(items []domain.Overtake, limit int, err error) ([]domain.Overtake, bool, error) {
	if err != nil {
		return nil, false, err
	}
	truncated := len(items) > limit
	if truncated {
		items = items[:limit]
	}
	if len(items) == 0 {
		return nil, false, ErrNotFound
	}
	return items, truncated, nil
}
