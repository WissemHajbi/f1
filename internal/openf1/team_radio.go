package openf1

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"oidysts/internal/domain"
)

type TeamRadioResult struct {
	Records  []domain.TeamRadio
	Endpoint string
	Payload  []byte
}

type rawTeamRadio struct {
	Date         time.Time `json:"date"`
	SessionKey   int       `json:"session_key"`
	MeetingKey   int       `json:"meeting_key"`
	DriverNumber int       `json:"driver_number"`
	RecordingURL string    `json:"recording_url"`
}

func (c *Client) TeamRadio(ctx context.Context, sessionKey int, driverNumber *int) (TeamRadioResult, error) {
	if sessionKey <= 0 {
		return TeamRadioResult{}, fmt.Errorf("session key must be positive")
	}
	query := url.Values{"session_key": {strconv.Itoa(sessionKey)}}
	if driverNumber != nil {
		if *driverNumber <= 0 {
			return TeamRadioResult{}, fmt.Errorf("driver number must be positive")
		}
		query.Set("driver_number", strconv.Itoa(*driverNumber))
	}
	endpoint := c.baseURL + "/team_radio?" + query.Encode()
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return TeamRadioResult{}, fmt.Errorf("fetch team radio: %w", err)
	}
	var raw []rawTeamRadio
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return TeamRadioResult{}, fmt.Errorf("decode team radio: %w", err)
	}
	if len(raw) == 0 {
		return TeamRadioResult{}, fmt.Errorf("team radio returned no records")
	}
	records := make([]domain.TeamRadio, 0, len(raw))
	for index, item := range raw {
		recordingURL, parseErr := url.Parse(item.RecordingURL)
		if item.SessionKey != sessionKey || item.Date.IsZero() || item.DriverNumber <= 0 || parseErr != nil ||
			recordingURL.Scheme != "https" || recordingURL.Host == "" {
			return TeamRadioResult{}, fmt.Errorf("invalid team radio record at index %d", index)
		}
		identity := strings.Join([]string{strconv.Itoa(item.SessionKey), strconv.Itoa(item.DriverNumber),
			item.Date.UTC().Format(time.RFC3339Nano), item.RecordingURL}, "\x1f")
		digest := sha256.Sum256([]byte(identity))
		records = append(records, domain.TeamRadio{ID: hex.EncodeToString(digest[:]), Timestamp: item.Date.UTC(),
			SessionKey: item.SessionKey, MeetingKey: item.MeetingKey, DriverNumber: item.DriverNumber,
			RecordingSource: item.RecordingURL})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Timestamp.Before(records[j].Timestamp) })
	return TeamRadioResult{Records: records, Endpoint: endpoint, Payload: response.Body}, nil
}
