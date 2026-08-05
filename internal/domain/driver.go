package domain

import "time"

type Session struct {
	Key       int       `json:"session_key"`
	Name      string    `json:"session_name"`
	Type      string    `json:"session_type"`
	Year      int       `json:"year"`
	Location  string    `json:"location"`
	DateStart time.Time `json:"date_start"`
	DateEnd   time.Time `json:"date_end"`
}

type Driver struct {
	SessionKey    int     `json:"session_key"`
	Number        int     `json:"driver_number"`
	BroadcastName string  `json:"broadcast_name"`
	FullName      string  `json:"full_name"`
	Acronym       string  `json:"name_acronym"`
	TeamName      string  `json:"team_name"`
	TeamColour    string  `json:"team_colour"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	HeadshotURL   string  `json:"headshot_url"`
	CountryCode   *string `json:"country_code"`
}

type DriverRoster struct {
	Session  Session   `json:"session"`
	Drivers  []Driver  `json:"drivers"`
	SyncedAt time.Time `json:"synced_at"`
}
