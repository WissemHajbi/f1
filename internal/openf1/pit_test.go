package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestPitStops(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"driver_number":4,
			"lap_number":20,"date":"2025-03-02T14:30:00.250Z","pit_duration":22.8}]`))
	}))
	defer server.Close()
	driver := 4
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.PitStops(context.Background(), 200, &driver)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.PitStops) != 1 || result.PitStops[0].Duration == nil || *result.PitStops[0].Duration != 22.8 {
		t.Fatalf("unexpected pit stops: %+v", result.PitStops)
	}
}
