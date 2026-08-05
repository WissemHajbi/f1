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

type ClassificationsResult struct {
	Classifications []domain.SessionClassification
	Pages           []ResultPage
}

type rawQualifyingResult struct {
	Number   string `json:"number"`
	Position string `json:"position"`
	Driver   struct {
		ID, Code, GivenName, FamilyName, Nationality string
	} `json:"Driver"`
	Constructor rawConstructor `json:"Constructor"`
	Q1          string         `json:"Q1"`
	Q2          string         `json:"Q2"`
	Q3          string         `json:"Q3"`
}

func (r *rawQualifyingResult) UnmarshalJSON(data []byte) error {
	type alias rawQualifyingResult
	var value struct {
		alias
		Driver struct {
			ID          string `json:"driverId"`
			Code        string `json:"code"`
			GivenName   string `json:"givenName"`
			FamilyName  string `json:"familyName"`
			Nationality string `json:"nationality"`
		} `json:"Driver"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = rawQualifyingResult(value.alias)
	r.Driver.ID, r.Driver.Code, r.Driver.GivenName = value.Driver.ID, value.Driver.Code, value.Driver.GivenName
	r.Driver.FamilyName, r.Driver.Nationality = value.Driver.FamilyName, value.Driver.Nationality
	return nil
}

type qualifyingRace struct {
	Season, Round, RaceName, Date, Time string
	Circuit                             struct {
		ID       string `json:"circuitId"`
		Name     string `json:"circuitName"`
		Location struct {
			Latitude  string `json:"lat"`
			Longitude string `json:"long"`
			Locality  string `json:"locality"`
			Country   string `json:"country"`
		} `json:"Location"`
	} `json:"Circuit"`
	Results []rawQualifyingResult `json:"QualifyingResults"`
}

func (r *qualifyingRace) UnmarshalJSON(data []byte) error {
	type alias qualifyingRace
	var value struct {
		alias
		Season   string `json:"season"`
		Round    string `json:"round"`
		RaceName string `json:"raceName"`
		Date     string `json:"date"`
		Time     string `json:"time"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*r = qualifyingRace(value.alias)
	r.Season, r.Round, r.RaceName, r.Date, r.Time = value.Season, value.Round, value.RaceName, value.Date, value.Time
	return nil
}

type qualifyingResponse struct {
	MRData struct {
		Total     string `json:"total"`
		RaceTable struct {
			Races []qualifyingRace `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

func (c *Client) Classifications(ctx context.Context, year int) (ClassificationsResult, error) {
	qualifying, qualifyingPages, err := c.qualifyingClassifications(ctx, year)
	if err != nil {
		return ClassificationsResult{}, err
	}
	if err := wait(ctx, 260*time.Millisecond); err != nil {
		return ClassificationsResult{}, err
	}
	sprints, sprintPages, err := c.sprintClassifications(ctx, year)
	if err != nil {
		return ClassificationsResult{}, err
	}
	items := append(qualifying, sprints...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Round == items[j].Round {
			return items[i].Type < items[j].Type
		}
		return items[i].Round < items[j].Round
	})
	return ClassificationsResult{Classifications: items, Pages: append(qualifyingPages, sprintPages...)}, nil
}

func (c *Client) qualifyingClassifications(ctx context.Context, year int) ([]domain.SessionClassification, []ResultPage, error) {
	const limit = 100
	byRound := map[int]*domain.SessionClassification{}
	var pages []ResultPage
	for offset := 0; ; offset += limit {
		if offset > 0 {
			if err := wait(ctx, 260*time.Millisecond); err != nil {
				return nil, nil, err
			}
		}
		endpoint := fmt.Sprintf("%s/%d/qualifying.json?limit=%d&offset=%d", c.baseURL, year, limit, offset)
		response, err := c.fetch.Get(ctx, endpoint, "application/json")
		if err != nil {
			return nil, nil, fmt.Errorf("fetch qualifying offset %d: %w", offset, err)
		}
		var payload qualifyingResponse
		if err := json.Unmarshal(response.Body, &payload); err != nil {
			return nil, nil, fmt.Errorf("decode qualifying: %w", err)
		}
		total, err := strconv.Atoi(payload.MRData.Total)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid qualifying total")
		}
		records := 0
		for _, race := range payload.MRData.RaceTable.Races {
			item, err := convertQualifying(race, year)
			if err != nil {
				return nil, nil, err
			}
			records += len(item.Results)
			if existing := byRound[item.Round]; existing != nil {
				existing.Results = append(existing.Results, item.Results...)
			} else {
				byRound[item.Round] = &item
			}
		}
		pages = append(pages, ResultPage{Endpoint: endpoint, Payload: response.Body, Records: records})
		if offset+records >= total {
			break
		}
		if records == 0 {
			return nil, nil, fmt.Errorf("empty qualifying page at offset %d", offset)
		}
	}
	if len(byRound) == 0 {
		return nil, nil, fmt.Errorf("qualifying returned no records for %d", year)
	}
	return finalizeClassifications(byRound, c.now().UTC()), pages, nil
}

func (c *Client) sprintClassifications(ctx context.Context, year int) ([]domain.SessionClassification, []ResultPage, error) {
	const limit = 100
	byRound := map[int]*domain.SessionClassification{}
	var pages []ResultPage
	for offset := 0; ; offset += limit {
		if offset > 0 {
			if err := wait(ctx, 260*time.Millisecond); err != nil {
				return nil, nil, err
			}
		}
		endpoint := fmt.Sprintf("%s/%d/sprint.json?limit=%d&offset=%d", c.baseURL, year, limit, offset)
		response, err := c.fetch.Get(ctx, endpoint, "application/json")
		if err != nil {
			return nil, nil, fmt.Errorf("fetch sprint offset %d: %w", offset, err)
		}
		var payload resultsResponse
		if err := json.Unmarshal(response.Body, &payload); err != nil {
			return nil, nil, fmt.Errorf("decode sprint: %w", err)
		}
		total, err := strconv.Atoi(payload.MRData.Total)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid sprint total")
		}
		if total == 0 {
			return []domain.SessionClassification{}, []ResultPage{{Endpoint: endpoint, Payload: response.Body}}, nil
		}
		records := 0
		for _, race := range payload.MRData.RaceTable.Races {
			item, err := convertSprint(race, year)
			if err != nil {
				return nil, nil, err
			}
			records += len(item.Results)
			if existing := byRound[item.Round]; existing != nil {
				existing.Results = append(existing.Results, item.Results...)
			} else {
				byRound[item.Round] = &item
			}
		}
		pages = append(pages, ResultPage{Endpoint: endpoint, Payload: response.Body, Records: records})
		if offset+records >= total {
			break
		}
		if records == 0 {
			return nil, nil, fmt.Errorf("empty sprint page at offset %d", offset)
		}
	}
	return finalizeClassifications(byRound, c.now().UTC()), pages, nil
}

func convertQualifying(raw qualifyingRace, year int) (domain.SessionClassification, error) {
	season, round, err := parseSeasonRound(raw.Season, raw.Round, year)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	item, err := classificationBase(season, round, "qualifying", raw.RaceName,
		raw.Circuit.ID, raw.Circuit.Name, raw.Circuit.Location.Latitude, raw.Circuit.Location.Longitude,
		raw.Circuit.Location.Locality, raw.Circuit.Location.Country)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	for _, result := range raw.Results {
		position, err := strconv.Atoi(result.Position)
		if err != nil {
			return domain.SessionClassification{}, err
		}
		item.Results = append(item.Results, domain.ClassificationResult{Position: position, DriverID: result.Driver.ID,
			DriverNumber: result.Number, DriverCode: result.Driver.Code, GivenName: result.Driver.GivenName,
			FamilyName: result.Driver.FamilyName, Nationality: result.Driver.Nationality, Q1: result.Q1, Q2: result.Q2, Q3: result.Q3,
			Constructor: domain.StandingConstructor{ID: result.Constructor.ID, Name: result.Constructor.Name, Nationality: result.Constructor.Nationality}})
	}
	return item, nil
}

func convertSprint(raw resultRace, year int) (domain.SessionClassification, error) {
	season, round, err := parseSeasonRound(raw.Season, raw.Round, year)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	item, err := classificationBase(season, round, "sprint", raw.RaceName,
		raw.Circuit.ID, raw.Circuit.Name, raw.Circuit.Location.Latitude, raw.Circuit.Location.Longitude,
		raw.Circuit.Location.Locality, raw.Circuit.Location.Country)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	for _, rawResult := range raw.SprintResults {
		result, err := convertResult(rawResult)
		if err != nil {
			return domain.SessionClassification{}, err
		}
		item.Results = append(item.Results, domain.ClassificationResult{Position: result.Position, DriverID: result.DriverID,
			DriverNumber: result.DriverNumber, DriverCode: result.DriverCode, GivenName: result.GivenName,
			FamilyName: result.FamilyName, Nationality: result.Nationality, Constructor: result.Constructor,
			Points: &result.Points, Grid: &result.Grid, Laps: &result.Laps, Status: result.Status, Time: result.Time,
			TimeMillis: result.TimeMillis, FastestLap: result.FastestLap})
	}
	return item, nil
}

func classificationBase(season, round int, kind, name, circuitID, circuitName, lat, long, locality, country string) (domain.SessionClassification, error) {
	latitude, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	longitude, err := strconv.ParseFloat(long, 64)
	if err != nil {
		return domain.SessionClassification{}, err
	}
	return domain.SessionClassification{Season: season, Round: round, Type: kind, Name: name,
		Circuit: domain.Circuit{ID: circuitID, Name: circuitName, Latitude: latitude, Longitude: longitude, Locality: locality, Country: country},
		Results: []domain.ClassificationResult{}, Source: "jolpica"}, nil
}

func finalizeClassifications(byRound map[int]*domain.SessionClassification, syncedAt time.Time) []domain.SessionClassification {
	items := make([]domain.SessionClassification, 0, len(byRound))
	for _, item := range byRound {
		item.SyncedAt = syncedAt
		sort.Slice(item.Results, func(i, j int) bool { return item.Results[i].Position < item.Results[j].Position })
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Round < items[j].Round })
	return items
}
