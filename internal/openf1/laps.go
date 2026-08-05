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

type LapsResult struct {
	Laps     []domain.Lap
	Endpoint string
	Payload  []byte
}

type rawLap struct {
	SessionKey      int        `json:"session_key"`
	MeetingKey      int        `json:"meeting_key"`
	DriverNumber    int        `json:"driver_number"`
	LapNumber       int        `json:"lap_number"`
	DateStart       *time.Time `json:"date_start"`
	LapDuration     *float64   `json:"lap_duration"`
	Sector1Duration *float64   `json:"duration_sector_1"`
	Sector2Duration *float64   `json:"duration_sector_2"`
	Sector3Duration *float64   `json:"duration_sector_3"`
	I1Speed         *int       `json:"i1_speed"`
	I2Speed         *int       `json:"i2_speed"`
	SpeedTrap       *int       `json:"st_speed"`
	Sector1Segments []int      `json:"segments_sector_1"`
	Sector2Segments []int      `json:"segments_sector_2"`
	Sector3Segments []int      `json:"segments_sector_3"`
	IsPitOutLap     *bool      `json:"is_pit_out_lap"`
}

func (c *Client) Laps(ctx context.Context, sessionKey int, driverNumber *int) (LapsResult, error) {
	if sessionKey <= 0 {
		return LapsResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	if driverNumber != nil {
		if *driverNumber <= 0 {
			return LapsResult{}, fmt.Errorf("driver number must be positive")
		}
		query.Set("driver_number", strconv.Itoa(*driverNumber))
	}
	endpoint := c.baseURL + "/laps?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return LapsResult{}, fmt.Errorf("fetch laps: %w", err)
	}
	var raw []rawLap
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return LapsResult{}, fmt.Errorf("decode laps: %w", err)
	}
	if len(raw) == 0 {
		return LapsResult{}, fmt.Errorf("laps returned no records")
	}
	laps := make([]domain.Lap, 0, len(raw))
	for index, item := range raw {
		if item.SessionKey != sessionKey || item.DriverNumber <= 0 || item.LapNumber <= 0 {
			return LapsResult{}, fmt.Errorf("invalid lap at index %d", index)
		}
		if driverNumber != nil && item.DriverNumber != *driverNumber {
			return LapsResult{}, fmt.Errorf("unexpected driver at index %d", index)
		}
		pitOut := item.IsPitOutLap != nil && *item.IsPitOutLap
		laps = append(laps, domain.Lap{SessionKey: item.SessionKey, MeetingKey: item.MeetingKey,
			DriverNumber: item.DriverNumber, LapNumber: item.LapNumber, DateStart: item.DateStart,
			LapDuration: item.LapDuration, Sector1Duration: item.Sector1Duration,
			Sector2Duration: item.Sector2Duration, Sector3Duration: item.Sector3Duration,
			I1Speed: item.I1Speed, I2Speed: item.I2Speed, SpeedTrap: item.SpeedTrap,
			Sector1Segments: nonNilSegments(item.Sector1Segments), Sector2Segments: nonNilSegments(item.Sector2Segments),
			Sector3Segments: nonNilSegments(item.Sector3Segments), IsPitOutLap: pitOut})
	}
	sort.Slice(laps, func(i, j int) bool {
		if laps[i].DriverNumber == laps[j].DriverNumber {
			return laps[i].LapNumber < laps[j].LapNumber
		}
		return laps[i].DriverNumber < laps[j].DriverNumber
	})
	return LapsResult{Laps: laps, Endpoint: endpoint, Payload: response.Body}, nil
}

func nonNilSegments(value []int) []int {
	if value == nil {
		return []int{}
	}
	return value
}
