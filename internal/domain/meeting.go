package domain

import "time"

type Meeting struct {
	Key              int              `json:"meeting_key"`
	Name             string           `json:"meeting_name"`
	OfficialName     string           `json:"meeting_official_name"`
	Year             int              `json:"year"`
	Location         string           `json:"location"`
	CountryKey       int              `json:"country_key"`
	CountryCode      string           `json:"country_code"`
	CountryName      string           `json:"country_name"`
	CircuitKey       int              `json:"circuit_key"`
	CircuitShortName string           `json:"circuit_short_name"`
	DateStart        time.Time        `json:"date_start"`
	GMTOffset        string           `json:"gmt_offset"`
	Sessions         []MeetingSession `json:"sessions"`
	SyncedAt         time.Time        `json:"synced_at"`
}

type MeetingSession struct {
	Key              int       `json:"session_key"`
	MeetingKey       int       `json:"meeting_key"`
	Name             string    `json:"session_name"`
	Type             string    `json:"session_type"`
	Year             int       `json:"year"`
	Location         string    `json:"location"`
	CountryCode      string    `json:"country_code"`
	CountryName      string    `json:"country_name"`
	CircuitKey       int       `json:"circuit_key"`
	CircuitShortName string    `json:"circuit_short_name"`
	DateStart        time.Time `json:"date_start"`
	DateEnd          time.Time `json:"date_end"`
	GMTOffset        string    `json:"gmt_offset"`
	IsCancelled      bool      `json:"is_cancelled"`
}
