package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"oidysts/internal/domain"
)

func (s *Store) UpsertCircuitGeometry(ctx context.Context, season, round int, geometry domain.CircuitGeometry, syncedAt time.Time) error {
	if season < 1950 || round <= 0 || geometry.SessionKey <= 0 || len(geometry.Points) < 3 || geometry.EstimatedWidthM <= 0 || geometry.Attribution == "" {
		return fmt.Errorf("invalid circuit geometry")
	}
	points, err := json.Marshal(geometry.Points)
	if err != nil {
		return fmt.Errorf("encode circuit geometry: %w", err)
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO circuit_geometries
		(season,round,session_key,points,estimated_width_m,attribution,source_url,geometry_accuracy,source,synced_at)
		SELECT ?,?,?,?,?,?,?,?,'openstreetmap',?
		WHERE EXISTS (SELECT 1 FROM race_session_links WHERE season=? AND round=? AND session_key=?)
		ON CONFLICT(season,round) DO UPDATE SET session_key=excluded.session_key, points=excluded.points,
			estimated_width_m=excluded.estimated_width_m, attribution=excluded.attribution,
			source_url=excluded.source_url, geometry_accuracy=excluded.geometry_accuracy, synced_at=excluded.synced_at`,
		season, round, geometry.SessionKey, points, geometry.EstimatedWidthM, geometry.Attribution,
		geometry.SourceURL, geometry.GeometryAccuracy, syncedAt.UTC().Format(time.RFC3339Nano),
		season, round, geometry.SessionKey)
	if err != nil {
		return fmt.Errorf("upsert circuit geometry: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect circuit geometry upsert: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) loadCircuitGeometry(ctx context.Context, hub *domain.RaceHub) error {
	var points []byte
	var width float64
	var attribution, sourceURL, accuracy string
	err := s.db.QueryRowContext(ctx, `SELECT points,estimated_width_m,attribution,source_url,geometry_accuracy
		FROM circuit_geometries WHERE season=? AND round=? AND session_key=?`, hub.Season, hub.Round, hub.SessionKey).
		Scan(&points, &width, &attribution, &sourceURL, &accuracy)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("query circuit geometry: %w", err)
	}
	if err := json.Unmarshal(points, &hub.Track); err != nil {
		return fmt.Errorf("decode circuit geometry: %w", err)
	}
	hub.TrackSourceDriver = nil
	hub.TrackSourceLap = nil
	hub.TrackEstimatedWidthM = &width
	hub.TrackAttribution = attribution
	hub.TrackSourceURL = sourceURL
	hub.TrackAccuracy = accuracy
	return nil
}
