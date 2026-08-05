package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"oidysts/internal/domain"
)

type ResultPage struct {
	Endpoint string
	Payload  []byte
	Records  int
}

type ResultsResult struct {
	Races []domain.RaceClassification
	Pages []ResultPage
}

type rawFastestLap struct {
	Rank string `json:"rank"`
	Lap  string `json:"lap"`
	Time struct {
		Value string `json:"time"`
	} `json:"Time"`
	AverageSpeed struct {
		Units string `json:"units"`
		Speed string `json:"speed"`
	} `json:"AverageSpeed"`
}

type rawRaceResult struct {
	Number       string `json:"number"`
	Position     string `json:"position"`
	PositionText string `json:"positionText"`
	Points       string `json:"points"`
	Driver       struct {
		ID          string `json:"driverId"`
		Code        string `json:"code"`
		GivenName   string `json:"givenName"`
		FamilyName  string `json:"familyName"`
		Nationality string `json:"nationality"`
	} `json:"Driver"`
	Constructor rawConstructor `json:"Constructor"`
	Grid        string         `json:"grid"`
	Laps        string         `json:"laps"`
	Status      string         `json:"status"`
	Time        struct {
		Millis string `json:"millis"`
		Value  string `json:"time"`
	} `json:"Time"`
	FastestLap *rawFastestLap `json:"FastestLap"`
}

type resultRace struct {
	Season   string `json:"season"`
	Round    string `json:"round"`
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
	Date          string          `json:"date"`
	Time          string          `json:"time"`
	Results       []rawRaceResult `json:"Results"`
	SprintResults []rawRaceResult `json:"SprintResults"`
}

