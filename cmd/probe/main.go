package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"oidysts/internal/config"
	"oidysts/internal/fetch"
	"oidysts/internal/probe"
	"oidysts/internal/store"
)

func main() {
	var sources string
	var year int
	var staticPath string
	flag.StringVar(&sources, "sources", "all", "comma-separated: openf1,jolpica,rss,fia,static,all")
	flag.IntVar(&year, "year", 2025, "season used for calendar and standings checks")
	flag.StringVar(&staticPath, "static", "data/f1_teams_2026.json", "team metadata JSON path")
	flag.Parse()

	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	runner := probe.New(fetch.New(cfg.Timeout, cfg.UserAgent, cfg.MaxBody))
	results, probeErr := runner.Run(ctx, strings.Split(sources, ","), year, staticPath)
	failedToSave := false
	for _, result := range results {
		if err := db.Save(ctx, result); err != nil {
			logger.Error("save result", "source", result.Source, "resource", result.Resource, "error", err)
			failedToSave = true
			continue
		}
		fmt.Printf("PASS %-8s %-22s records=%-4d bytes=%d\n", result.Source, result.Resource, result.RecordCount, result.PayloadBytes)
	}
	if probeErr != nil {
		logger.Error("one or more probes failed", "error", probeErr)
	}
	if probeErr != nil || failedToSave {
		os.Exit(1)
	}
}
