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
