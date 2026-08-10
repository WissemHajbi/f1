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

type PitStopsResult struct {
	PitStops []domain.PitStop
	Endpoint string
	Payload  []byte
}

type rawPitStop struct {
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	LapNumber    int       `json:"lap_number"`
	Date         time.Time `json:"date"`
	Duration     *float64  `json:"pit_duration"`
}

func (c *Client) PitStops(ctx context.Context, sessionKey int, driverNumber *int) (PitStopsResult, error) {
	if sessionKey <= 0 {
		return PitStopsResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	if driverNumber != nil {
		if *driverNumber <= 0 {
			return PitStopsResult{}, fmt.Errorf("driver number must be positive")
		}
		query.Set("driver_number", strconv.Itoa(*driverNumber))
	}
	endpoint := c.baseURL + "/pit?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return PitStopsResult{}, fmt.Errorf("fetch pit stops: %w", err)
	}
	var raw []rawPitStop
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return PitStopsResult{}, fmt.Errorf("decode pit stops: %w", err)
	}
	if len(raw) == 0 {
		return PitStopsResult{}, fmt.Errorf("pit stops returned no records")
	}
	items := make([]domain.PitStop, 0, len(raw))
	for index, item := range raw {
		if item.SessionKey != sessionKey || item.DriverNumber <= 0 || item.LapNumber <= 0 || item.Date.IsZero() {
			return PitStopsResult{}, fmt.Errorf("invalid pit stop at index %d", index)
		}
		if driverNumber != nil && item.DriverNumber != *driverNumber {
			return PitStopsResult{}, fmt.Errorf("unexpected driver at index %d", index)
		}
		if item.Duration != nil && *item.Duration < 0 {
			return PitStopsResult{}, fmt.Errorf("negative duration at index %d", index)
		}
		items = append(items, domain.PitStop{
			SessionKey: item.SessionKey, MeetingKey: item.MeetingKey,
			DriverNumber: item.DriverNumber, LapNumber: item.LapNumber, Timestamp: item.Date.UTC(), Duration: item.Duration,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.Before(items[j].Timestamp) })
	return PitStopsResult{PitStops: items, Endpoint: endpoint, Payload: response.Body}, nil
}
