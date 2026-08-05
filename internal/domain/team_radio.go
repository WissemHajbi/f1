package domain

import "time"

type TeamRadio struct {
	ID              string    `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	SessionKey      int       `json:"session_key"`
	MeetingKey      int       `json:"meeting_key"`
	DriverNumber    int       `json:"driver_number"`
	AudioURL        string    `json:"audio_url"`
	RecordingSource string    `json:"-"`
	Audio           []byte    `json:"-"`
	ContentType     string    `json:"-"`
}

type TeamRadioQuery struct {
	SessionKey   int
	DriverNumber *int
	From         *time.Time
	To           *time.Time
	Limit        int
}
