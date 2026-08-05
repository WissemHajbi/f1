package domain

import "time"

type CarDataSample struct {
	Timestamp    time.Time `json:"timestamp"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	Speed        int       `json:"speed"`
	RPM          int       `json:"rpm"`
	Gear         int       `json:"gear"`
	Throttle     int       `json:"throttle"`
	Brake        int       `json:"brake"`
	DRS          int       `json:"drs"`
}

type LocationSample struct {
	Timestamp    time.Time `json:"timestamp"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	X            int       `json:"x"`
	Y            int       `json:"y"`
	Z            int       `json:"z"`
}

type LocationQuery struct {
	SessionKey   int
	DriverNumber int
	From         *time.Time
	To           *time.Time
	Limit        int
}

type CarDataQuery struct {
	SessionKey   int
	DriverNumber int
	From         *time.Time
	To           *time.Time
	Limit        int
}
