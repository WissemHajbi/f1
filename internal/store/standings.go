package store

import (
	"context"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) ReplaceStandings(ctx context.Context, season int, drivers []domain.DriverStanding, constructors []domain.ConstructorStanding) error {
	if len(drivers) == 0 || len(constructors) == 0 {
		return fmt.Errorf("standings cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin standings transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM driver_standings WHERE season=?`, season); err != nil {
		return fmt.Errorf("clear driver standings: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM constructor_standings WHERE season=?`, season); err != nil {
		return fmt.Errorf("clear constructor standings: %w", err)
	}
	driverStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO driver_standings
		(season, driver_id, round, position, points, wins, permanent_number, code, given_name,
		 family_name, date_of_birth, nationality, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'jolpica', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare driver standing: %w", err)
	}
	defer driverStatement.Close()
	teamStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO driver_standing_constructors
		(season, driver_id, constructor_id, constructor_name, constructor_nationality) VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare driver constructor: %w", err)
	}
	defer teamStatement.Close()
	for _, item := range drivers {
		if item.Season != season {
			return fmt.Errorf("unexpected driver standing season %d", item.Season)
		}
		_, err := driverStatement.ExecContext(ctx, item.Season, item.DriverID, item.Round, item.Position, item.Points,
			item.Wins, item.PermanentNumber, item.Code, item.GivenName, item.FamilyName, item.DateOfBirth,
			item.Nationality, item.SyncedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert driver %s standing: %w", item.DriverID, err)
		}
		for _, constructor := range item.Constructors {
			if _, err := teamStatement.ExecContext(ctx, season, item.DriverID, constructor.ID, constructor.Name,
				constructor.Nationality); err != nil {
				return fmt.Errorf("insert driver %s constructor: %w", item.DriverID, err)
			}
		}
	}
	constructorStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO constructor_standings
		(season, constructor_id, round, position, points, wins, constructor_name, constructor_nationality, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'jolpica', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare constructor standing: %w", err)
	}
	defer constructorStatement.Close()
	for _, item := range constructors {
		if item.Season != season {
			return fmt.Errorf("unexpected constructor standing season %d", item.Season)
		}
		_, err := constructorStatement.ExecContext(ctx, item.Season, item.Constructor.ID, item.Round, item.Position,
			item.Points, item.Wins, item.Constructor.Name, item.Constructor.Nationality,
			item.SyncedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert constructor %s standing: %w", item.Constructor.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit standings: %w", err)
	}
	return nil
}

func (s *Store) DriverStandings(ctx context.Context, season int) ([]domain.DriverStanding, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT season, round, position, points, wins, driver_id, permanent_number, code, given_name,
		       family_name, date_of_birth, nationality, synced_at
		FROM driver_standings WHERE season=? ORDER BY position
	`, season)
	if err != nil {
		return nil, fmt.Errorf("query driver standings: %w", err)
	}
	defer rows.Close()
	items := []domain.DriverStanding{}
	for rows.Next() {
		var item domain.DriverStanding
		var synced string
		if err := rows.Scan(&item.Season, &item.Round, &item.Position, &item.Points, &item.Wins, &item.DriverID,
			&item.PermanentNumber, &item.Code, &item.GivenName, &item.FamilyName, &item.DateOfBirth,
			&item.Nationality, &synced); err != nil {
			return nil, fmt.Errorf("scan driver standing: %w", err)
		}
		item.SyncedAt, err = time.Parse(time.RFC3339Nano, synced)
		if err != nil {
			return nil, fmt.Errorf("parse driver standing sync time: %w", err)
		}
		item.Constructors = []domain.StandingConstructor{}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate driver standings: %w", err)
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	if err := s.loadDriverStandingConstructors(ctx, season, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) loadDriverStandingConstructors(ctx context.Context, season int, items []domain.DriverStanding) error {
	byID := make(map[string]*domain.DriverStanding, len(items))
	for index := range items {
		byID[items[index].DriverID] = &items[index]
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT driver_id, constructor_id, constructor_name, constructor_nationality
		FROM driver_standing_constructors WHERE season=? ORDER BY driver_id, constructor_id
	`, season)
	if err != nil {
		return fmt.Errorf("query driver constructors: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var driverID string
		var constructor domain.StandingConstructor
		if err := rows.Scan(&driverID, &constructor.ID, &constructor.Name, &constructor.Nationality); err != nil {
			return fmt.Errorf("scan driver constructor: %w", err)
		}
		if item := byID[driverID]; item != nil {
			item.Constructors = append(item.Constructors, constructor)
		}
	}
	return rows.Err()
}

func (s *Store) ConstructorStandings(ctx context.Context, season int) ([]domain.ConstructorStanding, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT season, round, position, points, wins, constructor_id, constructor_name,
		       constructor_nationality, synced_at
		FROM constructor_standings WHERE season=? ORDER BY position
	`, season)
	if err != nil {
		return nil, fmt.Errorf("query constructor standings: %w", err)
	}
	defer rows.Close()
	items := []domain.ConstructorStanding{}
	for rows.Next() {
		var item domain.ConstructorStanding
		var synced string
		if err := rows.Scan(&item.Season, &item.Round, &item.Position, &item.Points, &item.Wins,
			&item.Constructor.ID, &item.Constructor.Name, &item.Constructor.Nationality, &synced); err != nil {
			return nil, fmt.Errorf("scan constructor standing: %w", err)
		}
		item.SyncedAt, err = time.Parse(time.RFC3339Nano, synced)
		if err != nil {
			return nil, fmt.Errorf("parse constructor standing sync time: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate constructor standings: %w", err)
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return items, nil
}
