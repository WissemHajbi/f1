package domain

import "time"

type RaceSessionLink struct {
	Season     int       `json:"season"`
	Round      int       `json:"round"`
	SessionKey int       `json:"session_key"`
	Source     string    `json:"source"`
	LinkedAt   time.Time `json:"linked_at"`
}

type RaceHubDriver struct {
	SessionKey    int     `json:"session_key"`
	Number        int     `json:"driver_number"`
	BroadcastName string  `json:"broadcast_name"`
	FullName      string  `json:"full_name"`
	Acronym       string  `json:"name_acronym"`
	TeamName      string  `json:"team_name"`
	TeamColour    string  `json:"team_colour"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	CountryCode   *string `json:"country_code,omitempty"`
}

type RaceDriverStats struct {
	Driver          RaceHubDriver `json:"driver"`
	CompletedLaps   int           `json:"completed_laps"`
	BestLapNumber   *int          `json:"best_lap_number,omitempty"`
	BestLapDuration *float64      `json:"best_lap_duration,omitempty"`
	BestSector1     *float64      `json:"best_sector_1,omitempty"`
	BestSector2     *float64      `json:"best_sector_2,omitempty"`
	BestSector3     *float64      `json:"best_sector_3,omitempty"`
	TopSpeed        *int          `json:"top_speed,omitempty"`
	PitStops        int           `json:"pit_stops"`
	Stints          int           `json:"stints"`
	Overtakes       int           `json:"overtakes"`
	RadioMessages   int           `json:"radio_messages"`
}

type RaceLapStat struct {
	LapNumber       int      `json:"lap_number"`
	Duration        *float64 `json:"duration,omitempty"`
	Sector1Duration *float64 `json:"sector_1_duration,omitempty"`
	Sector2Duration *float64 `json:"sector_2_duration,omitempty"`
	Sector3Duration *float64 `json:"sector_3_duration,omitempty"`
	SpeedTrap       *int     `json:"speed_trap,omitempty"`
	IsPitOutLap     bool     `json:"is_pit_out_lap"`
}

type TrackPoint struct {
	X         int        `json:"x"`
	Y         int        `json:"y"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type RaceHub struct {
	Season               int               `json:"season"`
	Round                int               `json:"round"`
	SessionKey           int               `json:"session_key"`
	Drivers              []RaceDriverStats `json:"drivers"`
	SelectedDriverNumber int               `json:"selected_driver_number"`
	Laps                 []RaceLapStat     `json:"laps"`
	Track                []TrackPoint      `json:"track"`
	TrackSourceDriver    *int              `json:"track_source_driver,omitempty"`
	TrackSourceLap       *int              `json:"track_source_lap,omitempty"`
	TrackEstimatedWidthM *float64          `json:"track_estimated_width_m,omitempty"`
	TrackAttribution     string            `json:"track_attribution,omitempty"`
	TrackSourceURL       string            `json:"track_source_url,omitempty"`
	TrackAccuracy        string            `json:"track_accuracy,omitempty"`
	DriverTrace          []TrackPoint      `json:"driver_trace"`
	DriverTraceLap       *int              `json:"driver_trace_lap,omitempty"`
}
