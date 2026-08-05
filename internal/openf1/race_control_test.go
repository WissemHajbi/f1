package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestRaceControl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"date":"2025-03-02T14:30:00Z",
			"category":"Flag","message":"YELLOW IN TRACK SECTOR 2","flag":"YELLOW","scope":"Sector",
			"driver_number":null,"lap_number":20,"sector":2}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.RaceControl(context.Background(), 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 1 || result.Events[0].Flag != "YELLOW" || result.Events[0].Sector == nil {
		t.Fatalf("unexpected events: %+v", result.Events)
	}
}
