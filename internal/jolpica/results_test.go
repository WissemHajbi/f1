package jolpica

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"MRData":{"total":"1","RaceTable":{"Races":[{
			"season":"2025","round":"1","raceName":"Test Grand Prix","date":"2025-03-02","time":"14:00:00Z",
			"Circuit":{"circuitId":"test","circuitName":"Test Circuit","Location":{"lat":"1.2","long":"3.4","locality":"City","country":"Country"}},
			"Results":[{"number":"4","position":"1","positionText":"1","points":"25","grid":"2","laps":"50","status":"Finished",
			"Driver":{"driverId":"test","code":"TST","givenName":"Test","familyName":"Driver","nationality":"Test"},
			"Constructor":{"constructorId":"team","name":"Team","nationality":"Test"},
			"Time":{"millis":"5400000","time":"1:30:00.000"},
			"FastestLap":{"rank":"1","lap":"42","Time":{"time":"1:20.000"},"AverageSpeed":{"units":"kph","speed":"210.5"}}
			}]}]}}}`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Results(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Races) != 1 || len(result.Races[0].Results) != 1 || result.Races[0].Results[0].FastestLap == nil {
		t.Fatalf("unexpected results: %+v", result)
	}
}
