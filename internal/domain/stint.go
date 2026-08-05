package domain

type Stint struct {
	SessionKey     int    `json:"session_key"`
	MeetingKey     int    `json:"meeting_key"`
	DriverNumber   int    `json:"driver_number"`
	StintNumber    int    `json:"stint_number"`
	LapStart       int    `json:"lap_start"`
	LapEnd         *int   `json:"lap_end,omitempty"`
	Compound       string `json:"compound"`
	TyreAgeAtStart *int   `json:"tyre_age_at_start,omitempty"`
}

type StintQuery struct {
	SessionKey   int
	DriverNumber *int
	Limit        int
}
