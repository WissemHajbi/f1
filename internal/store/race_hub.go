package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertRaceSessionLink(ctx context.Context, season, round, sessionKey int, linkedAt time.Time) error {
	if season < 1950 || round <= 0 || sessionKey <= 0 {
		return fmt.Errorf("season, round, and session key must be positive")
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO race_session_links (season, round, session_key, source, linked_at)
		SELECT ?, ?, ?, 'manual', ?
		WHERE EXISTS (SELECT 1 FROM events WHERE season=? AND round=?)
		  AND EXISTS (SELECT 1 FROM openf1_sessions WHERE session_key=? AND lower(type)='race')
		ON CONFLICT(season, round) DO UPDATE SET
			session_key=excluded.session_key, source=excluded.source, linked_at=excluded.linked_at`,
		season, round, sessionKey, linkedAt.UTC().Format(time.RFC3339Nano), season, round, sessionKey)
	if err != nil {
		return fmt.Errorf("upsert race-session link: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect race-session link: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RaceHub(ctx context.Context, season, round int, requestedDriver *int) (domain.RaceHub, error) {
	hub := domain.RaceHub{Season: season, Round: round, Drivers: []domain.RaceDriverStats{}, Laps: []domain.RaceLapStat{}, Track: []domain.TrackPoint{}}
	if err := s.db.QueryRowContext(ctx, `SELECT session_key FROM race_session_links WHERE season=? AND round=?`, season, round).Scan(&hub.SessionKey); err != nil {
		if err == sql.ErrNoRows {
			return domain.RaceHub{}, ErrNotFound
		}
		return domain.RaceHub{}, fmt.Errorf("query race-session link: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT d.session_key, d.driver_number, d.broadcast_name, d.full_name, d.acronym,
		       d.team_name, d.team_colour, d.first_name, d.last_name, d.country_code,
		       (SELECT COUNT(*) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT l.lap_number FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number
		          AND l.lap_duration IS NOT NULL ORDER BY l.lap_duration LIMIT 1),
		       (SELECT MIN(l.lap_duration) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT MIN(l.sector_1_duration) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT MIN(l.sector_2_duration) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT MIN(l.sector_3_duration) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT MAX(l.speed_trap) FROM laps l WHERE l.session_key=d.session_key AND l.driver_number=d.driver_number),
		       (SELECT COUNT(*) FROM pit_stops p WHERE p.session_key=d.session_key AND p.driver_number=d.driver_number),
		       (SELECT COUNT(*) FROM stints st WHERE st.session_key=d.session_key AND st.driver_number=d.driver_number),
		       (SELECT COUNT(*) FROM overtakes o WHERE o.session_key=d.session_key AND o.driver_number=d.driver_number),
		       (SELECT COUNT(*) FROM team_radio tr WHERE tr.session_key=d.session_key AND tr.driver_number=d.driver_number)
		FROM drivers d
		LEFT JOIN race_results rr ON rr.season=? AND rr.round=? AND CAST(rr.driver_number AS INTEGER)=d.driver_number
		WHERE d.session_key=?
		ORDER BY COALESCE(rr.position, 999), d.driver_number`, season, round, hub.SessionKey)
	if err != nil {
		return domain.RaceHub{}, fmt.Errorf("query race-hub drivers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var stats domain.RaceDriverStats
		var country sql.NullString
		var bestLap sql.NullInt64
		var bestDuration, sector1, sector2, sector3 sql.NullFloat64
		var topSpeed sql.NullInt64
		if err := rows.Scan(&stats.Driver.SessionKey, &stats.Driver.Number, &stats.Driver.BroadcastName,
			&stats.Driver.FullName, &stats.Driver.Acronym, &stats.Driver.TeamName, &stats.Driver.TeamColour,
			&stats.Driver.FirstName, &stats.Driver.LastName, &country,
			&stats.CompletedLaps, &bestLap, &bestDuration, &sector1, &sector2, &sector3, &topSpeed,
			&stats.PitStops, &stats.Stints, &stats.Overtakes, &stats.RadioMessages); err != nil {
			return domain.RaceHub{}, fmt.Errorf("scan race-hub driver: %w", err)
		}
		if country.Valid {
			stats.Driver.CountryCode = &country.String
		}
		stats.BestLapNumber = nullableInt(bestLap)
		stats.BestLapDuration = nullableFloat(bestDuration)
		stats.BestSector1 = nullableFloat(sector1)
		stats.BestSector2 = nullableFloat(sector2)
		stats.BestSector3 = nullableFloat(sector3)
		stats.TopSpeed = nullableInt(topSpeed)
		hub.Drivers = append(hub.Drivers, stats)
	}
	if err := rows.Err(); err != nil {
		return domain.RaceHub{}, fmt.Errorf("iterate race-hub drivers: %w", err)
	}
	if len(hub.Drivers) == 0 {
		return domain.RaceHub{}, ErrNotFound
	}

	hub.SelectedDriverNumber = hub.Drivers[0].Driver.Number
	if requestedDriver != nil {
		hub.SelectedDriverNumber = *requestedDriver
		found := false
		for _, driver := range hub.Drivers {
			if driver.Driver.Number == *requestedDriver {
				found = true
				break
			}
		}
		if !found {
			return domain.RaceHub{}, ErrNotFound
		}
	}
	if err := s.loadRaceHubLaps(ctx, &hub); err != nil {
		return domain.RaceHub{}, err
	}
	if err := s.loadRaceHubTrack(ctx, &hub); err != nil {
		return domain.RaceHub{}, err
	}
	return hub, nil
}

func (s *Store) loadRaceHubLaps(ctx context.Context, hub *domain.RaceHub) error {
	rows, err := s.db.QueryContext(ctx, `SELECT lap_number, lap_duration, sector_1_duration, sector_2_duration,
		sector_3_duration, speed_trap, is_pit_out_lap FROM laps
		WHERE session_key=? AND driver_number=? ORDER BY lap_number`, hub.SessionKey, hub.SelectedDriverNumber)
	if err != nil {
		return fmt.Errorf("query race-hub laps: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var lap domain.RaceLapStat
		var duration, sector1, sector2, sector3 sql.NullFloat64
		var speed sql.NullInt64
		if err := rows.Scan(&lap.LapNumber, &duration, &sector1, &sector2, &sector3, &speed, &lap.IsPitOutLap); err != nil {
			return fmt.Errorf("scan race-hub lap: %w", err)
		}
		lap.Duration = nullableFloat(duration)
		lap.Sector1Duration = nullableFloat(sector1)
		lap.Sector2Duration = nullableFloat(sector2)
		lap.Sector3Duration = nullableFloat(sector3)
		lap.SpeedTrap = nullableInt(speed)
		hub.Laps = append(hub.Laps, lap)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate race-hub laps: %w", err)
	}
	return nil
}

func (s *Store) loadRaceHubTrack(ctx context.Context, hub *domain.RaceHub) error {
	var driver, lap int
	var start string
	var duration float64
	err := s.db.QueryRowContext(ctx, `
		SELECT l.driver_number, l.lap_number, l.date_start, l.lap_duration
		FROM laps l JOIN location_samples p ON p.session_key=l.session_key AND p.driver_number=l.driver_number
		 AND julianday(p.sampled_at)>=julianday(l.date_start)
		 AND julianday(p.sampled_at)<=julianday(l.date_start)+(l.lap_duration/86400.0)
		WHERE l.session_key=? AND l.date_start IS NOT NULL AND l.lap_duration IS NOT NULL
		GROUP BY l.driver_number, l.lap_number
		ORDER BY COUNT(*) DESC LIMIT 1`, hub.SessionKey).Scan(&driver, &lap, &start, &duration)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("select race-hub track lap: %w", err)
	}
	startedAt, err := time.Parse(telemetryTimeLayout, start)
	if err != nil {
		return fmt.Errorf("parse race-hub track start: %w", err)
	}
	endedAt := startedAt.Add(time.Duration(duration * float64(time.Second)))
	rows, err := s.db.QueryContext(ctx, `SELECT x, y FROM location_samples
		WHERE session_key=? AND driver_number=? AND sampled_at>=? AND sampled_at<=? ORDER BY sampled_at`,
		hub.SessionKey, driver, startedAt.Format(telemetryTimeLayout), endedAt.Format(telemetryTimeLayout))
	if err != nil {
		return fmt.Errorf("query race-hub track: %w", err)
	}
	defer rows.Close()
	points := []domain.TrackPoint{}
	for rows.Next() {
		var point domain.TrackPoint
		if err := rows.Scan(&point.X, &point.Y); err != nil {
			return fmt.Errorf("scan race-hub track point: %w", err)
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate race-hub track: %w", err)
	}
	hub.Track = downsampleTrack(points, 220)
	hub.TrackSourceDriver = &driver
	hub.TrackSourceLap = &lap
	return nil
}

func downsampleTrack(points []domain.TrackPoint, maximum int) []domain.TrackPoint {
	if len(points) <= maximum {
		return points
	}
	result := make([]domain.TrackPoint, 0, maximum)
	for i := 0; i < maximum; i++ {
		index := i * (len(points) - 1) / (maximum - 1)
		result = append(result, points[index])
	}
	return result
}
