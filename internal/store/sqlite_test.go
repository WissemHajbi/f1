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

func TestReplaceAndReadStandings(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	drivers := []domain.DriverStanding{{Season: 2025, Round: 2, Position: 1, Points: 44.5, Wins: 1,
		DriverID: "test", GivenName: "Test", FamilyName: "Driver", Constructors: []domain.StandingConstructor{{ID: "team", Name: "Team"}}, SyncedAt: now}}
	constructors := []domain.ConstructorStanding{{Season: 2025, Round: 2, Position: 1, Points: 80, Wins: 2,
		Constructor: domain.StandingConstructor{ID: "team", Name: "Team"}, SyncedAt: now}}
	if err := db.ReplaceStandings(context.Background(), 2025, drivers, constructors); err != nil {
		t.Fatal(err)
	}
	gotDrivers, err := db.DriverStandings(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	gotConstructors, err := db.ConstructorStandings(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotDrivers) != 1 || len(gotDrivers[0].Constructors) != 1 || len(gotConstructors) != 1 {
		t.Fatalf("drivers=%+v constructors=%+v", gotDrivers, gotConstructors)
	}
}

func TestReplaceAndReadResults(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Truncate(time.Second)
	millis := int64(5_400_000)
	races := []domain.RaceClassification{{Season: 2025, Round: 1, Name: "Test Grand Prix", RaceAt: &now,
		Circuit: domain.Circuit{ID: "test", Name: "Test Circuit", Locality: "City", Country: "Country"}, SyncedAt: now,
		Results: []domain.RaceResult{{Position: 1, PositionText: "1", Points: 25, Grid: 2, Laps: 50,
			Status: "Finished", TimeMillis: &millis, DriverID: "test", GivenName: "Test", FamilyName: "Driver",
			Constructor: domain.StandingConstructor{ID: "team", Name: "Team"},
			FastestLap:  &domain.FastestLap{Rank: 1, Lap: 42, Time: "1:20.000", AverageSpeed: 210.5, SpeedUnits: "kph"}}}}}
	if err := db.ReplaceResults(context.Background(), 2025, races); err != nil {
		t.Fatal(err)
	}
	round := 1
	got, err := db.Results(context.Background(), 2025, &round)
	if err != nil {
		t.Fatal(err)
	}
	latest, err := db.LatestResult(context.Background(), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Results) != 1 || got[0].Results[0].FastestLap == nil || latest.Round != 1 {
		t.Fatalf("results=%+v latest=%+v", got, latest)
	}
}

func TestReplaceAndReadMeetings(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Truncate(time.Second)
	meetings := []domain.Meeting{{Key: 100, Year: 2025, Name: "Test Grand Prix", OfficialName: "Official Test",
		Location: "Test City", CountryKey: 1, CountryCode: "TST", CountryName: "Test", CircuitKey: 2,
		CircuitShortName: "Test Circuit", DateStart: now, GMTOffset: "00:00:00", SyncedAt: now,
		Sessions: []domain.MeetingSession{{Key: 200, MeetingKey: 100, Year: 2025, Name: "Race", Type: "Race",
			Location: "Test City", DateStart: now, DateEnd: now.Add(2 * time.Hour)}}}}
	if err := db.ReplaceMeetings(context.Background(), 2025, meetings); err != nil {
		t.Fatal(err)
	}
	got, err := db.Meetings(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := db.Sessions(context.Background(), 2025, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0].Sessions) != 1 || len(sessions) != 1 {
		t.Fatalf("meetings=%+v sessions=%+v", got, sessions)
	}
}

func TestUpsertAndReadCarData(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Truncate(time.Second)
	meetings := []domain.Meeting{{Key: 100, Year: 2025, Name: "Test", DateStart: now, SyncedAt: now,
		Sessions: []domain.MeetingSession{{Key: 200, MeetingKey: 100, Year: 2025, Name: "Race", Type: "Race", DateStart: now, DateEnd: now.Add(time.Hour)}}}}
	if err := db.ReplaceMeetings(context.Background(), 2025, meetings); err != nil {
		t.Fatal(err)
	}
	samples := []domain.CarDataSample{
		{Timestamp: now.Add(250 * time.Millisecond), SessionKey: 200, MeetingKey: 100, DriverNumber: 4, Speed: 250, RPM: 11000, Gear: 7, Throttle: 100},
		{Timestamp: now.Add(500 * time.Millisecond), SessionKey: 200, MeetingKey: 100, DriverNumber: 4, Speed: 255, RPM: 11200, Gear: 8, Throttle: 100},
	}
	if err := db.UpsertCarData(context.Background(), samples, now); err != nil {
		t.Fatal(err)
	}
	got, truncated, err := db.CarData(context.Background(), domain.CarDataQuery{SessionKey: 200, DriverNumber: 4, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !truncated || got[0].Speed != 250 {
		t.Fatalf("samples=%+v truncated=%v", got, truncated)
	}
}

func TestUpsertAndReadLaps(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC().Truncate(time.Second)
	meetings := []domain.Meeting{{Key: 100, Year: 2025, Name: "Test", DateStart: now, SyncedAt: now,
		Sessions: []domain.MeetingSession{{Key: 200, MeetingKey: 100, Year: 2025, Name: "Race", Type: "Race", DateStart: now, DateEnd: now.Add(time.Hour)}}}}
	if err := db.ReplaceMeetings(context.Background(), 2025, meetings); err != nil {
		t.Fatal(err)
	}
	duration, sector, speed := 90.5, 30.1, 310
	laps := []domain.Lap{{SessionKey: 200, MeetingKey: 100, DriverNumber: 4, LapNumber: 1,
		DateStart: &now, LapDuration: &duration, Sector1Duration: &sector, SpeedTrap: &speed,
		Sector1Segments: []int{2049, 2051}, Sector2Segments: []int{}, Sector3Segments: []int{2051}}}
	if err := db.UpsertLaps(context.Background(), laps, now); err != nil {
		t.Fatal(err)
	}
	driver := 4
	got, truncated, err := db.Laps(context.Background(), domain.LapQuery{SessionKey: 200, DriverNumber: &driver, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if truncated || len(got) != 1 || got[0].LapDuration == nil || len(got[0].Sector1Segments) != 2 {
		t.Fatalf("laps=%+v truncated=%v", got, truncated)
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
