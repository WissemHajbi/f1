package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"oidysts/internal/config"
	"oidysts/internal/domain"
	"oidysts/internal/fetch"
	"oidysts/internal/jolpica"
	"oidysts/internal/openf1"
	"oidysts/internal/store"
)

func main() {
	var resource, fromValue, toValue string
	var year, round, sessionKey, driverNumber, lapNumber int
	flag.StringVar(&resource, "resource", "drivers", "resource to sync: drivers, meetings, calendar, standings, results, classifications, race-link, laps, best-lap-location, lap-location, stints, pit, weather, race-control, overtakes, positions, intervals, location, team-radio, or car-data")
	flag.IntVar(&year, "year", 2025, "season to sync")
	flag.IntVar(&round, "round", 0, "Jolpica championship round for race-link")
	flag.IntVar(&sessionKey, "session", 0, "OpenF1 session key for bounded resources")
	flag.IntVar(&driverNumber, "driver", 0, "driver number for bounded resources")
	flag.IntVar(&lapNumber, "lap", 0, "lap number for lap-location")
	flag.StringVar(&fromValue, "from", "", "inclusive RFC3339 range start")
	flag.StringVar(&toValue, "to", "", "exclusive RFC3339 range end")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if resource != "drivers" && resource != "meetings" && resource != "calendar" && resource != "standings" && resource != "results" && resource != "classifications" && resource != "race-link" && resource != "laps" && resource != "best-lap-location" && resource != "lap-location" && resource != "stints" && resource != "pit" && resource != "weather" && resource != "race-control" && resource != "overtakes" && resource != "positions" && resource != "intervals" && resource != "location" && resource != "team-radio" && resource != "car-data" {
		logger.Error("unsupported resource", "resource", resource)
		os.Exit(2)
	}
	cfg := config.Load()
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	upstream := fetch.New(cfg.Timeout, cfg.UserAgent, cfg.MaxBody)

	switch resource {
	case "drivers":
		err = syncDrivers(ctx, db, upstream, year)
	case "meetings":
		err = syncMeetings(ctx, db, upstream, year)
	case "calendar":
		err = syncCalendar(ctx, db, upstream, year)
	case "standings":
		err = syncStandings(ctx, db, upstream, year)
	case "results":
		err = syncResults(ctx, db, upstream, year)
	case "classifications":
		err = syncClassifications(ctx, db, upstream, year)
	case "race-link":
		err = syncRaceLink(ctx, db, year, round, sessionKey)
	case "laps":
		err = syncLaps(ctx, db, upstream, sessionKey, driverNumber)
	case "best-lap-location":
		err = syncBestLapLocation(ctx, db, upstream, sessionKey, driverNumber)
	case "lap-location":
		err = syncLapLocation(ctx, db, upstream, sessionKey, driverNumber, lapNumber)
	case "stints":
		err = syncStints(ctx, db, upstream, sessionKey, driverNumber)
	case "pit":
		err = syncPitStops(ctx, db, upstream, sessionKey, driverNumber)
	case "weather":
		err = syncWeather(ctx, db, upstream, sessionKey)
	case "race-control":
		err = syncRaceControl(ctx, db, upstream, sessionKey)
	case "overtakes":
		err = syncOvertakes(ctx, db, upstream, sessionKey, driverNumber)
	case "positions":
		err = syncPositions(ctx, db, upstream, sessionKey, driverNumber)
	case "intervals":
		err = syncIntervals(ctx, db, upstream, sessionKey, driverNumber)
	case "location":
		err = syncLocation(ctx, db, upstream, sessionKey, driverNumber, fromValue, toValue)
	case "team-radio":
		err = syncTeamRadio(ctx, db, upstream, sessionKey, driverNumber)
	case "car-data":
		err = syncCarData(ctx, db, upstream, sessionKey, driverNumber, fromValue, toValue)
	}
	if err != nil {
		logger.Error("sync failed", "resource", resource, "year", year, "error", err)
		os.Exit(1)
	}
}

func syncRaceLink(ctx context.Context, db *store.Store, season, round, sessionKey int) error {
	if err := db.UpsertRaceSessionLink(ctx, season, round, sessionKey, time.Now().UTC()); err != nil {
		return err
	}
	fmt.Printf("SYNC race-link season=%d round=%d session=%d\n", season, round, sessionKey)
	return nil
}

