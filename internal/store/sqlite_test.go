package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"oidysts/internal/domain"
)

func TestSaveAndLatest(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	item := Snapshot{Source: "test", Resource: "sample", Endpoint: "https://example.test", FetchedAt: time.Now(),
		StatusCode: 200, ContentType: "application/json", RecordCount: 2, PayloadBytes: 2,
		Summary: json.RawMessage(`{"records":2}`), Payload: []byte(`[]`)}
	if err := db.Save(ctx, item); err != nil {
		t.Fatal(err)
	}
	items, err := db.Latest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RecordCount != 2 {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestReplaceAndReadCalendar(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Truncate(time.Second)
	raceAt := now.Add(24 * time.Hour)
	practiceAt := now.Add(12 * time.Hour)
	events := []domain.Event{{
		Season: 2025, Round: 1, Name: "Test Grand Prix", SourceURL: "https://example.test/race", RaceAt: &raceAt,
		Circuit:  domain.Circuit{ID: "test", Name: "Test Circuit", Locality: "City", Country: "Country", Latitude: 1.2, Longitude: 3.4},
		Sessions: []domain.EventSession{{Type: "practice_1", StartAt: &practiceAt}, {Type: "race", StartAt: &raceAt}}, SyncedAt: now,
	}}
	if err := db.ReplaceCalendar(context.Background(), 2025, events); err != nil {
		t.Fatal(err)
	}
	calendar, err := db.Calendar(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(calendar) != 1 || len(calendar[0].Sessions) != 2 {
		t.Fatalf("unexpected calendar: %+v", calendar)
	}
	next, err := db.NextEvent(context.Background(), now)
	if err != nil || next.Round != 1 {
		t.Fatalf("next=%+v err=%v", next, err)
	}
}

func TestSaveAndReadDriverRoster(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	country := "GBR"
	roster := domain.DriverRoster{
		Session:  domain.Session{Key: 10, Year: 2025, Name: "Race", Type: "Race", Location: "Test", DateStart: now.Add(-2 * time.Hour), DateEnd: now},
		Drivers:  []domain.Driver{{SessionKey: 10, Number: 4, FullName: "Test Driver", Acronym: "TST", TeamName: "Test Team", CountryCode: &country}},
		SyncedAt: now,
	}
	if err := db.SaveDriverRoster(context.Background(), roster); err != nil {
		t.Fatal(err)
	}
	got, err := db.LatestDriverRoster(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if got.Session.Key != 10 || len(got.Drivers) != 1 || got.Drivers[0].FullName != "Test Driver" {
		t.Fatalf("unexpected roster: %+v", got)
	}
}
