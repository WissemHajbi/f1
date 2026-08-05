package jolpica

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
	"oidysts/internal/fetch"
)

const baseURL = "https://api.jolpi.ca/ergast/f1"

type Client struct {
	fetch   *fetch.Client
	now     func() time.Time
	baseURL string
}

type CalendarResult struct {
	Events   []domain.Event
	Endpoint string
	Payload  []byte
}

type schedule struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

type race struct {
	Season   string `json:"season"`
	Round    string `json:"round"`
	URL      string `json:"url"`
	RaceName string `json:"raceName"`
	Circuit  struct {
		ID       string `json:"circuitId"`
		Name     string `json:"circuitName"`
		Location struct {
			Latitude  string `json:"lat"`
			Longitude string `json:"long"`
			Locality  string `json:"locality"`
			Country   string `json:"country"`
		} `json:"Location"`
	} `json:"Circuit"`
	Date             string   `json:"date"`
	Time             string   `json:"time"`
	FirstPractice    schedule `json:"FirstPractice"`
	SecondPractice   schedule `json:"SecondPractice"`
	ThirdPractice    schedule `json:"ThirdPractice"`
	Qualifying       schedule `json:"Qualifying"`
	Sprint           schedule `json:"Sprint"`
	SprintQualifying schedule `json:"SprintQualifying"`
	SprintShootout   schedule `json:"SprintShootout"`
}

type calendarResponse struct {
	MRData struct {
		Total     string `json:"total"`
		RaceTable struct {
			Races []race `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

func New(client *fetch.Client) *Client {
	return &Client{fetch: client, now: time.Now, baseURL: baseURL}
}

func (c *Client) Calendar(ctx context.Context, year int) (CalendarResult, error) {
	endpoint := c.baseURL + "/" + strconv.Itoa(year) + ".json?limit=100"
	response, err := c.fetch.Get(ctx, endpoint, "application/json")
	if err != nil {
		return CalendarResult{}, fmt.Errorf("fetch calendar: %w", err)
	}
	var payload calendarResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return CalendarResult{}, fmt.Errorf("decode calendar: %w", err)
	}
	if len(payload.MRData.RaceTable.Races) == 0 {
		return CalendarResult{}, fmt.Errorf("calendar returned no races for %d", year)
	}
	syncedAt := c.now().UTC()
	events := make([]domain.Event, 0, len(payload.MRData.RaceTable.Races))
	for index, item := range payload.MRData.RaceTable.Races {
		event, err := convertRace(item, syncedAt)
		if err != nil {
			return CalendarResult{}, fmt.Errorf("race %d: %w", index, err)
		}
		if event.Season != year {
			return CalendarResult{}, fmt.Errorf("unexpected season %d", event.Season)
		}
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Round < events[j].Round })
	return CalendarResult{Events: events, Endpoint: endpoint, Payload: response.Body}, nil
}

func convertRace(item race, syncedAt time.Time) (domain.Event, error) {
	season, err := strconv.Atoi(item.Season)
	if err != nil {
		return domain.Event{}, fmt.Errorf("invalid season %q", item.Season)
	}
	round, err := strconv.Atoi(item.Round)
	if err != nil {
		return domain.Event{}, fmt.Errorf("invalid round %q", item.Round)
	}
	latitude, err := strconv.ParseFloat(item.Circuit.Location.Latitude, 64)
	if err != nil {
		return domain.Event{}, fmt.Errorf("invalid latitude: %w", err)
	}
	longitude, err := strconv.ParseFloat(item.Circuit.Location.Longitude, 64)
	if err != nil {
		return domain.Event{}, fmt.Errorf("invalid longitude: %w", err)
	}
	raceAt, err := parseSchedule(schedule{Date: item.Date, Time: item.Time})
	if err != nil {
		return domain.Event{}, fmt.Errorf("race schedule: %w", err)
	}
	event := domain.Event{
		Season: season, Round: round, Name: item.RaceName, SourceURL: sanitizeURL(item.URL), RaceAt: raceAt, SyncedAt: syncedAt,
		Circuit: domain.Circuit{ID: item.Circuit.ID, Name: item.Circuit.Name, Locality: item.Circuit.Location.Locality,
			Country: item.Circuit.Location.Country, Latitude: latitude, Longitude: longitude},
	}
	for _, candidate := range []struct {
		name string
		data schedule
	}{
		{"practice_1", item.FirstPractice}, {"practice_2", item.SecondPractice}, {"practice_3", item.ThirdPractice},
		{"sprint_qualifying", item.SprintQualifying}, {"sprint_shootout", item.SprintShootout},
		{"sprint", item.Sprint}, {"qualifying", item.Qualifying}, {"race", schedule{Date: item.Date, Time: item.Time}},
	} {
		start, err := parseSchedule(candidate.data)
		if err != nil {
			return domain.Event{}, fmt.Errorf("%s schedule: %w", candidate.name, err)
		}
		if start != nil {
			event.Sessions = append(event.Sessions, domain.EventSession{Type: candidate.name, StartAt: start})
		}
	}
	sort.Slice(event.Sessions, func(i, j int) bool { return event.Sessions[i].StartAt.Before(*event.Sessions[j].StartAt) })
	return event, nil
}

func parseSchedule(value schedule) (*time.Time, error) {
	if value.Date == "" {
		return nil, nil
	}
	clock := strings.TrimSuffix(value.Time, "Z")
	if clock == "" {
		clock = "00:00:00"
	}
	parsed, err := time.Parse("2006-01-02T15:04:05", value.Date+"T"+clock)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func sanitizeURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return parsed.String()
}
