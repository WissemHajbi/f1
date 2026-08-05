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
	"oidysts/internal/fetch"
	"oidysts/internal/openf1"
	"oidysts/internal/store"
)

func main() {
	var resource string
	var year int
	flag.StringVar(&resource, "resource", "drivers", "resource to sync (currently: drivers)")
	flag.IntVar(&year, "year", 2025, "season to sync")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if resource != "drivers" {
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
	client := openf1.New(fetch.New(cfg.Timeout, cfg.UserAgent, cfg.MaxBody))
	result, err := client.LatestCompletedRaceRoster(ctx, year)
	if err != nil {
		logger.Error("fetch driver roster", "error", err)
		os.Exit(1)
	}
	if err := db.SaveDriverRoster(ctx, result.Roster); err != nil {
		logger.Error("save driver roster", "error", err)
		os.Exit(1)
	}
	summary, _ := json.Marshal(map[string]any{
		"season": year, "session_key": result.Roster.Session.Key, "drivers": len(result.Roster.Drivers),
	})
	for _, snapshot := range []store.Snapshot{
		{Source: "openf1", Resource: "sessions_sync", Endpoint: result.SessionsURL, FetchedAt: result.Roster.SyncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: 1, PayloadBytes: len(result.SessionsPayload), Summary: summary, Payload: result.SessionsPayload},
		{Source: "openf1", Resource: "drivers_sync", Endpoint: result.DriversURL, FetchedAt: result.Roster.SyncedAt,
			StatusCode: 200, ContentType: "application/json", RecordCount: len(result.Roster.Drivers), PayloadBytes: len(result.DriversPayload), Summary: summary, Payload: result.DriversPayload},
	} {
		if err := db.Save(ctx, snapshot); err != nil {
			logger.Error("save source snapshot", "resource", snapshot.Resource, "error", err)
			os.Exit(1)
		}
	}
	fmt.Printf("SYNC drivers season=%d session=%d (%s) records=%d\n", year, result.Roster.Session.Key,
		result.Roster.Session.Location, len(result.Roster.Drivers))
}
