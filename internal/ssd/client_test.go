package ssd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type recorded struct {
	path    string
	headers http.Header
	body    map[string]any
}

type fakeSSD struct {
	*httptest.Server
	mu          sync.Mutex
	requests    []recorded
	loginStatus int
	profileCode int
	profileBody string
	rows        [][]any
}

func newFakeSSD(t *testing.T) *fakeSSD {
	t.Helper()
	f := &fakeSSD{loginStatus: 200, profileCode: 200}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		f.mu.Lock()
		f.requests = append(f.requests, recorded{path: r.URL.Path, headers: r.Header.Clone(), body: body})
		f.mu.Unlock()

		switch r.URL.Path {
		case "/api/account/login":
			if f.loginStatus != 200 {
				w.WriteHeader(f.loginStatus)
				_, _ = io.WriteString(w, "bad creds")
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123", Path: "/", HttpOnly: true})
			http.SetCookie(w, &http.Cookie{Name: "XSRF", Value: "tok", Path: "/"})
			_, _ = io.WriteString(w, "{}")
		case "/api/consumption-production/profile-data":
			if f.profileCode != 200 {
				w.WriteHeader(f.profileCode)
				_, _ = io.WriteString(w, f.profileBody)
				return
			}
			rows := make([]map[string]any, 0, len(f.rows))
			for _, v := range f.rows {
				rows = append(rows, map[string]any{"values": v})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"rows": rows})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(f.Close)
	return f
}

func (f *fakeSSD) request(t *testing.T, path string) recorded {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.requests {
		if r.path == path {
			return r
		}
	}
	t.Fatalf("no request to %s", path)
	return recorded{}
}

func row(incoming, outgoing any) []any {
	return []any{"2026-03-15T00:15:00Z", 1, incoming, "Real", outgoing, "Real", 0, nil, 0, nil}
}

func newClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{BaseURL: baseURL, Username: "user@example.com", Password: "secret", PointOfDeliveryID: "4242"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var day = time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

func TestNewClientRequiresCredentials(t *testing.T) {
	full := Config{BaseURL: "http://x", Username: "u", Password: "p", PointOfDeliveryID: "1"}
	cases := map[string]func(*Config){
		"IMS_SSD_USERNAME":             func(c *Config) { c.Username = "" },
		"IMS_SSD_PASSWORD":             func(c *Config) { c.Password = "" },
		"IMS_SSD_POINT_OF_DELIVERY_ID": func(c *Config) { c.PointOfDeliveryID = "" },
	}
	for name, mutate := range cases {
		cfg := full
		mutate(&cfg)
		if _, err := NewClient(cfg); err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s: want error mentioning it, got %v", name, err)
		}
	}
}

func TestNewClientDefaultsBaseURL(t *testing.T) {
	c, err := NewClient(Config{Username: "u", Password: "p", PointOfDeliveryID: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != DefaultBaseURL || DefaultBaseURL != "https://ims.ssd.sk" {
		t.Fatalf("base URL = %q", c.baseURL)
	}
}

func TestLoginSendsJSONCredentialsAndBrowserHeaders(t *testing.T) {
	f := newFakeSSD(t)
	if _, err := newClient(t, f.URL).FetchDay(context.Background(), day); err != nil {
		t.Fatal(err)
	}

	r := f.request(t, "/api/account/login")
	if r.body["username"] != "user@example.com" || r.body["password"] != "secret" || len(r.body) != 2 {
		t.Errorf("login body = %v", r.body)
	}
	want := map[string]string{
		"Content-Type":     "application/json",
		"Accept":           "application/json",
		"X-Requested-With": "XMLHttpRequest",
		"Origin":           f.URL,
		"Referer":          f.URL + "/login",
	}
	for k, v := range want {
		if got := r.headers.Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestProfileRequestCarriesCookiesAndScreenHeaders(t *testing.T) {
	f := newFakeSSD(t)
	if _, err := newClient(t, f.URL).FetchDay(context.Background(), day); err != nil {
		t.Fatal(err)
	}

	r := f.request(t, "/api/consumption-production/profile-data")
	cookie := r.headers.Get("Cookie")
	if !strings.Contains(cookie, "session=abc123") || !strings.Contains(cookie, "XSRF=tok") {
		t.Errorf("Cookie = %q", cookie)
	}
	want := map[string]string{
		"X-Client-Screen-Number": "CP.015",
		"X-Client-Url":           "/consumption-production/profile-data",
		"Referer":                f.URL + "/consumption-production/profile-data",
		"Origin":                 f.URL,
		"X-Requested-With":       "XMLHttpRequest",
	}
	for k, v := range want {
		if got := r.headers.Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestProfileRequestFiltersByPointOfDeliveryAndUTCDay(t *testing.T) {
	f := newFakeSSD(t)
	if _, err := newClient(t, f.URL).FetchDay(context.Background(), day); err != nil {
		t.Fatal(err)
	}

	body := f.request(t, "/api/consumption-production/profile-data").body
	assertJSON(t, "page", body["page"], `{"totalRows":0,"currentPage":1,"pageSize":100}`)
	assertJSON(t, "sort", body["sort"], `[{"member":"meteringDatetime","sortOrder":"asc"}]`)
	assertJSON(t, "isExport", body["isExport"], `false`)
	assertJSON(t, "filters", body["filters"], `[
		{"member":"pointOfDeliveryId","operator":"Equals","type":"Int","value":"4242"},
		{"member":"meteringDatetime","operator":"Greater","type":"DateTimeMilliseconds",
		 "value":"2026-03-15T00:00:00.000Z","rangeOperator":"LowerOrEquals","rangeValue":"2026-03-16T00:00:00.000Z"},
		{"member":"meteringDatetime","operator":"LowerOrEquals","type":"DateTimeMilliseconds",
		 "value":"2026-03-16T00:00:00.000Z","rangeOperator":null,"rangeValue":null}]`)

	cols, _ := body["columns"].([]any)
	if len(cols) != 16 {
		t.Fatalf("columns = %d, want 16", len(cols))
	}
	wantMembers := []string{"meteringDatetime", "period", "actualConsumption", "actualConsumptionQualityType", "actualSupply"}
	for i, m := range wantMembers {
		if got := cols[i].(map[string]any)["member"]; got != m {
			t.Errorf("column %d = %v, want %s", i, got, m)
		}
	}
}

func TestFetchDayMapsConsumptionAndSupply(t *testing.T) {
	f := newFakeSSD(t)
	f.rows = [][]any{row(1.5, 0.25), row(nil, 3.0)}

	got, err := newClient(t, f.URL).FetchDay(context.Background(), day)
	if err != nil {
		t.Fatal(err)
	}
	want := []Interval{{Incoming: 1.5, Outgoing: 0.25}, {Incoming: 0, Outgoing: 3.0}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestFetchDayEmpty(t *testing.T) {
	f := newFakeSSD(t)
	got, err := newClient(t, f.URL).FetchDay(context.Background(), day)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLoginFailureIncludesStatusAndBody(t *testing.T) {
	f := newFakeSSD(t)
	f.loginStatus = 401
	_, err := newClient(t, f.URL).FetchDay(context.Background(), day)
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "bad creds") {
		t.Fatalf("err = %v", err)
	}
}

func TestProfileFailureIncludesStatusAndBody(t *testing.T) {
	f := newFakeSSD(t)
	f.profileCode, f.profileBody = 500, "boom"
	_, err := newClient(t, f.URL).FetchDay(context.Background(), day)
	if err == nil || !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
}

func assertJSON(t *testing.T, name string, got any, want string) {
	t.Helper()
	var w any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	g, _ := json.Marshal(got)
	wb, _ := json.Marshal(w)
	if string(g) != string(wb) {
		t.Errorf("%s =\n %s\nwant\n %s", name, g, wb)
	}
}
