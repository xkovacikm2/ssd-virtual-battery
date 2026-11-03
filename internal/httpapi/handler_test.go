package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/db"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/httpapi"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/readings"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/ssd"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/testdb"
)

var now = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

type fetcher struct {
	requested []string
	err       error
	empty     bool
}

func (f *fetcher) FetchDay(_ context.Context, day time.Time) ([]ssd.Interval, error) {
	f.requested = append(f.requested, day.Format(time.DateOnly))
	if f.err != nil || f.empty {
		return nil, f.err
	}
	return []ssd.Interval{{Incoming: 4, Outgoing: 8}}, nil
}

type env struct {
	server  *httptest.Server
	store   *readings.Store
	fetcher *fetcher
}

var static = fstest.MapFS{
	"index.html":           {Data: []byte("<!doctype html><svb-app></svb-app>")},
	"manifest.webmanifest": {Data: []byte(`{"name":"SSD Virtuálna batéria"}`)},
	"service-worker.js":    {Data: []byte("self.addEventListener('fetch', () => {})")},
	"assets/app-123.js":    {Data: []byte("console.log(1)")},
}

func setup(t *testing.T) env {
	t.Helper()
	pool := testdb.Empty(t)
	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	store := readings.NewStore(pool)
	f := &fetcher{}
	svc := &readings.Service{
		Store:      store,
		NewFetcher: func() (readings.Fetcher, error) { return f, nil },
		Now:        func() time.Time { return now },
	}
	srv := httptest.NewServer(httpapi.New(svc, pool, static))
	t.Cleanup(srv.Close)
	return env{server: srv, store: store, fetcher: f}
}

func (e env) seed(t *testing.T, date string, exported, imported readings.KWh) {
	t.Helper()
	day, _ := time.Parse(time.DateOnly, date)
	if err := e.store.Upsert(context.Background(), day, exported, imported); err != nil {
		t.Fatal(err)
	}
}

