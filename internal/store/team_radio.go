package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) TeamRadioExists(ctx context.Context, id string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM team_radio WHERE id=?)`, id).Scan(&exists)
	return exists == 1, err
}

func (s *Store) UpsertTeamRadio(ctx context.Context, records []domain.TeamRadio, syncedAt time.Time) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin team radio transaction: %w", err)
	}
	defer tx.Rollback()
	statement, err := tx.PrepareContext(ctx, `INSERT INTO team_radio
		(id, session_key, recorded_at, meeting_key, driver_number, recording_source, content_type, audio_data, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(id) DO UPDATE SET meeting_key=excluded.meeting_key, recording_source=excluded.recording_source,
			content_type=excluded.content_type, audio_data=excluded.audio_data, synced_at=excluded.synced_at`)
	if err != nil {
		return fmt.Errorf("prepare team radio upsert: %w", err)
	}
	defer statement.Close()
	for _, record := range records {
		if len(record.Audio) == 0 {
			return fmt.Errorf("team radio %s has no cached audio", record.ID)
		}
		if _, err := statement.ExecContext(ctx, record.ID, record.SessionKey,
			record.Timestamp.UTC().Format(telemetryTimeLayout), record.MeetingKey, record.DriverNumber,
			record.RecordingSource, record.ContentType, record.Audio, syncedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("upsert team radio %s: %w", record.ID, err)
		}
	}
	return tx.Commit()
}

func (s *Store) TeamRadio(ctx context.Context, query domain.TeamRadioQuery) ([]domain.TeamRadio, bool, error) {
	statement := `SELECT id, recorded_at, session_key, meeting_key, driver_number, content_type
	              FROM team_radio WHERE session_key=?`
	args := []any{query.SessionKey}
	if query.DriverNumber != nil {
		statement += ` AND driver_number=?`
		args = append(args, *query.DriverNumber)
	}
	if query.From != nil {
		statement += ` AND recorded_at>=?`
		args = append(args, query.From.UTC().Format(telemetryTimeLayout))
	}
	if query.To != nil {
		statement += ` AND recorded_at<=?`
		args = append(args, query.To.UTC().Format(telemetryTimeLayout))
	}
	statement += ` ORDER BY recorded_at LIMIT ?`
	args = append(args, query.Limit+1)
	rows, err := s.db.QueryContext(ctx, statement, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query team radio: %w", err)
	}
	defer rows.Close()
	records := []domain.TeamRadio{}
	for rows.Next() {
		var record domain.TeamRadio
		var timestamp string
		if err := rows.Scan(&record.ID, &timestamp, &record.SessionKey, &record.MeetingKey,
			&record.DriverNumber, &record.ContentType); err != nil {
			return nil, false, fmt.Errorf("scan team radio: %w", err)
		}
		record.Timestamp, err = time.Parse(telemetryTimeLayout, timestamp)
		if err != nil {
			return nil, false, fmt.Errorf("parse team radio time: %w", err)
		}
		record.AudioURL = "/v1/team-radio/" + record.ID + "/audio"
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	truncated := len(records) > query.Limit
	if truncated {
		records = records[:query.Limit]
	}
	if len(records) == 0 {
		return nil, false, ErrNotFound
	}
	return records, truncated, nil
}

func (s *Store) TeamRadioAudio(ctx context.Context, id string) ([]byte, string, error) {
	var audio []byte
	var contentType string
	err := s.db.QueryRowContext(ctx, `SELECT audio_data, content_type FROM team_radio WHERE id=?`, id).Scan(&audio, &contentType)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("query team radio audio: %w", err)
	}
	return audio, contentType, nil
}
