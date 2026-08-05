package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestTeamRadio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"date":"2025-03-02T14:10:00Z","session_key":200,"meeting_key":100,"driver_number":4,"recording_url":"https://media.example/radio.mp3"}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.TeamRadio(context.Background(), 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 1 || len(result.Records[0].ID) != 64 || result.Records[0].DriverNumber != 4 {
		t.Fatalf("unexpected records: %+v", result.Records)
	}
}
