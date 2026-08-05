package domain

import "time"

type RaceControlEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	Category     string    `json:"category"`
	Message      string    `json:"message"`
	Flag         string    `json:"flag,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	DriverNumber *int      `json:"driver_number,omitempty"`
	LapNumber    *int      `json:"lap_number,omitempty"`
	Sector       *int      `json:"sector,omitempty"`
}

type RaceControlQuery struct {
	SessionKey   int
	Category     string
	DriverNumber *int
	From         *time.Time
	To           *time.Time
	Limit        int
}
