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

type CarDataQuery struct {
	SessionKey   int
	DriverNumber int
	From         *time.Time
	To           *time.Time
	Limit        int
}
