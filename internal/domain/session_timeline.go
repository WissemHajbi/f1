package domain

import (
	"encoding/json"
	"time"
)

const (
	TimelineRaceControl = "race_control"
	TimelineOvertake    = "overtake"
	TimelinePitStop     = "pit_stop"
	TimelinePosition    = "position"
	TimelineStint       = "stint"
	TimelineTeamRadio   = "team_radio"
	TimelineWeather     = "weather"
)

var TimelineTypes = []string{
	TimelineRaceControl, TimelineOvertake, TimelinePitStop, TimelinePosition,
	TimelineStint, TimelineTeamRadio, TimelineWeather,
}

type SessionTimelineEvent struct {
	ID           string          `json:"id"`
	Timestamp    time.Time       `json:"timestamp"`
	SessionKey   int             `json:"session_key"`
	MeetingKey   int             `json:"meeting_key"`
	Type         string          `json:"type"`
	DriverNumber *int            `json:"driver_number,omitempty"`
	Title        string          `json:"title"`
	Data         json.RawMessage `json:"data"`
}

type SessionTimelineQuery struct {
	SessionKey   int
	DriverNumber *int
	Types        []string
	From         *time.Time
	To           *time.Time
	Limit        int
}