func syncOvertakes(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	driver := optionalDriver(driverNumber)
	result, err := openf1.New(upstream).Overtakes(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertOvertakes(ctx, result.Events, syncedAt); err != nil {
		return err
	}
	if err := saveTimelineSnapshot(ctx, db, "overtakes_sync", result.Endpoint, result.Payload, len(result.Events), sessionKey, driver, syncedAt); err != nil {
		return err
	}
	fmt.Printf("SYNC overtakes session=%d driver=%s records=%d\n", sessionKey, driverLabel(driver), len(result.Events))
	return nil
}

func syncPositions(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	driver := optionalDriver(driverNumber)
	result, err := openf1.New(upstream).Positions(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertPositions(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := saveTimelineSnapshot(ctx, db, "positions_sync", result.Endpoint, result.Payload, len(result.Samples), sessionKey, driver, syncedAt); err != nil {
		return err
	}
	fmt.Printf("SYNC positions session=%d driver=%s records=%d\n", sessionKey, driverLabel(driver), len(result.Samples))
	return nil
}

func syncIntervals(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	driver := optionalDriver(driverNumber)
	result, err := openf1.New(upstream).Intervals(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertIntervals(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := saveTimelineSnapshot(ctx, db, "intervals_sync", result.Endpoint, result.Payload, len(result.Samples), sessionKey, driver, syncedAt); err != nil {
		return err
	}
	fmt.Printf("SYNC intervals session=%d driver=%s records=%d\n", sessionKey, driverLabel(driver), len(result.Samples))
	return nil
}

func optionalDriver(number int) *int {
	if number <= 0 {
		return nil
	}
	return &number
}

func driverLabel(driver *int) string {
	if driver == nil {
		return "all"
	}
	return fmt.Sprintf("%d", *driver)
}

func saveTimelineSnapshot(ctx context.Context, db *store.Store, resource, endpoint string, payload []byte, records, sessionKey int, driver *int, syncedAt time.Time) error {
	return db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: resource, Endpoint: endpoint, FetchedAt: syncedAt,
		StatusCode: 200, ContentType: "application/json", RecordCount: records, PayloadBytes: len(payload),
		Summary: mustJSON(map[string]any{"session_key": sessionKey, "driver_number": driver, "records": records}), Payload: payload,
	})
}

func syncRaceControl(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey int) error {
	result, err := openf1.New(upstream).RaceControl(ctx, sessionKey)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertRaceControl(ctx, result.Events, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "race_control_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Events),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key": sessionKey,
			"events":      len(result.Events),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC race-control session=%d records=%d\n", sessionKey, len(result.Events))
	return nil
}

func syncWeather(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey int) error {
	result, err := openf1.New(upstream).Weather(ctx, sessionKey)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertWeather(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "weather_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Samples),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key": sessionKey,
			"samples":     len(result.Samples),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC weather session=%d records=%d\n", sessionKey, len(result.Samples))
	return nil
}

func syncPitStops(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	var driver *int
	if driverNumber > 0 {
		driver = &driverNumber
	}
	result, err := openf1.New(upstream).PitStops(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertPitStops(ctx, result.PitStops, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "pit_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.PitStops),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driver, "pit_stops": len(result.PitStops),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	label := "all"
	if driver != nil {
		label = fmt.Sprintf("%d", *driver)
	}
	fmt.Printf("SYNC pit session=%d driver=%s records=%d\n", sessionKey, label, len(result.PitStops))
	return nil
}

func syncStints(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	var driver *int
	if driverNumber > 0 {
		driver = &driverNumber
	}
	result, err := openf1.New(upstream).Stints(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertStints(ctx, result.Stints, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "stints_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Stints),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driver, "stints": len(result.Stints),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	label := "all"
	if driver != nil {
		label = fmt.Sprintf("%d", *driver)
	}
	fmt.Printf("SYNC stints session=%d driver=%s records=%d\n", sessionKey, label, len(result.Stints))
	return nil
}

func syncLaps(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	var driver *int
	if driverNumber > 0 {
		driver = &driverNumber
	}
	result, err := openf1.New(upstream).Laps(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertLaps(ctx, result.Laps, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "laps_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Laps),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driver, "laps": len(result.Laps),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	label := "all"
	if driver != nil {
		label = fmt.Sprintf("%d", *driver)
	}
	fmt.Printf("SYNC laps session=%d driver=%s records=%d\n", sessionKey, label, len(result.Laps))
	return nil
}

func syncTeamRadio(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	driver := optionalDriver(driverNumber)
	result, err := openf1.New(upstream).TeamRadio(ctx, sessionKey, driver)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	ticker := time.NewTicker(350 * time.Millisecond)
	defer ticker.Stop()
	downloaded, cached := 0, 0
	for _, record := range result.Records {
		exists, err := db.TeamRadioExists(ctx, record.ID)
		if err != nil {
			return err
		}
		if exists {
			cached++
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		audio, err := upstream.Get(ctx, record.RecordingSource, "audio/*")
		if err != nil {
			return fmt.Errorf("download team radio %s: %w", record.ID, err)
		}
		record.Audio, record.ContentType = audio.Body, audio.ContentType
		if record.ContentType == "" {
			record.ContentType = "application/octet-stream"
		}
		if err := db.UpsertTeamRadio(ctx, []domain.TeamRadio{record}, syncedAt); err != nil {
			return err
		}
		downloaded++
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "team_radio_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Records),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driver, "records": len(result.Records), "downloaded": downloaded, "cached": cached,
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC team-radio session=%d driver=%s records=%d downloaded=%d cached=%d\n",
		sessionKey, driverLabel(driver), len(result.Records), downloaded, cached)
	return nil
}

func syncBestLapLocation(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int) error {
	if sessionKey <= 0 || driverNumber <= 0 {
		return fmt.Errorf("best-lap-location requires positive -session and -driver")
	}
	lap, from, to, err := db.BestLapWindow(ctx, sessionKey, driverNumber)
	if err != nil {
		return err
	}
	result, err := openf1.New(upstream).Location(ctx, sessionKey, driverNumber, from, to)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertLocation(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "best_lap_location_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Samples),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key": sessionKey, "driver_number": driverNumber, "lap_number": lap,
			"from": from, "to": to, "samples": len(result.Samples),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC best-lap-location session=%d driver=%d lap=%d samples=%d\n", sessionKey, driverNumber, lap, len(result.Samples))
	return nil
}

func syncLapLocation(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber, lapNumber int) error {
	if sessionKey <= 0 || driverNumber <= 0 || lapNumber <= 0 {
		return fmt.Errorf("lap-location requires positive -session, -driver, and -lap")
	}
	from, to, err := db.LapWindow(ctx, sessionKey, driverNumber, lapNumber)
	if err != nil {
		return err
	}
	result, err := openf1.New(upstream).Location(ctx, sessionKey, driverNumber, from, to)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertLocation(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "lap_location_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Samples),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key": sessionKey, "driver_number": driverNumber, "lap_number": lapNumber,
			"from": from, "to": to, "samples": len(result.Samples),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC lap-location session=%d driver=%d lap=%d samples=%d\n", sessionKey, driverNumber, lapNumber, len(result.Samples))
	return nil
}

func syncLocation(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int, fromValue, toValue string) error {
	from, err := time.Parse(time.RFC3339Nano, fromValue)
	if err != nil {
		return fmt.Errorf("invalid -from RFC3339 timestamp: %w", err)
	}
	to, err := time.Parse(time.RFC3339Nano, toValue)
	if err != nil {
		return fmt.Errorf("invalid -to RFC3339 timestamp: %w", err)
	}
	result, err := openf1.New(upstream).Location(ctx, sessionKey, driverNumber, from, to)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertLocation(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "location_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Samples),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driverNumber, "from": from, "to": to, "samples": len(result.Samples),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC location session=%d driver=%d samples=%d from=%s to=%s\n", sessionKey, driverNumber,
		len(result.Samples), from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))
	return nil
}

func syncCarData(ctx context.Context, db *store.Store, upstream *fetch.Client, sessionKey, driverNumber int, fromValue, toValue string) error {
	from, err := time.Parse(time.RFC3339Nano, fromValue)
	if err != nil {
		return fmt.Errorf("invalid -from RFC3339 timestamp: %w", err)
	}
	to, err := time.Parse(time.RFC3339Nano, toValue)
	if err != nil {
		return fmt.Errorf("invalid -to RFC3339 timestamp: %w", err)
	}
	result, err := openf1.New(upstream).CarData(ctx, sessionKey, driverNumber, from, to)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()
	if err := db.UpsertCarData(ctx, result.Samples, syncedAt); err != nil {
		return err
	}
	if err := db.Save(ctx, store.Snapshot{
		Source: "openf1", Resource: "car_data_sync", Endpoint: result.Endpoint,
		FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Samples),
		PayloadBytes: len(result.Payload), Summary: mustJSON(map[string]any{
			"session_key":   sessionKey,
			"driver_number": driverNumber, "from": from, "to": to, "samples": len(result.Samples),
		}), Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC car-data session=%d driver=%d samples=%d from=%s to=%s\n", sessionKey, driverNumber,
		len(result.Samples), from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))
	return nil
}

func syncMeetings(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := openf1.New(upstream).Meetings(ctx, year)
	if err != nil {
		return err
	}
	if err := db.ReplaceMeetings(ctx, year, result.Meetings); err != nil {
		return err
	}
	sessions := 0
	for _, meeting := range result.Meetings {
		sessions += len(meeting.Sessions)
	}
	syncedAt := result.Meetings[0].SyncedAt
	for _, snapshot := range []store.Snapshot{
		{
			Source: "openf1", Resource: "meetings_sync", Endpoint: result.MeetingsURL, FetchedAt: syncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Meetings), PayloadBytes: len(result.MeetingsPayload),
			Summary: mustJSON(map[string]any{"season": year, "meetings": len(result.Meetings)}), Payload: result.MeetingsPayload,
		},
		{
			Source: "openf1", Resource: "all_sessions_sync", Endpoint: result.SessionsURL, FetchedAt: syncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: sessions, PayloadBytes: len(result.SessionsPayload),
			Summary: mustJSON(map[string]any{"season": year, "sessions": sessions}), Payload: result.SessionsPayload,
		},
	} {
		if err := db.Save(ctx, snapshot); err != nil {
			return err
		}
	}
	fmt.Printf("SYNC meetings season=%d meetings=%d sessions=%d\n", year, len(result.Meetings), sessions)
	return nil
}

func syncDrivers(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := openf1.New(upstream).LatestCompletedRaceRoster(ctx, year)
	if err != nil {
		return err
	}
	if err := db.SaveDriverRoster(ctx, result.Roster); err != nil {
		return err
	}
	summary, _ := json.Marshal(map[string]any{
		"season": year, "session_key": result.Roster.Session.Key, "drivers": len(result.Roster.Drivers),
	})
	for _, snapshot := range []store.Snapshot{
		{
			Source: "openf1", Resource: "sessions_sync", Endpoint: result.SessionsURL, FetchedAt: result.Roster.SyncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: 1, PayloadBytes: len(result.SessionsPayload), Summary: summary, Payload: result.SessionsPayload,
		},
		{
			Source: "openf1", Resource: "drivers_sync", Endpoint: result.DriversURL, FetchedAt: result.Roster.SyncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Roster.Drivers), PayloadBytes: len(result.DriversPayload), Summary: summary, Payload: result.DriversPayload,
		},
	} {
		if err := db.Save(ctx, snapshot); err != nil {
			return err
		}
	}
	fmt.Printf("SYNC drivers season=%d session=%d (%s) records=%d\n", year, result.Roster.Session.Key,
		result.Roster.Session.Location, len(result.Roster.Drivers))
	return nil
}

func syncClassifications(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := jolpica.New(upstream).Classifications(ctx, year)
	if err != nil {
		return err
	}
	if err := db.ReplaceClassifications(ctx, year, result.Classifications); err != nil {
		return err
	}
	qualifying, sprints := 0, 0
	for _, item := range result.Classifications {
		if item.Type == "qualifying" {
			qualifying++
		} else if item.Type == "sprint" {
			sprints++
		}
	}
	syncedAt := result.Classifications[0].SyncedAt
	for index, page := range result.Pages {
		if err := db.Save(ctx, store.Snapshot{
			Source: "jolpica", Resource: fmt.Sprintf("classifications_sync_page_%d", index+1),
			Endpoint: page.Endpoint, FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json",
			RecordCount: page.Records, PayloadBytes: len(page.Payload), Summary: mustJSON(map[string]any{
				"season": year,
				"page":   index + 1, "records": page.Records,
			}), Payload: page.Payload,
		}); err != nil {
			return err
		}
	}
	fmt.Printf("SYNC classifications season=%d qualifying=%d sprints=%d pages=%d\n", year, qualifying, sprints, len(result.Pages))
	return nil
}

func syncResults(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := jolpica.New(upstream).Results(ctx, year)
	if err != nil {
		return err
	}
	if err := db.ReplaceResults(ctx, year, result.Races); err != nil {
		return err
	}
	syncedAt := result.Races[0].SyncedAt
	total := 0
	for index, page := range result.Pages {
		total += page.Records
		snapshot := store.Snapshot{
			Source: "jolpica", Resource: fmt.Sprintf("results_sync_page_%d", index+1),
			Endpoint: page.Endpoint, FetchedAt: syncedAt, StatusCode: 200, ContentType: "application/json",
			RecordCount: page.Records, PayloadBytes: len(page.Payload),
			Summary: mustJSON(map[string]any{"season": year, "page": index + 1, "records": page.Records}), Payload: page.Payload,
		}
		if err := db.Save(ctx, snapshot); err != nil {
			return err
		}
	}
	fmt.Printf("SYNC results season=%d races=%d classifications=%d pages=%d\n", year, len(result.Races), total, len(result.Pages))
	return nil
}

func syncStandings(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := jolpica.New(upstream).Standings(ctx, year)
	if err != nil {
		return err
	}
	if err := db.ReplaceStandings(ctx, year, result.Drivers, result.Constructors); err != nil {
		return err
	}
	syncedAt := result.Drivers[0].SyncedAt
	for _, snapshot := range []store.Snapshot{
		{
			Source: "jolpica", Resource: "driver_standings_sync", Endpoint: result.DriversEndpoint, FetchedAt: syncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Drivers), PayloadBytes: len(result.DriversPayload),
			Summary: mustJSON(map[string]any{"season": year, "standings": len(result.Drivers)}), Payload: result.DriversPayload,
		},
		{
			Source: "jolpica", Resource: "constructor_standings_sync", Endpoint: result.ConstructorsEndpoint, FetchedAt: syncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Constructors), PayloadBytes: len(result.ConstructorsPayload),
			Summary: mustJSON(map[string]any{"season": year, "standings": len(result.Constructors)}), Payload: result.ConstructorsPayload,
		},
	} {
		if err := db.Save(ctx, snapshot); err != nil {
			return err
		}
	}
	fmt.Printf("SYNC standings season=%d drivers=%d constructors=%d\n", year, len(result.Drivers), len(result.Constructors))
	return nil
}

func mustJSON(value any) json.RawMessage {
	payload, _ := json.Marshal(value)
	return payload
}

func syncCalendar(ctx context.Context, db *store.Store, upstream *fetch.Client, year int) error {
	result, err := jolpica.New(upstream).Calendar(ctx, year)
	if err != nil {
		return err
	}
	if err := db.ReplaceCalendar(ctx, year, result.Events); err != nil {
		return err
	}
	syncedAt := result.Events[0].SyncedAt
	summary, _ := json.Marshal(map[string]any{"season": year, "events": len(result.Events)})
	if err := db.Save(ctx, store.Snapshot{
		Source: "jolpica", Resource: "calendar_sync", Endpoint: result.Endpoint, FetchedAt: syncedAt,
		StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Events),
		PayloadBytes: len(result.Payload), Summary: summary, Payload: result.Payload,
	}); err != nil {
		return err
	}
	fmt.Printf("SYNC calendar season=%d events=%d\n", year, len(result.Events))
	return nil
}
