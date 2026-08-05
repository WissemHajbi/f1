package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestLaps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"driver_number":4,"lap_number":1,
			"date_start":"2025-03-02T14:01:00Z","lap_duration":90.5,"duration_sector_1":30.1,
			"duration_sector_2":29.9,"duration_sector_3":30.5,"i1_speed":250,"i2_speed":280,
			"st_speed":310,"segments_sector_1":[2049,2051],"segments_sector_2":[2049],
			"segments_sector_3":[2051],"is_pit_out_lap":false}]`))
	}))
	defer server.Close()
	driver := 4
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Laps(context.Background(), 200, &driver)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Laps) != 1 || result.Laps[0].LapDuration == nil || result.Laps[0].SpeedTrap == nil || *result.Laps[0].SpeedTrap != 310 {
		t.Fatalf("unexpected laps: %+v", result.Laps)
	}
}
