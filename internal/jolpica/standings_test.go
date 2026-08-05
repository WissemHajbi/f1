package jolpica

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestStandings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "driverstandings") {
			_, _ = w.Write([]byte(`{"MRData":{"StandingsTable":{"StandingsLists":[{"season":"2025","round":"2","DriverStandings":[{
				"position":"1","points":"44.5","wins":"1","Driver":{"driverId":"test","permanentNumber":"4","code":"TST","givenName":"Test","familyName":"Driver","dateOfBirth":"2000-01-01","nationality":"Test"},
				"Constructors":[{"constructorId":"team","name":"Team","nationality":"Test"}]
			}]}]}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"MRData":{"StandingsTable":{"StandingsLists":[{"season":"2025","round":"2","ConstructorStandings":[{
			"position":"1","points":"80","wins":"2","Constructor":{"constructorId":"team","name":"Team","nationality":"Test"}
		}]}]}}}`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Standings(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Drivers) != 1 || result.Drivers[0].Points != 44.5 || len(result.Constructors) != 1 {
		t.Fatalf("unexpected standings: %+v", result)
	}
}
