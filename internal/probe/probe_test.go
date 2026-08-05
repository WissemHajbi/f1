package probe

import "testing"

func TestParseJSONArray(t *testing.T) {
	count, _, err := parseJSONArray([]byte(`[{"id":1},{"id":2}]`))
	if err != nil || count != 2 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestParseJolpica(t *testing.T) {
	count, _, err := parseJolpica([]byte(`{"MRData":{"total":"24","limit":"100"}}`))
	if err != nil || count != 24 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestParseRSS(t *testing.T) {
	body := []byte(`<rss><channel><title>F1 News</title><item><title>A</title><link>https://example.test/a</link></item></channel></rss>`)
	count, _, err := parseRSS(body)
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

func TestParseFIA(t *testing.T) {
	body := []byte(`<html><body><a href="/document.pdf">2026 Car Display Submissions</a></body></html>`)
	count, _, err := parseFIA(body)
	if err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
