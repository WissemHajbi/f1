package domain

import "time"

type SessionClassification struct {
	Season   int                    `json:"season"`
	Round    int                    `json:"round"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	StartAt  *time.Time             `json:"start_at"`
	Circuit  Circuit                `json:"circuit"`
	Results  []ClassificationResult `json:"results"`
	Source   string                 `json:"source"`
	SyncedAt time.Time              `json:"synced_at"`
}

type ClassificationResult struct {
	Position     int                 `json:"position"`
	DriverID     string              `json:"driver_id"`
	DriverNumber string              `json:"driver_number,omitempty"`
	DriverCode   string              `json:"driver_code,omitempty"`
	GivenName    string              `json:"given_name"`
	FamilyName   string              `json:"family_name"`
	Nationality  string              `json:"nationality"`
	Constructor  StandingConstructor `json:"constructor"`
	Q1           string              `json:"q1,omitempty"`
	Q2           string              `json:"q2,omitempty"`
	Q3           string              `json:"q3,omitempty"`
	Points       *float64            `json:"points,omitempty"`
	Grid         *int                `json:"grid,omitempty"`
	Laps         *int                `json:"laps,omitempty"`
	Status       string              `json:"status,omitempty"`
	Time         string              `json:"time,omitempty"`
	TimeMillis   *int64              `json:"time_millis,omitempty"`
	FastestLap   *FastestLap         `json:"fastest_lap,omitempty"`
}
