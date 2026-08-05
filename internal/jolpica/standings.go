package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"oidysts/internal/domain"
)

type StandingsResult struct {
	Drivers              []domain.DriverStanding
	Constructors         []domain.ConstructorStanding
	DriversEndpoint      string
	DriversPayload       []byte
	ConstructorsEndpoint string
	ConstructorsPayload  []byte
}

type rawConstructor struct {
	ID          string `json:"constructorId"`
	Name        string `json:"name"`
	Nationality string `json:"nationality"`
}

type standingsList struct {
	Season          string `json:"season"`
	Round           string `json:"round"`
	DriverStandings []struct {
		Position string `json:"position"`
		Points   string `json:"points"`
		Wins     string `json:"wins"`
		Driver   struct {
			ID              string `json:"driverId"`
			PermanentNumber string `json:"permanentNumber"`
			Code            string `json:"code"`
			GivenName       string `json:"givenName"`
			FamilyName      string `json:"familyName"`
			DateOfBirth     string `json:"dateOfBirth"`
			Nationality     string `json:"nationality"`
		} `json:"Driver"`
		Constructors []rawConstructor `json:"Constructors"`
	} `json:"DriverStandings"`
	ConstructorStandings []struct {
		Position    string         `json:"position"`
		Points      string         `json:"points"`
		Wins        string         `json:"wins"`
		Constructor rawConstructor `json:"Constructor"`
	} `json:"ConstructorStandings"`
}

type standingsResponse struct {
	MRData struct {
		StandingsTable struct {
			Lists []standingsList `json:"StandingsLists"`
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

func (c *Client) Standings(ctx context.Context, year int) (StandingsResult, error) {
	driversEndpoint := c.baseURL + "/" + strconv.Itoa(year) + "/driverstandings.json?limit=100"
	driverResponse, err := c.fetch.Get(ctx, driversEndpoint, "application/json")
	if err != nil {
		return StandingsResult{}, fmt.Errorf("fetch driver standings: %w", err)
	}
	if err := wait(ctx, 260*time.Millisecond); err != nil {
		return StandingsResult{}, err
	}
	constructorsEndpoint := c.baseURL + "/" + strconv.Itoa(year) + "/constructorstandings.json?limit=100"
	constructorResponse, err := c.fetch.Get(ctx, constructorsEndpoint, "application/json")
	if err != nil {
		return StandingsResult{}, fmt.Errorf("fetch constructor standings: %w", err)
	}

	var driverPayload, constructorPayload standingsResponse
	if err := json.Unmarshal(driverResponse.Body, &driverPayload); err != nil {
		return StandingsResult{}, fmt.Errorf("decode driver standings: %w", err)
	}
	if err := json.Unmarshal(constructorResponse.Body, &constructorPayload); err != nil {
		return StandingsResult{}, fmt.Errorf("decode constructor standings: %w", err)
	}
	syncedAt := c.now().UTC()
	drivers, err := convertDriverStandings(driverPayload, year, syncedAt)
	if err != nil {
		return StandingsResult{}, err
	}
	constructors, err := convertConstructorStandings(constructorPayload, year, syncedAt)
	if err != nil {
		return StandingsResult{}, err
	}
	return StandingsResult{Drivers: drivers, Constructors: constructors, DriversEndpoint: driversEndpoint,
		DriversPayload: driverResponse.Body, ConstructorsEndpoint: constructorsEndpoint, ConstructorsPayload: constructorResponse.Body}, nil
}

func convertDriverStandings(payload standingsResponse, year int, syncedAt time.Time) ([]domain.DriverStanding, error) {
	if len(payload.MRData.StandingsTable.Lists) == 0 {
		return nil, fmt.Errorf("driver standings returned no list for %d", year)
	}
	list := payload.MRData.StandingsTable.Lists[0]
	season, round, err := parseSeasonRound(list.Season, list.Round, year)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DriverStanding, 0, len(list.DriverStandings))
	for _, item := range list.DriverStandings {
		position, points, wins, err := parseStandingNumbers(item.Position, item.Points, item.Wins)
		if err != nil {
			return nil, fmt.Errorf("driver %s: %w", item.Driver.ID, err)
		}
		constructors := make([]domain.StandingConstructor, 0, len(item.Constructors))
		for _, constructor := range item.Constructors {
			constructors = append(constructors, domain.StandingConstructor{ID: constructor.ID, Name: constructor.Name, Nationality: constructor.Nationality})
		}
		out = append(out, domain.DriverStanding{Season: season, Round: round, Position: position, Points: points, Wins: wins,
			DriverID: item.Driver.ID, PermanentNumber: item.Driver.PermanentNumber, Code: item.Driver.Code,
			GivenName: item.Driver.GivenName, FamilyName: item.Driver.FamilyName, DateOfBirth: item.Driver.DateOfBirth,
			Nationality: item.Driver.Nationality, Constructors: constructors, SyncedAt: syncedAt})
	}
	return out, nil
}

func convertConstructorStandings(payload standingsResponse, year int, syncedAt time.Time) ([]domain.ConstructorStanding, error) {
	if len(payload.MRData.StandingsTable.Lists) == 0 {
		return nil, fmt.Errorf("constructor standings returned no list for %d", year)
	}
	list := payload.MRData.StandingsTable.Lists[0]
	season, round, err := parseSeasonRound(list.Season, list.Round, year)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ConstructorStanding, 0, len(list.ConstructorStandings))
	for _, item := range list.ConstructorStandings {
		position, points, wins, err := parseStandingNumbers(item.Position, item.Points, item.Wins)
		if err != nil {
			return nil, fmt.Errorf("constructor %s: %w", item.Constructor.ID, err)
		}
		out = append(out, domain.ConstructorStanding{Season: season, Round: round, Position: position, Points: points, Wins: wins,
			Constructor: domain.StandingConstructor{ID: item.Constructor.ID, Name: item.Constructor.Name, Nationality: item.Constructor.Nationality}, SyncedAt: syncedAt})
	}
	return out, nil
}

func parseSeasonRound(seasonValue, roundValue string, expected int) (int, int, error) {
	season, err := strconv.Atoi(seasonValue)
	if err != nil || season != expected {
		return 0, 0, fmt.Errorf("invalid season %q", seasonValue)
	}
	round, err := strconv.Atoi(roundValue)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid round %q", roundValue)
	}
	return season, round, nil
}

func parseStandingNumbers(positionValue, pointsValue, winsValue string) (int, float64, int, error) {
	position, err := strconv.Atoi(positionValue)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid position %q", positionValue)
	}
	points, err := strconv.ParseFloat(pointsValue, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid points %q", pointsValue)
	}
	wins, err := strconv.Atoi(winsValue)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid wins %q", winsValue)
	}
	return position, points, wins, nil
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
