package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestCarData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date>") == "" || r.URL.Query().Get("date<") == "" {
			t.Error("missing bounded dates")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:00:00.250Z","session_key":200,"meeting_key":100,"driver_number":4,"speed":250,"rpm":11000,"n_gear":7,"throttle":100,"brake":0,"drs":12}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	from := time.Date(2025, 3, 2, 14, 0, 0, 0, time.UTC)
	result, err := client.CarData(context.Background(), 200, 4, from, from.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Samples) != 1 || result.Samples[0].Gear != 7 || result.Samples[0].DRS != 12 {
		t.Fatalf("unexpected car data: %+v", result.Samples)
	}
}

func TestCarDataRejectsLargeWindow(t *testing.T) {
	client := New(fetch.New(time.Second, "test/1.0", 1<<20))
	from := time.Now()
	if _, err := client.CarData(context.Background(), 200, 4, from, from.Add(MaxCarDataWindow+time.Second)); err == nil {
		t.Fatal("expected range error")
	}
}
