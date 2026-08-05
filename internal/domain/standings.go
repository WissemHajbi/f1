package domain

import "time"

type StandingConstructor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Nationality string `json:"nationality"`
}

type DriverStanding struct {
	Season          int                   `json:"season"`
	Round           int                   `json:"round"`
	Position        int                   `json:"position"`
	Points          float64               `json:"points"`
	Wins            int                   `json:"wins"`
	DriverID        string                `json:"driver_id"`
	PermanentNumber string                `json:"permanent_number,omitempty"`
	Code            string                `json:"code,omitempty"`
	GivenName       string                `json:"given_name"`
	FamilyName      string                `json:"family_name"`
	DateOfBirth     string                `json:"date_of_birth"`
	Nationality     string                `json:"nationality"`
	Constructors    []StandingConstructor `json:"constructors"`
	SyncedAt        time.Time             `json:"synced_at"`
}

type ConstructorStanding struct {
	Season      int                 `json:"season"`
	Round       int                 `json:"round"`
	Position    int                 `json:"position"`
	Points      float64             `json:"points"`
	Wins        int                 `json:"wins"`
	Constructor StandingConstructor `json:"constructor"`
	SyncedAt    time.Time           `json:"synced_at"`
}