type resultsResponse struct {
	MRData struct {
		Total     string `json:"total"`
		RaceTable struct {
			Races []resultRace `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

func (c *Client) Results(ctx context.Context, year int) (ResultsResult, error) {
	const pageSize = 100
	byRound := map[int]*domain.RaceClassification{}
	var pages []ResultPage
	for offset := 0; ; offset += pageSize {
		if offset > 0 {
			if err := wait(ctx, 260*time.Millisecond); err != nil {
				return ResultsResult{}, err
			}
		}
		endpoint := c.baseURL + "/" + strconv.Itoa(year) + "/results.json?limit=" + strconv.Itoa(pageSize) + "&offset=" + strconv.Itoa(offset)
		response, err := c.fetch.Get(ctx, endpoint, "application/json")
		if err != nil {
			return ResultsResult{}, fmt.Errorf("fetch results offset %d: %w", offset, err)
		}
		var payload resultsResponse
		if err := json.Unmarshal(response.Body, &payload); err != nil {
			return ResultsResult{}, fmt.Errorf("decode results offset %d: %w", offset, err)
		}
		total, err := strconv.Atoi(payload.MRData.Total)
		if err != nil {
			return ResultsResult{}, fmt.Errorf("invalid results total %q", payload.MRData.Total)
		}
		records := 0
		for _, rawRace := range payload.MRData.RaceTable.Races {
			races, err := convertResultRace(rawRace, year)
			if err != nil {
				return ResultsResult{}, err
			}
			records += len(races.Results)
			existing := byRound[races.Round]
			if existing == nil {
				byRound[races.Round] = &races
			} else {
				existing.Results = append(existing.Results, races.Results...)
			}
		}
		pages = append(pages, ResultPage{Endpoint: endpoint, Payload: response.Body, Records: records})
		if offset+records >= total {
			break
		}
		if records == 0 {
			return ResultsResult{}, fmt.Errorf("empty results page before total at offset %d", offset)
		}
	}
	if len(byRound) == 0 {
		return ResultsResult{}, fmt.Errorf("results returned no races for %d", year)
	}
	syncedAt := c.now().UTC()
	races := make([]domain.RaceClassification, 0, len(byRound))
	for _, race := range byRound {
		race.SyncedAt = syncedAt
		sort.Slice(race.Results, func(i, j int) bool { return race.Results[i].Position < race.Results[j].Position })
		races = append(races, *race)
	}
	sort.Slice(races, func(i, j int) bool { return races[i].Round < races[j].Round })
	return ResultsResult{Races: races, Pages: pages}, nil
}

func convertResultRace(item resultRace, expectedYear int) (domain.RaceClassification, error) {
	season, round, err := parseSeasonRound(item.Season, item.Round, expectedYear)
	if err != nil {
		return domain.RaceClassification{}, err
	}
	latitude, err := strconv.ParseFloat(item.Circuit.Location.Latitude, 64)
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("round %d latitude: %w", round, err)
	}
	longitude, err := strconv.ParseFloat(item.Circuit.Location.Longitude, 64)
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("round %d longitude: %w", round, err)
	}
	raceAt, err := parseSchedule(schedule{Date: item.Date, Time: item.Time})
	if err != nil {
		return domain.RaceClassification{}, fmt.Errorf("round %d race time: %w", round, err)
	}
	race := domain.RaceClassification{Season: season, Round: round, Name: item.RaceName, RaceAt: raceAt,
		Circuit: domain.Circuit{ID: item.Circuit.ID, Name: item.Circuit.Name, Locality: item.Circuit.Location.Locality,
			Country: item.Circuit.Location.Country, Latitude: latitude, Longitude: longitude}}
	for _, raw := range item.Results {
		result, err := convertResult(raw)
		if err != nil {
			return domain.RaceClassification{}, fmt.Errorf("round %d driver %s: %w", round, raw.Driver.ID, err)
		}
		race.Results = append(race.Results, result)
	}
	return race, nil
}

func convertResult(raw rawRaceResult) (domain.RaceResult, error) {
	position, err := strconv.Atoi(raw.Position)
	if err != nil {
		return domain.RaceResult{}, fmt.Errorf("invalid position %q", raw.Position)
	}
	points, err := strconv.ParseFloat(raw.Points, 64)
	if err != nil {
		return domain.RaceResult{}, fmt.Errorf("invalid points %q", raw.Points)
	}
	grid, err := strconv.Atoi(raw.Grid)
	if err != nil {
		return domain.RaceResult{}, fmt.Errorf("invalid grid %q", raw.Grid)
	}
	laps, err := strconv.Atoi(raw.Laps)
	if err != nil {
		return domain.RaceResult{}, fmt.Errorf("invalid laps %q", raw.Laps)
	}
	result := domain.RaceResult{Position: position, PositionText: raw.PositionText, Points: points, Grid: grid,
		Laps: laps, Status: raw.Status, Time: raw.Time.Value, DriverID: raw.Driver.ID, DriverNumber: raw.Number,
		DriverCode: raw.Driver.Code, GivenName: raw.Driver.GivenName, FamilyName: raw.Driver.FamilyName,
		Nationality: raw.Driver.Nationality, Constructor: domain.StandingConstructor{ID: raw.Constructor.ID,
			Name: raw.Constructor.Name, Nationality: raw.Constructor.Nationality}}
	if raw.Time.Millis != "" {
		value, err := strconv.ParseInt(raw.Time.Millis, 10, 64)
		if err != nil {
			return domain.RaceResult{}, fmt.Errorf("invalid milliseconds %q", raw.Time.Millis)
		}
		result.TimeMillis = &value
	}
	if raw.FastestLap != nil {
		rank, err := strconv.Atoi(raw.FastestLap.Rank)
		if err != nil {
			return domain.RaceResult{}, fmt.Errorf("invalid fastest-lap rank %q", raw.FastestLap.Rank)
		}
		lap, err := strconv.Atoi(raw.FastestLap.Lap)
		if err != nil {
			return domain.RaceResult{}, fmt.Errorf("invalid fastest-lap lap %q", raw.FastestLap.Lap)
		}
		speed, err := strconv.ParseFloat(raw.FastestLap.AverageSpeed.Speed, 64)
		if err != nil && raw.FastestLap.AverageSpeed.Speed != "" {
			return domain.RaceResult{}, fmt.Errorf("invalid fastest-lap speed")
		}
		result.FastestLap = &domain.FastestLap{Rank: rank, Lap: lap, Time: raw.FastestLap.Time.Value,
			AverageSpeed: speed, SpeedUnits: raw.FastestLap.AverageSpeed.Units}
	}
	return result, nil
}
