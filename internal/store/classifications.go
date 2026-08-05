package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) ReplaceClassifications(ctx context.Context, season int, items []domain.SessionClassification) error {
	if len(items) == 0 {
		return fmt.Errorf("classifications cannot be empty")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin classifications transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM session_classifications WHERE season=? AND source='jolpica'`, season); err != nil {
		return fmt.Errorf("clear classifications: %w", err)
	}
	eventStmt, err := tx.PrepareContext(ctx, `INSERT INTO session_classifications
		(season, round, type, name, start_at, circuit_id, circuit_name, locality, country, latitude, longitude, source, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer eventStmt.Close()
	resultStmt, err := tx.PrepareContext(ctx, `INSERT INTO session_classification_results
		(season, round, type, position, driver_id, driver_number, driver_code, given_name, family_name,
		 nationality, constructor_id, constructor_name, constructor_nationality, q1, q2, q3, points,
		 grid_position, laps, status, result_time, time_millis, fastest_lap_rank, fastest_lap_number,
		 fastest_lap_time, fastest_lap_speed, fastest_lap_speed_units)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer resultStmt.Close()
	for _, item := range items {
		if item.Season != season || (item.Type != "qualifying" && item.Type != "sprint") || len(item.Results) == 0 {
			return fmt.Errorf("invalid %s classification round %d", item.Type, item.Round)
		}
		var start any
		if item.StartAt != nil {
			start = item.StartAt.UTC().Format(time.RFC3339Nano)
		}
		_, err := eventStmt.ExecContext(ctx, item.Season, item.Round, item.Type, item.Name, start,
			item.Circuit.ID, item.Circuit.Name, item.Circuit.Locality, item.Circuit.Country, item.Circuit.Latitude,
			item.Circuit.Longitude, item.Source, item.SyncedAt.UTC().Format(time.RFC3339Nano))
		if err != nil {
			return fmt.Errorf("insert %s round %d: %w", item.Type, item.Round, err)
		}
		for _, result := range item.Results {
			var fastestRank, fastestNumber, fastestTime, fastestSpeed, fastestUnits any
			if result.FastestLap != nil {
				fastestRank, fastestNumber, fastestTime = result.FastestLap.Rank, result.FastestLap.Lap, result.FastestLap.Time
				fastestSpeed, fastestUnits = result.FastestLap.AverageSpeed, result.FastestLap.SpeedUnits
			}
			_, err := resultStmt.ExecContext(ctx, item.Season, item.Round, item.Type, result.Position, result.DriverID,
				result.DriverNumber, result.DriverCode, result.GivenName, result.FamilyName, result.Nationality,
				result.Constructor.ID, result.Constructor.Name, result.Constructor.Nationality, result.Q1, result.Q2,
				result.Q3, result.Points, result.Grid, result.Laps, result.Status, result.Time, result.TimeMillis,
				fastestRank, fastestNumber, fastestTime, fastestSpeed, fastestUnits)
			if err != nil {
				return fmt.Errorf("insert %s driver %s: %w", item.Type, result.DriverID, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit classifications: %w", err)
	}
	return nil
}

func (s *Store) Classifications(ctx context.Context, season int, kind string, round *int) ([]domain.SessionClassification, error) {
	query := `SELECT season, round, type, name, start_at, circuit_id, circuit_name, locality, country,
	                 latitude, longitude, source, synced_at
	          FROM session_classifications WHERE season=? AND type=?`
	args := []any{season, kind}
	if round != nil {
		query += ` AND round=?`
		args = append(args, *round)
	}
	query += ` ORDER BY round`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query classifications: %w", err)
	}
	defer rows.Close()
	items := []domain.SessionClassification{}
	for rows.Next() {
		var item domain.SessionClassification
		var start sql.NullString
		var synced string
		if err := rows.Scan(&item.Season, &item.Round, &item.Type, &item.Name, &start, &item.Circuit.ID,
			&item.Circuit.Name, &item.Circuit.Locality, &item.Circuit.Country, &item.Circuit.Latitude,
			&item.Circuit.Longitude, &item.Source, &synced); err != nil {
			return nil, err
		}
		if start.Valid {
			parsed, err := time.Parse(time.RFC3339Nano, start.String)
			if err != nil {
				return nil, err
			}
			item.StartAt = &parsed
		}
		item.SyncedAt, err = time.Parse(time.RFC3339Nano, synced)
		if err != nil {
			return nil, err
		}
		item.Results = []domain.ClassificationResult{}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	if err := s.attachClassificationResults(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) attachClassificationResults(ctx context.Context, items []domain.SessionClassification) error {
	byKey := map[string]*domain.SessionClassification{}
	for i := range items {
		byKey[fmt.Sprintf("%d:%d:%s", items[i].Season, items[i].Round, items[i].Type)] = &items[i]
	}
	season, kind := items[0].Season, items[0].Type
	rows, err := s.db.QueryContext(ctx, `SELECT season, round, type, position, driver_id, driver_number,
		driver_code, given_name, family_name, nationality, constructor_id, constructor_name,
		constructor_nationality, q1, q2, q3, points, grid_position, laps, status, result_time,
		time_millis, fastest_lap_rank, fastest_lap_number, fastest_lap_time, fastest_lap_speed,
		fastest_lap_speed_units FROM session_classification_results
		WHERE season=? AND type=? ORDER BY round, position`, season, kind)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var result domain.ClassificationResult
		var itemSeason, round int
		var itemType string
		var points, fastestSpeed sql.NullFloat64
		var grid, laps, millis, fastestRank, fastestNumber sql.NullInt64
		var fastestTime, fastestUnits sql.NullString
		if err := rows.Scan(&itemSeason, &round, &itemType, &result.Position, &result.DriverID,
			&result.DriverNumber, &result.DriverCode, &result.GivenName, &result.FamilyName, &result.Nationality,
			&result.Constructor.ID, &result.Constructor.Name, &result.Constructor.Nationality, &result.Q1,
			&result.Q2, &result.Q3, &points, &grid, &laps, &result.Status, &result.Time, &millis,
			&fastestRank, &fastestNumber, &fastestTime, &fastestSpeed, &fastestUnits); err != nil {
			return err
		}
		if points.Valid {
			result.Points = &points.Float64
		}
		result.Grid = nullableInt(grid)
		result.Laps = nullableInt(laps)
		if millis.Valid {
			result.TimeMillis = &millis.Int64
		}
		if fastestRank.Valid {
			result.FastestLap = &domain.FastestLap{Rank: int(fastestRank.Int64), Lap: int(fastestNumber.Int64), Time: fastestTime.String, AverageSpeed: fastestSpeed.Float64, SpeedUnits: fastestUnits.String}
		}
		if item := byKey[fmt.Sprintf("%d:%d:%s", itemSeason, round, itemType)]; item != nil {
			item.Results = append(item.Results, result)
		}
	}
	return rows.Err()
}
