package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:10:00Z","session_key":200,"meeting_key":100,"driver_number":4,"x":120,"y":-35,"z":7}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Location(context.Background(), 200, 4, time.Date(2025, 3, 2, 14, 0, 0, 0, time.UTC), time.Date(2025, 3, 2, 14, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Samples) != 1 || result.Samples[0].X != 120 || result.Samples[0].Y != -35 {
		t.Fatalf("unexpected samples: %+v", result.Samples)
	}
}

func TestLocationRejectsLargeWindow(t *testing.T) {
	client := New(fetch.New(time.Second, "test/1.0", 1<<20))
	from := time.Now()
	if _, err := client.Location(context.Background(), 200, 4, from, from.Add(MaxLocationWindow+time.Second)); err == nil {
		t.Fatal("expected window error")
	}
}
