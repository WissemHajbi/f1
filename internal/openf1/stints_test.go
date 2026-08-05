package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestStints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"driver_number":4,"stint_number":1,
			"lap_start":1,"lap_end":20,"compound":"medium","tyre_age_at_start":2}]`))
	}))
	defer server.Close()
	driver := 4
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Stints(context.Background(), 200, &driver)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stints) != 1 || result.Stints[0].Compound != "MEDIUM" || result.Stints[0].LapEnd == nil {
		t.Fatalf("unexpected stints: %+v", result.Stints)
	}
}
