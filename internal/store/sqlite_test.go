package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestSaveAndLatest(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	item := Snapshot{Source: "test", Resource: "sample", Endpoint: "https://example.test", FetchedAt: time.Now(),
		StatusCode: 200, ContentType: "application/json", RecordCount: 2, PayloadBytes: 2,
		Summary: json.RawMessage(`{"records":2}`), Payload: []byte(`[]`)}
	if err := db.Save(ctx, item); err != nil {
		t.Fatal(err)
	}
	items, err := db.Latest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].RecordCount != 2 {
		t.Fatalf("unexpected items: %+v", items)
	}
}
