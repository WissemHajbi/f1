package domain

import "time"

type Circuit struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Locality  string  `json:"locality"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type EventSession struct {
	Type    string     `json:"type"`
	StartAt *time.Time `json:"start_at"`
}

type Event struct {
	Season    int            `json:"season"`
	Round     int            `json:"round"`
	Name      string         `json:"name"`
	SourceURL string         `json:"source_url"`
	RaceAt    *time.Time     `json:"race_at"`
	Circuit   Circuit        `json:"circuit"`
	Sessions  []EventSession `json:"sessions"`
	SyncedAt  time.Time      `json:"synced_at"`
}
