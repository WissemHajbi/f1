package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

var ErrNotFound = errors.New("not found")

func (s *Store) SaveDriverRoster(ctx context.Context, roster domain.DriverRoster) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin roster transaction: %w", err)
	}
	defer tx.Rollback()

	syncedAt := roster.SyncedAt.UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (session_key, year, name, type, location, date_start, date_end, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
		ON CONFLICT(session_key) DO UPDATE SET
			year=excluded.year, name=excluded.name, type=excluded.type, location=excluded.location,
			date_start=excluded.date_start, date_end=excluded.date_end, source=excluded.source, synced_at=excluded.synced_at
	`, roster.Session.Key, roster.Session.Year, roster.Session.Name, roster.Session.Type, roster.Session.Location,
		roster.Session.DateStart.UTC().Format(time.RFC3339Nano), roster.Session.DateEnd.UTC().Format(time.RFC3339Nano), syncedAt)
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM drivers WHERE session_key=?`, roster.Session.Key); err != nil {
		return fmt.Errorf("replace drivers: %w", err)
	}
	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO drivers
		(session_key, driver_number, broadcast_name, full_name, acronym, team_name, team_colour,
		 first_name, last_name, headshot_url, country_code, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'openf1', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare driver insert: %w", err)
	}
	defer statement.Close()
	for _, driver := range roster.Drivers {
		if _, err := statement.ExecContext(ctx, roster.Session.Key, driver.Number, driver.BroadcastName, driver.FullName,
			driver.Acronym, driver.TeamName, driver.TeamColour, driver.FirstName, driver.LastName,
			driver.HeadshotURL, driver.CountryCode, syncedAt); err != nil {
			return fmt.Errorf("insert driver %d: %w", driver.Number, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit roster: %w", err)
	}
	return nil
}

func (s *Store) LatestDriverRoster(ctx context.Context, year int) (domain.DriverRoster, error) {
	var roster domain.DriverRoster
	var start, end, synced string
	err := s.db.QueryRowContext(ctx, `
		SELECT session_key, year, name, type, location, date_start, date_end, synced_at
		FROM sessions WHERE year=?
		ORDER BY date_start DESC LIMIT 1
	`, year).Scan(&roster.Session.Key, &roster.Session.Year, &roster.Session.Name, &roster.Session.Type,
		&roster.Session.Location, &start, &end, &synced)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DriverRoster{}, ErrNotFound
	}
	if err != nil {
		return domain.DriverRoster{}, fmt.Errorf("query latest session: %w", err)
	}
	roster.Session.DateStart, err = time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return domain.DriverRoster{}, fmt.Errorf("parse session start: %w", err)
	}
	roster.Session.DateEnd, err = time.Parse(time.RFC3339Nano, end)
	if err != nil {
		return domain.DriverRoster{}, fmt.Errorf("parse session end: %w", err)
	}
	roster.SyncedAt, err = time.Parse(time.RFC3339Nano, synced)
	if err != nil {
		return domain.DriverRoster{}, fmt.Errorf("parse sync time: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT session_key, driver_number, broadcast_name, full_name, acronym, team_name, team_colour,
		       first_name, last_name, headshot_url, country_code
		FROM drivers WHERE session_key=? ORDER BY driver_number
	`, roster.Session.Key)
	if err != nil {
		return domain.DriverRoster{}, fmt.Errorf("query drivers: %w", err)
	}
	defer rows.Close()
	roster.Drivers = []domain.Driver{}
	for rows.Next() {
		var driver domain.Driver
		var country sql.NullString
		if err := rows.Scan(&driver.SessionKey, &driver.Number, &driver.BroadcastName, &driver.FullName, &driver.Acronym,
			&driver.TeamName, &driver.TeamColour, &driver.FirstName, &driver.LastName, &driver.HeadshotURL, &country); err != nil {
			return domain.DriverRoster{}, fmt.Errorf("scan driver: %w", err)
		}
		if country.Valid {
			driver.CountryCode = &country.String
		}
		roster.Drivers = append(roster.Drivers, driver)
	}
	if err := rows.Err(); err != nil {
		return domain.DriverRoster{}, fmt.Errorf("iterate drivers: %w", err)
	}
	return roster, nil
}
