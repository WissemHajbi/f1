package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type Snapshot struct {
	Source       string          `json:"source"`
	Resource     string          `json:"resource"`
	Endpoint     string          `json:"endpoint"`
	FetchedAt    time.Time       `json:"fetched_at"`
	StatusCode   int             `json:"status_code"`
	ContentType  string          `json:"content_type"`
	RecordCount  int             `json:"record_count"`
	PayloadBytes int             `json:"payload_bytes"`
	Summary      json.RawMessage `json:"summary,omitempty"`
	Payload      []byte          `json:"-"`
}

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
		path = filepath.ToSlash(path)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate sqlite: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Save(ctx context.Context, snapshot Snapshot) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO source_snapshots
		(source, resource, endpoint, fetched_at, status_code, content_type, record_count, payload_bytes, summary, payload)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, snapshot.Source, snapshot.Resource, snapshot.Endpoint, snapshot.FetchedAt.UTC().Format(time.RFC3339Nano),
		snapshot.StatusCode, snapshot.ContentType, snapshot.RecordCount, snapshot.PayloadBytes, []byte(snapshot.Summary), snapshot.Payload)
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func (s *Store) Latest(ctx context.Context) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT source, resource, endpoint, fetched_at, status_code, content_type, record_count, payload_bytes, summary
		FROM source_snapshots s
		WHERE id = (SELECT id FROM source_snapshots WHERE source=s.source AND resource=s.resource ORDER BY fetched_at DESC, id DESC LIMIT 1)
		ORDER BY source, resource
	`)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()
	var snapshots []Snapshot
	for rows.Next() {
		var item Snapshot
		var fetched string
		if err := rows.Scan(&item.Source, &item.Resource, &item.Endpoint, &fetched, &item.StatusCode, &item.ContentType,
			&item.RecordCount, &item.PayloadBytes, &item.Summary); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		item.FetchedAt, _ = time.Parse(time.RFC3339Nano, fetched)
		snapshots = append(snapshots, item)
	}
	return snapshots, rows.Err()
}

func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
