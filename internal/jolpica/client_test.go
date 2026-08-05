package jolpica

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestCalendar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"MRData":{"total":"1","RaceTable":{"Races":[{
			"season":"2025","round":"1","url":"https://example.test/race","raceName":"Test Grand Prix",
			"Circuit":{"circuitId":"test","circuitName":"Test Circuit","Location":{"lat":"1.2","long":"3.4","locality":"City","country":"Country"}},
			"date":"2025-03-02","time":"14:00:00Z","FirstPractice":{"date":"2025-02-28","time":"12:00:00Z"},
			"Qualifying":{"date":"2025-03-01","time":"14:00:00Z"}
		}]}}}`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Calendar(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 1 || result.Events[0].Name != "Test Grand Prix" || len(result.Events[0].Sessions) != 3 {
		t.Fatalf("unexpected calendar: %+v", result.Events)
	}
}
