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

const MaxLocationWindow = 15 * time.Minute

type LocationResult struct {
	Samples  []domain.LocationSample
	Endpoint string
	Payload  []byte
}

type rawLocation struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	X            int       `json:"x"`
	Y            int       `json:"y"`
	Z            int       `json:"z"`
}

func (c *Client) Location(ctx context.Context, sessionKey, driverNumber int, from, to time.Time) (LocationResult, error) {
	if sessionKey <= 0 || driverNumber <= 0 {
		return LocationResult{}, fmt.Errorf("session and driver numbers must be positive")
	}
	from, to = from.UTC(), to.UTC()
	if !to.After(from) {
		return LocationResult{}, fmt.Errorf("to must be after from")
	}
	if to.Sub(from) > MaxLocationWindow {
		return LocationResult{}, fmt.Errorf("time range exceeds %s", MaxLocationWindow)
	}
	query := url.Values{"session_key": {strconv.Itoa(sessionKey)}, "driver_number": {strconv.Itoa(driverNumber)},
		"date>": {from.Format(time.RFC3339Nano)}, "date<": {to.Format(time.RFC3339Nano)}}
	endpoint := c.baseURL + "/location?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return LocationResult{}, fmt.Errorf("fetch location: %w", err)
	}
	var raw []rawLocation
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return LocationResult{}, fmt.Errorf("decode location: %w", err)
	}
	if len(raw) == 0 {
		return LocationResult{}, fmt.Errorf("location returned no samples")
	}
	samples := make([]domain.LocationSample, 0, len(raw))
	for index, item := range raw {
		if item.SessionKey != sessionKey || item.DriverNumber != driverNumber || item.Date.IsZero() {
			return LocationResult{}, fmt.Errorf("invalid location sample at index %d", index)
		}
		samples = append(samples, domain.LocationSample{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, DriverNumber: item.DriverNumber, X: item.X, Y: item.Y, Z: item.Z})
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].Timestamp.Before(samples[j].Timestamp) })
	return LocationResult{Samples: samples, Endpoint: endpoint, Payload: response.Body}, nil
}
