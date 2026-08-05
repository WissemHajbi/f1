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

const MaxCarDataWindow = 15 * time.Minute

type CarDataResult struct {
	Samples  []domain.CarDataSample
	Endpoint string
	Payload  []byte
}

type rawCarData struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	Speed        int       `json:"speed"`
	RPM          int       `json:"rpm"`
	Gear         int       `json:"n_gear"`
	Throttle     int       `json:"throttle"`
	Brake        int       `json:"brake"`
	DRS          int       `json:"drs"`
}

func (c *Client) CarData(ctx context.Context, sessionKey, driverNumber int, from, to time.Time) (CarDataResult, error) {
	if sessionKey <= 0 || driverNumber <= 0 {
		return CarDataResult{}, fmt.Errorf("session and driver numbers must be positive")
	}
	from, to = from.UTC(), to.UTC()
	if !to.After(from) {
		return CarDataResult{}, fmt.Errorf("to must be after from")
	}
	if to.Sub(from) > MaxCarDataWindow {
		return CarDataResult{}, fmt.Errorf("time range exceeds %s", MaxCarDataWindow)
	}
	query := url.Values{}
	query.Set("session_key", strconv.Itoa(sessionKey))
	query.Set("driver_number", strconv.Itoa(driverNumber))
	query.Set("date>", from.Format(time.RFC3339Nano))
	query.Set("date<", to.Format(time.RFC3339Nano))
	endpoint := c.baseURL + "/car_data?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return CarDataResult{}, fmt.Errorf("fetch car data: %w", err)
	}
	var raw []rawCarData
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return CarDataResult{}, fmt.Errorf("decode car data: %w", err)
	}
	if len(raw) == 0 {
		return CarDataResult{}, fmt.Errorf("car data returned no samples")
	}
	samples := make([]domain.CarDataSample, 0, len(raw))
	for index, item := range raw {
		if item.SessionKey != sessionKey || item.DriverNumber != driverNumber || item.Date.IsZero() {
			return CarDataResult{}, fmt.Errorf("invalid car data sample at index %d", index)
		}
		samples = append(samples, domain.CarDataSample{Timestamp: item.Date.UTC(), SessionKey: item.SessionKey,
			MeetingKey: item.MeetingKey, DriverNumber: item.DriverNumber, Speed: item.Speed, RPM: item.RPM,
			Gear: item.Gear, Throttle: item.Throttle, Brake: item.Brake, DRS: item.DRS})
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].Timestamp.Before(samples[j].Timestamp) })
	return CarDataResult{Samples: samples, Endpoint: endpoint, Payload: response.Body}, nil
}
