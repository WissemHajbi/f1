package store

import (
	"context"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) ReplaceMeetings(ctx context.Context, year int, meetings []domain.Meeting) error {
	if len(meetings) == 0 {
		return fmt.Errorf("meetings cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin meetings transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM openf1_meetings WHERE year=?`, year); err != nil {
		return fmt.Errorf("clear meetings: %w", err)
	}
	meetingStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO openf1_meetings
		(meeting_key, year, name, official_name, location, country_key, country_code, country_name,
		 circuit_key, circuit_short_name, date_start, gmt_offset, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare meeting: %w", err)
	}
	defer meetingStatement.Close()
	sessionStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO openf1_sessions
		(session_key, meeting_key, year, name, type, location, country_code, country_name, circuit_key,
		 circuit_short_name, date_start, date_end, gmt_offset, is_cancelled, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare session: %w", err)
	}
	defer sessionStatement.Close()
	for _, meeting := range meetings {
		if meeting.Year != year || meeting.Key <= 0 {
			return fmt.Errorf("invalid meeting %d", meeting.Key)
		}
		syncedAt := meeting.SyncedAt.UTC().Format(time.RFC3339Nano)
		_, err := meetingStatement.ExecContext(ctx, meeting.Key, meeting.Year, meeting.Name, meeting.OfficialName,
			meeting.Location, meeting.CountryKey, meeting.CountryCode, meeting.CountryName, meeting.CircuitKey,
			meeting.CircuitShortName, meeting.DateStart.UTC().Format(time.RFC3339Nano), meeting.GMTOffset, syncedAt)
		if err != nil {
			return fmt.Errorf("insert meeting %d: %w", meeting.Key, err)
		}
		for _, session := range meeting.Sessions {
			_, err := sessionStatement.ExecContext(ctx, session.Key, meeting.Key, session.Year, session.Name, session.Type,
				session.Location, session.CountryCode, session.CountryName, session.CircuitKey, session.CircuitShortName,
				session.DateStart.UTC().Format(time.RFC3339Nano), session.DateEnd.UTC().Format(time.RFC3339Nano),
				session.GMTOffset, session.IsCancelled, syncedAt)
			if err != nil {
				return fmt.Errorf("insert session %d: %w", session.Key, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit meetings: %w", err)
	}
	return nil
}

func (s *Store) Meetings(ctx context.Context, year int) ([]domain.Meeting, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT meeting_key, year, name, official_name, location, country_key, country_code, country_name,
		       circuit_key, circuit_short_name, date_start, gmt_offset, synced_at
		FROM openf1_meetings WHERE year=? ORDER BY date_start
	`, year)
	if err != nil {
		return nil, fmt.Errorf("query meetings: %w", err)
	}
	defer rows.Close()
	meetings := []domain.Meeting{}
	for rows.Next() {
		var item domain.Meeting
		var dateStart, syncedAt string
		if err := rows.Scan(&item.Key, &item.Year, &item.Name, &item.OfficialName, &item.Location, &item.CountryKey,
			&item.CountryCode, &item.CountryName, &item.CircuitKey, &item.CircuitShortName, &dateStart,
			&item.GMTOffset, &syncedAt); err != nil {
			return nil, fmt.Errorf("scan meeting: %w", err)
		}
		item.DateStart, err = time.Parse(time.RFC3339Nano, dateStart)
		if err != nil {
			return nil, fmt.Errorf("parse meeting start: %w", err)
		}
		item.SyncedAt, err = time.Parse(time.RFC3339Nano, syncedAt)
		if err != nil {
			return nil, fmt.Errorf("parse meeting sync: %w", err)
		}
		item.Sessions = []domain.MeetingSession{}
		meetings = append(meetings, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meetings: %w", err)
	}
	if len(meetings) == 0 {
		return nil, ErrNotFound
	}
	if err := s.attachMeetingSessions(ctx, year, meetings); err != nil {
		return nil, err
	}
	return meetings, nil
}

func (s *Store) Sessions(ctx context.Context, year int, meetingKey *int) ([]domain.MeetingSession, error) {
	query := `SELECT session_key, meeting_key, year, name, type, location, country_code, country_name, circuit_key,
	                 circuit_short_name, date_start, date_end, gmt_offset, is_cancelled
	          FROM openf1_sessions WHERE year=?`
	args := []any{year}
	if meetingKey != nil {
		query += ` AND meeting_key=?`
		args = append(args, *meetingKey)
	}
	query += ` ORDER BY date_start`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()
	items := []domain.MeetingSession{}
	for rows.Next() {
		item, err := scanMeetingSession(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items, nil
}

func (s *Store) attachMeetingSessions(ctx context.Context, year int, meetings []domain.Meeting) error {
	items, err := s.Sessions(ctx, year, nil)
	if err != nil {
		return err
	}
	byKey := make(map[int]*domain.Meeting, len(meetings))
	for index := range meetings {
		byKey[meetings[index].Key] = &meetings[index]
	}
	for _, session := range items {
		if meeting := byKey[session.MeetingKey]; meeting != nil {
			meeting.Sessions = append(meeting.Sessions, session)
		}
	}
	return nil
}

type meetingSessionScan func(dest ...any) error

func scanMeetingSession(scan meetingSessionScan) (domain.MeetingSession, error) {
	var item domain.MeetingSession
	var dateStart, dateEnd string
	if err := scan(&item.Key, &item.MeetingKey, &item.Year, &item.Name, &item.Type, &item.Location,
		&item.CountryCode, &item.CountryName, &item.CircuitKey, &item.CircuitShortName, &dateStart,
		&dateEnd, &item.GMTOffset, &item.IsCancelled); err != nil {
		return domain.MeetingSession{}, fmt.Errorf("scan session: %w", err)
	}
	var err error
	item.DateStart, err = time.Parse(time.RFC3339Nano, dateStart)
	if err != nil {
		return domain.MeetingSession{}, fmt.Errorf("parse session start: %w", err)
	}
	item.DateEnd, err = time.Parse(time.RFC3339Nano, dateEnd)
	if err != nil {
		return domain.MeetingSession{}, fmt.Errorf("parse session end: %w", err)
	}
	return item, nil
}
