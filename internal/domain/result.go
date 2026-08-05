package domain

import "time"

type FastestLap struct {
	Rank         int     `json:"rank"`
	Lap          int     `json:"lap"`
	Time         string  `json:"time"`
	AverageSpeed float64 `json:"average_speed"`
	SpeedUnits   string  `json:"speed_units"`
}

type RaceResult struct {
	Position     int                 `json:"position"`
	PositionText string              `json:"position_text"`
	Points       float64             `json:"points"`
	Grid         int                 `json:"grid"`
	Laps         int                 `json:"laps"`
	Status       string              `json:"status"`
	Time         string              `json:"time,omitempty"`
	TimeMillis   *int64              `json:"time_millis,omitempty"`
	DriverID     string              `json:"driver_id"`
	DriverNumber string              `json:"driver_number,omitempty"`
	DriverCode   string              `json:"driver_code,omitempty"`
	GivenName    string              `json:"given_name"`
	FamilyName   string              `json:"family_name"`
	Nationality  string              `json:"nationality"`
	Constructor  StandingConstructor `json:"constructor"`
	FastestLap   *FastestLap         `json:"fastest_lap,omitempty"`
}

type RaceClassification struct {
	Season   int          `json:"season"`
	Round    int          `json:"round"`
	Name     string       `json:"name"`
	RaceAt   *time.Time   `json:"race_at"`
	Circuit  Circuit      `json:"circuit"`
	Results  []RaceResult `json:"results"`
	SyncedAt time.Time    `json:"synced_at"`
}
