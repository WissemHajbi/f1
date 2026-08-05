package domain

import "time"

type WeatherSample struct {
	Timestamp        time.Time `json:"timestamp"`
	SessionKey       int       `json:"session_key"`
	MeetingKey       int       `json:"meeting_key"`
	AirTemperature   *float64  `json:"air_temperature,omitempty"`
	TrackTemperature *float64  `json:"track_temperature,omitempty"`
	Humidity         *float64  `json:"humidity,omitempty"`
	Pressure         *float64  `json:"pressure,omitempty"`
	Rainfall         *float64  `json:"rainfall,omitempty"`
	WindDirection    *float64  `json:"wind_direction,omitempty"`
	WindSpeed        *float64  `json:"wind_speed,omitempty"`
}

type WeatherQuery struct {
	SessionKey int
	From       *time.Time
	To         *time.Time
	Limit      int
}
