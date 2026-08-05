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
	"oidysts/internal/fetch"
)

const baseURL = "https://api.openf1.org/v1"

type Client struct {
	fetch   *fetch.Client
	now     func() time.Time
	baseURL string
}

type RosterResult struct {
	Roster          domain.DriverRoster
	SessionsURL     string
	SessionsPayload []byte
	DriversURL      string
	DriversPayload  []byte
}

func New(client *fetch.Client) *Client {
	return &Client{fetch: client, now: time.Now, baseURL: baseURL}
}

func (c *Client) LatestCompletedRaceRoster(ctx context.Context, year int) (RosterResult, error) {
	sessionsURL := c.baseURL + "/sessions?year=" + strconv.Itoa(year) + "&session_name=" + url.QueryEscape("Race")
	response, err := c.fetch.Get(ctx, sessionsURL, "application/json")
	if err != nil {
		return RosterResult{}, fmt.Errorf("fetch sessions: %w", err)
	}
	sessionsPayload := append([]byte(nil), response.Body...)
	var sessions []domain.Session
	if err := json.Unmarshal(response.Body, &sessions); err != nil {
		return RosterResult{}, fmt.Errorf("decode sessions: %w", err)
	}
	cutoff := c.now().UTC().Add(-30 * time.Minute)
	completed := sessions[:0]
	for _, session := range sessions {
		if session.Year == year && !session.DateEnd.IsZero() && session.DateEnd.Before(cutoff) {
			completed = append(completed, session)
		}
	}
	if len(completed) == 0 {
		return RosterResult{}, fmt.Errorf("no completed race found for %d", year)
	}
	sort.Slice(completed, func(i, j int) bool { return completed[i].DateEnd.After(completed[j].DateEnd) })
	session := completed[0]

	driversURL := c.baseURL + "/drivers?session_key=" + strconv.Itoa(session.Key)
	response, err = c.fetch.Get(ctx, driversURL, "application/json")
	if err != nil {
		return RosterResult{}, fmt.Errorf("fetch drivers: %w", err)
	}
	var drivers []domain.Driver
	if err := json.Unmarshal(response.Body, &drivers); err != nil {
		return RosterResult{}, fmt.Errorf("decode drivers: %w", err)
	}
	if len(drivers) == 0 {
		return RosterResult{}, fmt.Errorf("session %d returned no drivers", session.Key)
	}
	for i := range drivers {
		if drivers[i].SessionKey != session.Key || drivers[i].Number <= 0 {
			return RosterResult{}, fmt.Errorf("invalid driver row at index %d", i)
		}
	}
	sort.Slice(drivers, func(i, j int) bool { return drivers[i].Number < drivers[j].Number })
	return RosterResult{
		Roster:      domain.DriverRoster{Session: session, Drivers: drivers, SyncedAt: c.now().UTC()},
		SessionsURL: sessionsURL, SessionsPayload: sessionsPayload,
		DriversURL: driversURL, DriversPayload: response.Body,
	}, nil
}
