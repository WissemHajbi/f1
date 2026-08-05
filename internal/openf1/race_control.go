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

type RaceControlResult struct {
	Events   []domain.RaceControlEvent
	Endpoint string
	Payload  []byte
}

type rawRaceControl struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	Category     string    `json:"category"`
	Message      string    `json:"message"`
	Flag         string    `json:"flag"`
	Scope        string    `json:"scope"`
	DriverNumber *int      `json:"driver_number"`
	LapNumber    *int      `json:"lap_number"`
	Sector       *int      `json:"sector"`
}

func (c *Client) RaceControl(ctx context.Context, sessionKey int) (RaceControlResult, error) {
	if sessionKey <= 0 {
		return RaceControlResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	endpoint := c.baseURL + "/race_control?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return RaceControlResult{}, fmt.Errorf("fetch race control: %w", err)
	}
	var raw []rawRaceControl
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return RaceControlResult{}, fmt.Errorf("decode race control: %w", err)
	}
	if len(raw) == 0 {
		return RaceControlResult{}, fmt.Errorf("race control returned no records")
	}
	events := make([]domain.RaceControlEvent, 0, len(raw))
	for index, item := range raw {
		item.Category, item.Message = strings.TrimSpace(item.Category), strings.TrimSpace(item.Message)
		if item.SessionKey != sessionKey || item.Date.IsZero() || item.Category == "" || item.Message == "" {
			return RaceControlResult{}, fmt.Errorf("invalid race-control event at index %d", index)
		}
		events = append(events, domain.RaceControlEvent{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, Category: item.Category, Message: item.Message, Flag: item.Flag,
			Scope: item.Scope, DriverNumber: item.DriverNumber, LapNumber: item.LapNumber, Sector: item.Sector})
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })
	return RaceControlResult{Events: events, Endpoint: endpoint, Payload: response.Body}, nil
}