func get(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func TestDashboardSyncsMissingReadingsFirst(t *testing.T) {
	e := setup(t)
	e.seed(t, "2026-11-01", 100, 100)

	res, body := get(t, e.server.URL+"/api/dashboard")

	if res.StatusCode != 200 || !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		t.Fatalf("status %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	if len(e.fetcher.requested) != 1 || e.fetcher.requested[0] != "2026-11-02" {
		t.Fatalf("requested %v", e.fetcher.requested)
	}
	if !strings.Contains(body, `"total_exported_to_grid":3.00`) {
		t.Fatalf("summary should include synced day: %s", body)
	}
}

func TestDashboardPayloadShape(t *testing.T) {
	e := setup(t)
	e.seed(t, "2025-12-31", 2000, 800)
	e.seed(t, "2026-11-01", 1000, 400)
	e.seed(t, "2026-11-02", 2000, 600)

	_, body := get(t, e.server.URL+"/api/dashboard")

	want := `{
		"year": 2026,
		"max_capacity": 6000,
		"has_readings": true,
		"summary": {"total_exported_to_grid": 30.00, "total_imported_from_grid": 10.00, "net_grid_flow": 20.00},
		"chart_data": [
			{"date": "2026-11-01", "cumulative_exported": 10.00, "cumulative_imported": 4.00},
			{"date": "2026-11-02", "cumulative_exported": 30.00, "cumulative_imported": 10.00}
		],
		"previous_year": [{"date": "2025-12-31", "exported_to_grid": 20.00, "imported_from_grid": 8.00}]
	}`
	assertSameJSON(t, body, want)
}

func TestDashboardRendersWhenSyncFails(t *testing.T) {
	e := setup(t)
	e.seed(t, "2026-11-01", 2000, 1000)
	e.fetcher.err = errors.New("SSD down")

	res, body := get(t, e.server.URL+"/api/dashboard")

	if res.StatusCode != 200 || !strings.Contains(body, `"total_exported_to_grid":20.00`) {
		t.Fatalf("status %d body %s", res.StatusCode, body)
	}
}

func TestDashboardWithoutReadings(t *testing.T) {
	e := setup(t)
	e.fetcher.empty = true
	_, body := get(t, e.server.URL+"/api/dashboard")
	assertSameJSON(t, body, `{
		"year": 2026, "max_capacity": 6000, "has_readings": false,
		"summary": {"total_exported_to_grid": 0, "total_imported_from_grid": 0, "net_grid_flow": 0},
		"chart_data": [], "previous_year": []
	}`)
}

func TestReadingsDoesNotSync(t *testing.T) {
	e := setup(t)
	res, _ := get(t, e.server.URL+"/api/readings?tab=daily")
	if res.StatusCode != 200 || len(e.fetcher.requested) != 0 {
		t.Fatalf("status %d requested %v", res.StatusCode, e.fetcher.requested)
	}
}

func TestReadingsTabsAndPaging(t *testing.T) {
	e := setup(t)
	e.seed(t, "2026-11-02", 2550, 875)
	e.seed(t, "2026-10-01", 100, 100)

	_, body := get(t, e.server.URL+"/api/readings?tab=daily&page=1")
	var daily struct {
		Tab      string `json:"tab"`
		Page     int    `json:"page"`
		HasNewer bool   `json:"has_newer"`
		HasOlder bool   `json:"has_older"`
		Daily    []struct {
			Date string `json:"date"`
		} `json:"daily"`
	}
	if err := json.Unmarshal([]byte(body), &daily); err != nil {
		t.Fatal(err)
	}
	if daily.Tab != "daily" || daily.Page != 1 || !daily.HasNewer || !daily.HasOlder || len(daily.Daily) != 7 || daily.Daily[0].Date != "2026-10-26" {
		t.Fatalf("%s", body)
	}

	_, body = get(t, e.server.URL+"/api/readings?tab=monthly")
	assertSameJSON(t, body, `{"tab":"monthly","page":0,"has_newer":false,"has_older":false,"monthly":[
		{"month":11,"exported":25.50,"imported":8.75,"difference":16.75},
		{"month":10,"exported":1.00,"imported":1.00,"difference":0.00}]}`)

	_, body = get(t, e.server.URL+"/api/readings?tab=weekly")
	if !strings.Contains(body, `"weekly":[{"start_date":"2026-11-02","end_date":"2026-11-02"`) {
		t.Fatalf("%s", body)
	}
}

func TestReadingsClampsBadParams(t *testing.T) {
	e := setup(t)
	for _, q := range []string{"tab=invalid&page=-3", "page=abc", ""} {
		_, body := get(t, e.server.URL+"/api/readings?"+q)
		if !strings.Contains(body, `"tab":"daily","page":0,"has_newer":false`) {
			t.Errorf("%q => %s", q, body)
		}
	}
}

func TestHealthz(t *testing.T) {
	e := setup(t)
	res, body := get(t, e.server.URL+"/healthz")
	if res.StatusCode != 200 || body != "ok" {
		t.Fatalf("%d %q", res.StatusCode, body)
	}
}

func TestServesAppShellOnDashboardRoutes(t *testing.T) {
	e := setup(t)
	for _, path := range []string{"/", "/dashboard"} {
		res, body := get(t, e.server.URL+path)
		if res.StatusCode != 200 || !strings.Contains(body, "<svb-app>") || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
			t.Errorf("%s => %d %s", path, res.StatusCode, body)
		}
	}
}

func TestServesStaticFiles(t *testing.T) {
	e := setup(t)
	res, _ := get(t, e.server.URL+"/manifest.webmanifest")
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/manifest+json" {
		t.Fatalf("manifest %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	res, _ = get(t, e.server.URL+"/service-worker.js")
	if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("service worker %d %q", res.StatusCode, res.Header.Get("Cache-Control"))
	}
	res, _ = get(t, e.server.URL+"/assets/app-123.js")
	if res.StatusCode != 200 || !strings.Contains(res.Header.Get("Cache-Control"), "immutable") {
		t.Fatalf("asset %d %q", res.StatusCode, res.Header.Get("Cache-Control"))
	}
}

func TestUnknownPathsAre404(t *testing.T) {
	e := setup(t)
	for _, path := range []string{"/nope", "/api/nope", "/assets/missing.js"} {
		if res, _ := get(t, e.server.URL+path); res.StatusCode != 404 {
			t.Errorf("%s => %d", path, res.StatusCode)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	e := setup(t)
	res, _ := get(t, e.server.URL+"/")
	if res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
}

func assertSameJSON(t *testing.T, got, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("invalid JSON %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	if string(gb) != string(wb) {
		t.Fatalf("got  %s\nwant %s", gb, wb)
	}
}
