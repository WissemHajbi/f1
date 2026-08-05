package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestLatestCompletedRaceRoster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/sessions":
			_, _ = w.Write([]byte(`[
				{"session_key":1,"session_name":"Race","session_type":"Race","year":2025,"location":"Old","date_start":"2025-01-01T12:00:00Z","date_end":"2025-01-01T14:00:00Z"},
				{"session_key":2,"session_name":"Race","session_type":"Race","year":2025,"location":"Latest","date_start":"2025-02-01T12:00:00Z","date_end":"2025-02-01T14:00:00Z"}
			]`))
		case "/drivers":
			if r.URL.Query().Get("session_key") != "2" {
				t.Errorf("unexpected session key")
			}
			_, _ = w.Write([]byte(`[{"session_key":2,"driver_number":4,"full_name":"Test DRIVER","name_acronym":"TST","team_name":"Test Team"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: func() time.Time {
		return time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	}, baseURL: server.URL}
	result, err := client.LatestCompletedRaceRoster(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if result.Roster.Session.Key != 2 || len(result.Roster.Drivers) != 1 {
		t.Fatalf("unexpected result: %+v", result.Roster)
	}
}
