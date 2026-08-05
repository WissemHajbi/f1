package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"oidysts/internal/domain"
)

type StintsResult struct {
	Stints   []domain.Stint
	Endpoint string
	Payload  []byte
}

func (c *Client) Stints(ctx context.Context, sessionKey int, driverNumber *int) (StintsResult, error) {
	if sessionKey <= 0 {
		return StintsResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": []string{strconv.Itoa(sessionKey)}}
	if driverNumber != nil {
		if *driverNumber <= 0 {
			return StintsResult{}, fmt.Errorf("driver number must be positive")
		}
		query.Set("driver_number", strconv.Itoa(*driverNumber))
	}
	endpoint := c.baseURL + "/stints?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return StintsResult{}, fmt.Errorf("fetch stints: %w", err)
	}
	var stints []domain.Stint
	if err := json.Unmarshal(response.Body, &stints); err != nil {
		return StintsResult{}, fmt.Errorf("decode stints: %w", err)
	}
	if len(stints) == 0 {
		return StintsResult{}, fmt.Errorf("stints returned no records")
	}
	for index := range stints {
		item := &stints[index]
		item.Compound = strings.ToUpper(strings.TrimSpace(item.Compound))
		if item.SessionKey != sessionKey || item.DriverNumber <= 0 || item.StintNumber <= 0 || item.LapStart <= 0 || item.Compound == "" {
			return StintsResult{}, fmt.Errorf("invalid stint at index %d", index)
		}
		if driverNumber != nil && item.DriverNumber != *driverNumber {
			return StintsResult{}, fmt.Errorf("unexpected driver at index %d", index)
		}
		if item.LapEnd != nil && *item.LapEnd < item.LapStart {
			return StintsResult{}, fmt.Errorf("invalid lap range at index %d", index)
		}
	}
	sort.Slice(stints, func(i, j int) bool {
		if stints[i].DriverNumber == stints[j].DriverNumber {
			return stints[i].StintNumber < stints[j].StintNumber
		}
		return stints[i].DriverNumber < stints[j].DriverNumber
	})
	return StintsResult{Stints: stints, Endpoint: endpoint, Payload: response.Body}, nil
}
