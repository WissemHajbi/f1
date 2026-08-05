package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"oidysts/internal/domain"
)

type WeatherResult struct {
	Samples  []domain.WeatherSample
	Endpoint string
	Payload  []byte
}

type rawWeather struct {
	Date             time.Time `json:"date"`
	SessionKey       int       `json:"session_key"`
	MeetingKey       int       `json:"meeting_key"`
	AirTemperature   *float64  `json:"air_temperature"`
	TrackTemperature *float64  `json:"track_temperature"`
	Humidity         *float64  `json:"humidity"`
	Pressure         *float64  `json:"pressure"`
	Rainfall         *float64  `json:"rainfall"`
	WindDirection    *float64  `json:"wind_direction"`
	WindSpeed        *float64  `json:"wind_speed"`
}

func (c *Client) Weather(ctx context.Context, sessionKey int) (WeatherResult, error) {
	if sessionKey <= 0 {
		return WeatherResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	endpoint := c.baseURL + "/weather?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return WeatherResult{}, fmt.Errorf("fetch weather: %w", err)
	}
	var raw []rawWeather
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return WeatherResult{}, fmt.Errorf("decode weather: %w", err)
	}
	if len(raw) == 0 {
		return WeatherResult{}, fmt.Errorf("weather returned no records")
	}
	samples := make([]domain.WeatherSample, 0, len(raw))
	for index, item := range raw {
		if item.SessionKey != sessionKey || item.Date.IsZero() {
			return WeatherResult{}, fmt.Errorf("invalid weather sample at index %d", index)
		}
		if item.Humidity != nil && (*item.Humidity < 0 || *item.Humidity > 100) {
			return WeatherResult{}, fmt.Errorf("invalid humidity at index %d", index)
		}
		samples = append(samples, domain.WeatherSample{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, AirTemperature: item.AirTemperature, TrackTemperature: item.TrackTemperature,
			Humidity: item.Humidity, Pressure: item.Pressure, Rainfall: item.Rainfall,
			WindDirection: item.WindDirection, WindSpeed: item.WindSpeed})
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].Timestamp.Before(samples[j].Timestamp) })
	return WeatherResult{Samples: samples, Endpoint: endpoint, Payload: response.Body}, nil
}
