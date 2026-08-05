package domain

import "time"

type Overtake struct {
	Timestamp             time.Time `json:"timestamp"`
	SessionKey            int       `json:"session_key"`
	MeetingKey            int       `json:"meeting_key"`
	DriverNumber          int       `json:"driver_number"`
	OvertakenDriverNumber int       `json:"overtaken_driver_number"`
	Position              int       `json:"position"`
}

type PositionSample struct {
	Timestamp    time.Time `json:"timestamp"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	Position     int       `json:"position"`
}

type IntervalSample struct {
	Timestamp          time.Time `json:"timestamp"`
	SessionKey         int       `json:"session_key"`
	MeetingKey         int       `json:"meeting_key"`
	DriverNumber       int       `json:"driver_number"`
	GapToLeader        string    `json:"gap_to_leader,omitempty"`
	GapToLeaderSeconds *float64  `json:"gap_to_leader_seconds,omitempty"`
	Interval           string    `json:"interval,omitempty"`
	IntervalSeconds    *float64  `json:"interval_seconds,omitempty"`
}

type TimelineQuery struct {
	SessionKey   int
	DriverNumber *int
	From         *time.Time
	To           *time.Time
	Limit        int
}

type OvertakeQuery struct {
	TimelineQuery
	OvertakenDriverNumber *int
}
