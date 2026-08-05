package domain

import "time"

type Lap struct {
	SessionKey      int        `json:"session_key"`
	MeetingKey      int        `json:"meeting_key"`
	DriverNumber    int        `json:"driver_number"`
	LapNumber       int        `json:"lap_number"`
	DateStart       *time.Time `json:"date_start,omitempty"`
	LapDuration     *float64   `json:"lap_duration,omitempty"`
	Sector1Duration *float64   `json:"sector_1_duration,omitempty"`
	Sector2Duration *float64   `json:"sector_2_duration,omitempty"`
	Sector3Duration *float64   `json:"sector_3_duration,omitempty"`
	I1Speed         *int       `json:"i1_speed,omitempty"`
	I2Speed         *int       `json:"i2_speed,omitempty"`
	SpeedTrap       *int       `json:"speed_trap,omitempty"`
	Sector1Segments []int      `json:"sector_1_segments"`
	Sector2Segments []int      `json:"sector_2_segments"`
	Sector3Segments []int      `json:"sector_3_segments"`
	IsPitOutLap     bool       `json:"is_pit_out_lap"`
}

type LapQuery struct {
	SessionKey   int
	DriverNumber *int
	LapNumber    *int
	Limit        int
}
