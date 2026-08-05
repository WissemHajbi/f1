package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) ReplaceResults(ctx context.Context, season int, races []domain.RaceClassification) error {
	if len(races) == 0 {
		return fmt.Errorf("results cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin results transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM result_races WHERE season=?`, season); err != nil {
		return fmt.Errorf("clear results: %w", err)
	}
	raceStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO result_races
		(season, round, name, race_at, circuit_id, circuit_name, locality, country, latitude, longitude, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'jolpica', ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare result race: %w", err)
	}
	defer raceStatement.Close()
	resultStatement, err := tx.PrepareContext(ctx, `
		INSERT INTO race_results
		(season, round, position, position_text, points, grid_position, laps, status, result_time, time_millis,
		 driver_id, driver_number, driver_code, given_name, family_name, driver_nationality,
		 constructor_id, constructor_name, constructor_nationality, fastest_lap_rank, fastest_lap_number,
		 fastest_lap_time, fastest_lap_speed, fastest_lap_speed_units)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare race result: %w", err)
	}
	defer resultStatement.Close()
	for _, race := range races {
		if race.Season != season || race.RaceAt == nil || len(race.Results) == 0 {
			return fmt.Errorf("invalid classification for round %d", race.Round)
		}
		_, err := raceStatement.ExecContext(ctx, race.Season, race.Round, race.Name, race.RaceAt.UTC().Format(time.RFC3339Nano),
			race.Circuit.ID, race.Circuit.Name, race.Circuit.Locality, race.Circuit.Country, race.Circuit.Latitude,
			race.Circuit.Longitude, race.SyncedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert result race %d: %w", race.Round, err)
		}
		for _, result := range race.Results {
			var fastestRank, fastestNumber, fastestTime, fastestSpeed, fastestUnits any
			if result.FastestLap != nil {
				fastestRank, fastestNumber, fastestTime = result.FastestLap.Rank, result.FastestLap.Lap, result.FastestLap.Time
				fastestSpeed, fastestUnits = result.FastestLap.AverageSpeed, result.FastestLap.SpeedUnits
			}
			_, err := resultStatement.ExecContext(ctx, race.Season, race.Round, result.Position, result.PositionText,
				result.Points, result.Grid, result.Laps, result.Status, result.Time, result.TimeMillis,
				result.DriverID, result.DriverNumber, result.DriverCode, result.GivenName, result.FamilyName,
				result.Nationality, result.Constructor.ID, result.Constructor.Name, result.Constructor.Nationality,
				fastestRank, fastestNumber, fastestTime, fastestSpeed, fastestUnits)
			if err != nil {
				return fmt.Errorf("insert round %d driver %s result: %w", race.Round, result.DriverID, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit results: %w", err)
	}
	return nil
}

func (s *Store) Results(ctx context.Context, season int, round *int) ([]domain.RaceClassification, error) {
	query := `SELECT season, round, name, race_at, circuit_id, circuit_name, locality, country, latitude, longitude, synced_at FROM result_races WHERE season=?`
	args := []any{season}
	if round != nil {
		query += ` AND round=?`
		args = append(args, *round)
	}
	query += ` ORDER BY round`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query result races: %w", err)
	}
	defer rows.Close()
	races := []domain.RaceClassification{}
	for rows.Next() {
		race, err := scanClassification(rows.Scan)
		if err != nil {
			return nil, err
		}
		races = append(races, race)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate result races: %w", err)
	}
	if len(races) == 0 {
		return nil, ErrNotFound
	}
	if err := s.loadRaceResults(ctx, races); err != nil {
		return nil, err
	}
	return races, nil
}

func (s *Store) LatestResult(ctx context.Context, before time.Time) (domain.RaceClassification, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT season, round, name, race_at, circuit_id, circuit_name, locality, country, latitude, longitude, synced_at
		FROM result_races WHERE race_at<=? ORDER BY race_at DESC LIMIT 1
	`, before.UTC().Format(time.RFC3339Nano))
	race, err := scanClassification(row.Scan)
	if err != nil {
		return domain.RaceClassification{}, err
	}
	races := []domain.RaceClassification{race}
	if err := s.loadRaceResults(ctx, races); err != nil {
		return domain.RaceClassification{}, err
	}
	return races[0], nil
}

type classificationScan func(dest ...any) error

func scanClassification(scan classificationScan) (domain.RaceClassification, error) {
	var race domain.RaceClassification
	var raceAt, syncedAt string
	err := scan(&race.Season, &race.Round, &race.Name, &raceAt, &race.Circuit.ID, &race.Circuit.Name,
		&race.Circuit.Locality, &race.Circuit.Country, &race.Circuit.Latitude, &race.Circuit.Longitude, &syncedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RaceClassification{}, ErrNotFound
	}
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("scan classification: %w", err)
	}
	parsed, err := time.Parse(time.RFC3339Nano, raceAt)
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("parse result race time: %w", err)
	}
	race.RaceAt = &parsed
	race.SyncedAt, err = time.Parse(time.RFC3339Nano, syncedAt)
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("parse result sync time: %w", err)
	}
	return race, nil
}

func (s *Store) loadRaceResults(ctx context.Context, races []domain.RaceClassification) error {
	byRace := make(map[[2]int]*domain.RaceClassification, len(races))
	for index := range races {
		races[index].Results = []domain.RaceResult{}
		byRace[[2]int{races[index].Season, races[index].Round}] = &races[index]
	}
	minSeason, maxSeason := races[0].Season, races[0].Season
	for _, race := range races {
		if race.Season < minSeason {
			minSeason = race.Season
		}
		if race.Season > maxSeason {
			maxSeason = race.Season
		}
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT season, round, position, position_text, points, grid_position, laps, status, result_time,
		       time_millis, driver_id, driver_number, driver_code, given_name, family_name, driver_nationality,
		       constructor_id, constructor_name, constructor_nationality, fastest_lap_rank, fastest_lap_number,
		       fastest_lap_time, fastest_lap_speed, fastest_lap_speed_units
		FROM race_results WHERE season BETWEEN ? AND ? ORDER BY season, round, position
	`, minSeason, maxSeason)
	if err != nil {
		return fmt.Errorf("query race results: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var season, round int
		var result domain.RaceResult
		var millis sql.NullInt64
		var fastestRank, fastestNumber sql.NullInt64
		var fastestTime, fastestUnits sql.NullString
		var fastestSpeed sql.NullFloat64
		err := rows.Scan(&season, &round, &result.Position, &result.PositionText, &result.Points, &result.Grid,
			&result.Laps, &result.Status, &result.Time, &millis, &result.DriverID, &result.DriverNumber,
			&result.DriverCode, &result.GivenName, &result.FamilyName, &result.Nationality, &result.Constructor.ID,
			&result.Constructor.Name, &result.Constructor.Nationality, &fastestRank, &fastestNumber, &fastestTime,
			&fastestSpeed, &fastestUnits)
		if err != nil {
			return fmt.Errorf("scan race result: %w", err)
		}
		race := byRace[[2]int{season, round}]
		if race == nil {
			continue
		}
		if millis.Valid {
			result.TimeMillis = &millis.Int64
		}
		if fastestRank.Valid {
			result.FastestLap = &domain.FastestLap{Rank: int(fastestRank.Int64), Lap: int(fastestNumber.Int64),
				Time: fastestTime.String, AverageSpeed: fastestSpeed.Float64, SpeedUnits: fastestUnits.String}
		}
		race.Results = append(race.Results, result)
	}
	return rows.Err()
}
