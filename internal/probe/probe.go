package probe

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"oidysts/internal/fetch"
	"oidysts/internal/store"
)

type Probe struct {
	client *fetch.Client
	now    func() time.Time
}

type target struct {
	source, resource, endpoint, accept string
	parse                              func([]byte) (int, any, error)
}

func New(client *fetch.Client) *Probe { return &Probe{client: client, now: time.Now} }

func (p *Probe) Run(ctx context.Context, sources []string, year int, staticPath string) ([]store.Snapshot, error) {
	selected := make(map[string]bool, len(sources))
	for _, source := range sources {
		selected[strings.ToLower(strings.TrimSpace(source))] = true
	}

	var out []store.Snapshot
	var failures []error
	lastRequest := map[string]time.Time{}
	for _, target := range targets(year) {
		if !selected["all"] && !selected[target.source] {
			continue
		}
		if err := waitForProvider(ctx, target.source, lastRequest[target.source]); err != nil {
			failures = append(failures, fmt.Errorf("%s/%s: %w", target.source, target.resource, err))
			continue
		}
		item, err := p.fetch(ctx, target)
		lastRequest[target.source] = time.Now()
		if err != nil {
			failures = append(failures, fmt.Errorf("%s/%s: %w", target.source, target.resource, err))
			continue
		}
		out = append(out, item)
	}
	if selected["all"] || selected["static"] {
		item, err := p.readStatic(staticPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("static/teams: %w", err))
		} else {
			out = append(out, item)
		}
	}
	return out, errors.Join(failures...)
}

func waitForProvider(ctx context.Context, source string, previous time.Time) error {
	minimumGap := map[string]time.Duration{"openf1": 350 * time.Millisecond, "jolpica": 260 * time.Millisecond}[source]
	remaining := minimumGap - time.Since(previous)
	if previous.IsZero() || remaining <= 0 {
		return nil
	}
	timer := time.NewTimer(remaining)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (p *Probe) fetch(ctx context.Context, target target) (store.Snapshot, error) {
	response, err := p.client.Get(ctx, target.endpoint, target.accept)
	if err != nil {
		return store.Snapshot{}, err
	}
	count, summary, err := target.parse(response.Body)
	if err != nil {
		return store.Snapshot{}, fmt.Errorf("validate payload: %w", err)
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return store.Snapshot{}, fmt.Errorf("encode summary: %w", err)
	}
	return store.Snapshot{
		Source: target.source, Resource: target.resource, Endpoint: target.endpoint,
		FetchedAt: p.now().UTC(), StatusCode: response.StatusCode, ContentType: response.ContentType,
		RecordCount: count, PayloadBytes: len(response.Body), Summary: summaryJSON, Payload: response.Body,
	}, nil
}

func (p *Probe) readStatic(path string) (store.Snapshot, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return store.Snapshot{}, err
	}
	var doc struct {
		Season int               `json:"season"`
		Teams  []json.RawMessage `json:"teams"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return store.Snapshot{}, err
	}
	if doc.Season == 0 || doc.Teams == nil {
		return store.Snapshot{}, fmt.Errorf("season and teams are required")
	}
	summary, _ := json.Marshal(map[string]any{"season": doc.Season, "teams": len(doc.Teams)})
	return store.Snapshot{Source: "static", Resource: "teams", Endpoint: path, FetchedAt: p.now().UTC(),
		StatusCode: 200, ContentType: "application/json", RecordCount: len(doc.Teams), PayloadBytes: len(body), Summary: summary, Payload: body}, nil
}

func targets(year int) []target {
	openF1 := "https://api.openf1.org/v1"
	jolpica := "https://api.jolpi.ca/ergast/f1"
	y := strconv.Itoa(year)
	return []target{
		{"openf1", "sessions", openF1 + "/sessions?year=" + y + "&session_name=Race", "application/json", parseJSONArray},
		{"openf1", "drivers", openF1 + "/drivers?session_key=9693&driver_number=1", "application/json", parseJSONArray},
		{"openf1", "lap_sample", openF1 + "/laps?session_key=9693&driver_number=1&lap_number=1", "application/json", parseJSONArray},
		{"openf1", "telemetry_sample", openF1 + "/car_data?session_key=9693&driver_number=1&date%3E2025-03-16T04:03:00&date%3C2025-03-16T04:03:02", "application/json", parseJSONArray},
		{"openf1", "location_sample", openF1 + "/location?session_key=9693&driver_number=1&date%3E2025-03-16T04:03:00&date%3C2025-03-16T04:03:02", "application/json", parseJSONArray},
		{"jolpica", "calendar", jolpica + "/" + y + ".json?limit=100", "application/json", parseJolpica},
		{"jolpica", "driver_standings", jolpica + "/" + y + "/driverstandings.json?limit=100", "application/json", parseJolpica},
		{"jolpica", "constructor_standings", jolpica + "/" + y + "/constructorstandings.json?limit=100", "application/json", parseJolpica},
		{"rss", "motorsport", "https://www.motorsport.com/rss/f1/news/", "application/rss+xml, application/xml, text/xml", parseRSS},
		{"rss", "sky", "https://www.skysports.com/rss/12040", "application/rss+xml, application/xml, text/xml", parseRSS},
		{"rss", "bbc", "https://feeds.bbci.co.uk/sport/formula1/rss.xml", "application/rss+xml, application/xml, text/xml", parseRSS},
		{"fia", "documents", "https://www.fia.com/documents/championships/fia-formula-one-world-championship-14", "text/html", parseFIA},
	}
}

func parseJSONArray(body []byte) (int, any, error) {
	var rows []json.RawMessage
	if err := json.Unmarshal(body, &rows); err != nil {
		return 0, nil, err
	}
	return len(rows), map[string]any{"records": len(rows)}, nil
}

func parseJolpica(body []byte) (int, any, error) {
	var doc struct {
		MRData struct {
			Total string `json:"total"`
			Limit string `json:"limit"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return 0, nil, err
	}
	total, err := strconv.Atoi(doc.MRData.Total)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid MRData.total %q", doc.MRData.Total)
	}
	return total, map[string]any{"total": total, "limit": doc.MRData.Limit}, nil
}

type rssDocument struct {
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
		} `xml:"item"`
	} `xml:"channel"`
}

func parseRSS(body []byte) (int, any, error) {
	var doc rssDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return 0, nil, err
	}
	if strings.TrimSpace(doc.Channel.Title) == "" {
		return 0, nil, fmt.Errorf("missing channel title")
	}
	return len(doc.Channel.Items), map[string]any{"channel": doc.Channel.Title, "items": len(doc.Channel.Items)}, nil
}

func parseFIA(body []byte) (int, any, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return 0, nil, err
	}
	links, matches := 0, 0
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			var href string
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
			}
			if href != "" {
				links++
				text := strings.ToLower(nodeText(node) + " " + href)
				if strings.Contains(text, "car display") || strings.Contains(text, "automobile display") {
					matches++
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return matches, map[string]any{"links": links, "candidate_documents": matches, "page": sanitizeURL("https://www.fia.com")}, nil
}

func nodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var b strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(nodeText(child))
		b.WriteByte(' ')
	}
	return b.String()
}

func sanitizeURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return parsed.Scheme + "://" + parsed.Host
}
