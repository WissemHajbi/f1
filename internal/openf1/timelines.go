package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"oidysts/internal/domain"
)

type OvertakesResult struct {
	Events   []domain.Overtake
	Endpoint string
	Payload  []byte
}
type PositionsResult struct {
	Samples  []domain.PositionSample
	Endpoint string
	Payload  []byte
}
type IntervalsResult struct {
	Samples  []domain.IntervalSample
	Endpoint string
	Payload  []byte
}

type rawOvertake struct {
	Date                  time.Time `json:"date"`
	SessionKey            int       `json:"session_key"`
	MeetingKey            int       `json:"meeting_key"`
	DriverNumber          int       `json:"overtaking_driver_number"`
	OvertakenDriverNumber int       `json:"overtaken_driver_number"`
	Position              int       `json:"position"`
}
type rawPosition struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	Position     int       `json:"position"`
}
type rawInterval struct {
	Date         time.Time       `json:"date"`
	SessionKey   int             `json:"session_key"`
	MeetingKey   int             `json:"meeting_key"`
	DriverNumber int             `json:"driver_number"`
	GapToLeader  json.RawMessage `json:"gap_to_leader"`
	Interval     json.RawMessage `json:"interval"`
}

func (c *Client) Overtakes(ctx context.Context, sessionKey int, driverNumber *int) (OvertakesResult, error) {
	endpoint, response, err := c.timelineResponse(ctx, "overtakes", "overtaking_driver_number", sessionKey, driverNumber)
	if err != nil {
		return OvertakesResult{}, fmt.Errorf("fetch overtakes: %w", err)
	}
	var raw []rawOvertake
	if err := json.Unmarshal(response, &raw); err != nil {
		return OvertakesResult{}, fmt.Errorf("decode overtakes: %w", err)
	}
	if len(raw) == 0 {
		return OvertakesResult{}, fmt.Errorf("overtakes returned no records")
	}
	items := make([]domain.Overtake, 0, len(raw))
	for i, item := range raw {
		if item.SessionKey != sessionKey || item.Date.IsZero() || item.DriverNumber <= 0 || item.OvertakenDriverNumber <= 0 || item.Position <= 0 {
			return OvertakesResult{}, fmt.Errorf("invalid overtake at index %d", i)
		}
		items = append(items, domain.Overtake{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey, MeetingKey: item.MeetingKey,
			DriverNumber: item.DriverNumber, OvertakenDriverNumber: item.OvertakenDriverNumber, Position: item.Position})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.Before(items[j].Timestamp) })
	return OvertakesResult{Events: items, Endpoint: endpoint, Payload: response}, nil
}

func (c *Client) Positions(ctx context.Context, sessionKey int, driverNumber *int) (PositionsResult, error) {
	endpoint, response, err := c.timelineResponse(ctx, "position", "driver_number", sessionKey, driverNumber)
	if err != nil {
		return PositionsResult{}, fmt.Errorf("fetch positions: %w", err)
	}
	var raw []rawPosition
	if err := json.Unmarshal(response, &raw); err != nil {
		return PositionsResult{}, fmt.Errorf("decode positions: %w", err)
	}
	if len(raw) == 0 {
		return PositionsResult{}, fmt.Errorf("positions returned no records")
	}
	items := make([]domain.PositionSample, 0, len(raw))
	for i, item := range raw {
		if item.SessionKey != sessionKey || item.Date.IsZero() || item.DriverNumber <= 0 || item.Position <= 0 {
			return PositionsResult{}, fmt.Errorf("invalid position at index %d", i)
		}
		items = append(items, domain.PositionSample{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, DriverNumber: item.DriverNumber, Position: item.Position})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.Before(items[j].Timestamp) })
	return PositionsResult{Samples: items, Endpoint: endpoint, Payload: response}, nil
}

func (c *Client) Intervals(ctx context.Context, sessionKey int, driverNumber *int) (IntervalsResult, error) {
	endpoint, response, err := c.timelineResponse(ctx, "intervals", "driver_number", sessionKey, driverNumber)
	if err != nil {
		return IntervalsResult{}, fmt.Errorf("fetch intervals: %w", err)
	}
	var raw []rawInterval
	if err := json.Unmarshal(response, &raw); err != nil {
		return IntervalsResult{}, fmt.Errorf("decode intervals: %w", err)
	}
	if len(raw) == 0 {
		return IntervalsResult{}, fmt.Errorf("intervals returned no records")
	}
	items := make([]domain.IntervalSample, 0, len(raw))
	for i, item := range raw {
		if item.SessionKey != sessionKey || item.Date.IsZero() || item.DriverNumber <= 0 {
			return IntervalsResult{}, fmt.Errorf("invalid interval at index %d", i)
		}
		gap, gapSeconds, err := flexibleNumber(item.GapToLeader)
		if err != nil {
			return IntervalsResult{}, fmt.Errorf("gap at index %d: %w", i, err)
		}
		interval, intervalSeconds, err := flexibleNumber(item.Interval)
		if err != nil {
			return IntervalsResult{}, fmt.Errorf("interval at index %d: %w", i, err)
		}
		items = append(items, domain.IntervalSample{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, DriverNumber: item.DriverNumber, GapToLeader: gap,
			GapToLeaderSeconds: gapSeconds, Interval: interval, IntervalSeconds: intervalSeconds})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.Before(items[j].Timestamp) })
	return IntervalsResult{Samples: items, Endpoint: endpoint, Payload: response}, nil
}

func (c *Client) timelineResponse(ctx context.Context, resource, driverField string, sessionKey int, driverNumber *int) (string, []byte, error) {
	if sessionKey <= 0 {
		return "", nil, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	if driverNumber != nil {
		if *driverNumber <= 0 {
			return "", nil, fmt.Errorf("driver number must be positive")
		}
		query.Set(driverField, strconv.Itoa(*driverNumber))
	}
	endpoint := c.baseURL + "/" + resource + "?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return "", nil, err
	}
	return endpoint, response.Body, nil
}

func flexibleNumber(raw json.RawMessage) (string, *float64, error) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return "", nil, nil
	}
	if strings.HasPrefix(value, `"`) {
		if err := json.Unmarshal(raw, &value); err != nil {
			return "", nil, err
		}
	}
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return value, &number, nil
	}
	return value, nil, nil
}
