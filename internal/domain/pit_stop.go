package domain

import "time"

type PitStop struct {
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	LapNumber    int       `json:"lap_number"`
	Timestamp    time.Time `json:"timestamp"`
	Duration     *float64  `json:"duration,omitempty"`
}

type PitStopQuery struct {
	SessionKey   int
	DriverNumber *int
	Limit        int
}
