package openf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oidysts/internal/fetch"
)

func TestMeetings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/meetings" {
			_, _ = w.Write([]byte(`[{"meeting_key":100,"meeting_name":"Test Grand Prix","meeting_official_name":"Test Formula 1 Grand Prix","year":2025,"location":"Test City","country_key":1,"country_code":"TST","country_name":"Test","circuit_key":2,"circuit_short_name":"Test Circuit","date_start":"2025-03-01T00:00:00Z","gmt_offset":"00:00:00"}]`))
			return
		}
		_, _ = w.Write([]byte(`[{"session_key":200,"meeting_key":100,"session_name":"Race","session_type":"Race","year":2025,"location":"Test City","country_code":"TST","country_name":"Test","circuit_key":2,"circuit_short_name":"Test Circuit","date_start":"2025-03-02T14:00:00Z","date_end":"2025-03-02T16:00:00Z","gmt_offset":"00:00:00","is_cancelled":false}]`))
	}))
	defer server.Close()
	client := &Client{fetch: fetch.New(time.Second, "test/1.0", 1<<20), now: time.Now, baseURL: server.URL}
	result, err := client.Meetings(context.Background(), 2025)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Meetings) != 1 || len(result.Meetings[0].Sessions) != 1 || result.Meetings[0].Sessions[0].Key != 200 {
		t.Fatalf("unexpected meetings: %+v", result.Meetings)
	}
}
