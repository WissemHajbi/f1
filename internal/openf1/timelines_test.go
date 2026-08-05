package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestTimelines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/overtakes":
			_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:10:00Z","session_key":200,"meeting_key":100,"overtaking_driver_number":4,"overtaken_driver_number":5,"position":3}]`))
		case "/position":
			_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:10:00Z","session_key":200,"meeting_key":100,"driver_number":4,"position":3}]`))
		case "/intervals":
			_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:10:00Z","session_key":200,"meeting_key":100,"driver_number":4,"gap_to_leader":2.5,"interval":"+1 LAP"}]`))
		}
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	overtakes, err := client.Overtakes(context.Background(), 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	positions, err := client.Positions(context.Background(), 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	intervals, err := client.Intervals(context.Background(), 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(overtakes.Events) != 1 || len(positions.Samples) != 1 || len(intervals.Samples) != 1 ||
		intervals.Samples[0].GapToLeaderSeconds == nil || intervals.Samples[0].Interval != "+1 LAP" {
		t.Fatalf("overtakes=%+v positions=%+v intervals=%+v", overtakes, positions, intervals)
	}
}
