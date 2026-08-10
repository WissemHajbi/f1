package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"oidysts/internal/domain"
)

func TestRaceSessionLinkSurvivesCalendarRefresh(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	raceAt := now.Add(time.Hour)
	events := []domain.Event{{Season: 2025, Round: 24, Name: "Test Grand Prix", SourceURL: "test",
		RaceAt: &raceAt, Circuit: domain.Circuit{ID: "test", Name: "Test", Locality: "City", Country: "Country"}, SyncedAt: now}}
	if err := db.ReplaceCalendar(ctx, 2025, events); err != nil {
		t.Fatal(err)
	}
	_, err = db.db.ExecContext(ctx, `INSERT INTO openf1_meetings
		(meeting_key,year,name,official_name,location,country_key,country_code,country_name,circuit_key,circuit_short_name,date_start,gmt_offset,source,synced_at)
		VALUES (1276,2025,'Test','Test','City',1,'TST','Country',2,'Test',?, '+00:00','openf1',?)`, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.db.ExecContext(ctx, `INSERT INTO openf1_sessions
		(session_key,meeting_key,year,name,type,location,country_code,country_name,circuit_key,circuit_short_name,date_start,date_end,gmt_offset,is_cancelled,source,synced_at)
		VALUES (9839,1276,2025,'Race','Race','City','TST','Country',2,'Test',?,?, '+00:00',0,'openf1',?)`,
		now.Format(time.RFC3339Nano), raceAt.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertRaceSessionLink(ctx, 2025, 24, 9839, now); err != nil {
		t.Fatal(err)
	}
	events[0].Name = "Updated Grand Prix"
	if err := db.ReplaceCalendar(ctx, 2025, events); err != nil {
		t.Fatal(err)
	}
	var sessionKey int
	if err := db.db.QueryRowContext(ctx, `SELECT session_key FROM race_session_links WHERE season=2025 AND round=24`).Scan(&sessionKey); err != nil {
		t.Fatal(err)
	}
	if sessionKey != 9839 {
		t.Fatalf("session key=%d", sessionKey)
	}
}

func TestRaceSessionLinkRejectsUnknownEntities(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.UpsertRaceSessionLink(context.Background(), 2025, 24, 9839, time.Now())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
