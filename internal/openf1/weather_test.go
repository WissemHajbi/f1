package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestWeather(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"date":"2025-03-02T14:00:00Z",
			"air_temperature":25.5,"track_temperature":35.2,"humidity":60,"pressure":1012.4,
			"rainfall":0,"wind_direction":180,"wind_speed":3.2}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Weather(context.Background(), 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Samples) != 1 || result.Samples[0].TrackTemperature == nil || *result.Samples[0].TrackTemperature != 35.2 {
		t.Fatalf("unexpected weather: %+v", result.Samples)
	}
}
