// Package ssd is a client for the ims.ssd.sk metering portal.
package ssd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const DefaultBaseURL = "https://ims.ssd.sk"

type Config struct {
	BaseURL           string
	Username          string
	Password          string
	PointOfDeliveryID string
}

// Interval is one 15-minute metering period in kW.
type Interval struct {
	Incoming float64
	Outgoing float64
}

type Client struct {
	baseURL string
	cfg     Config
}

func NewClient(cfg Config) (*Client, error) {
	var missing []error
	if cfg.Username == "" {
		missing = append(missing, errors.New("IMS_SSD_USERNAME environment variable is required"))
	}
	if cfg.Password == "" {
		missing = append(missing, errors.New("IMS_SSD_PASSWORD environment variable is required"))
	}
	if cfg.PointOfDeliveryID == "" {
		missing = append(missing, errors.New("IMS_SSD_POINT_OF_DELIVERY_ID environment variable is required"))
	}
	if err := errors.Join(missing...); err != nil {
		return nil, err
	}
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{baseURL: base, cfg: cfg}, nil
}

// FetchDay logs in and returns the metering intervals of the given UTC day.
func (c *Client) FetchDay(ctx context.Context, day time.Time) ([]Interval, error) {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{Jar: jar, Timeout: 20 * time.Second}

	login := map[string]string{"username": c.cfg.Username, "password": c.cfg.Password}
	if _, err := c.post(ctx, httpClient, "/api/account/login", "/login", nil, login); err != nil {
		return nil, err
	}

	screen := map[string]string{
		"X-Client-Screen-Number": "CP.015",
		"X-Client-Url":           "/consumption-production/profile-data",
	}
	raw, err := c.post(ctx, httpClient, "/api/consumption-production/profile-data",
		"/consumption-production/profile-data", screen, c.profileRequest(day))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Rows []struct {
			Values []json.RawMessage `json:"values"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode profile data: %w", err)
	}
	out := make([]Interval, 0, len(resp.Rows))
	for _, r := range resp.Rows {
		out = append(out, Interval{Incoming: number(r.Values, 2), Outgoing: number(r.Values, 4)})
	}
	return out, nil
}

func number(values []json.RawMessage, i int) float64 {
	if i >= len(values) {
		return 0
	}
	var f float64
	if json.Unmarshal(values[i], &f) == nil {
		return f
	}
	var s string
	if json.Unmarshal(values[i], &s) == nil {
		_, _ = fmt.Sscan(s, &f)
	}
	return f
}

func (c *Client) post(ctx context.Context, hc *http.Client, path, referer string, extra map[string]string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+referer)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	for k, v := range extra {
		req.Header.Set(k, v)
	}

	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("HTTP Error: %d - %s\nBody: %s", res.StatusCode, http.StatusText(res.StatusCode), data)
	}
	return data, nil
}

type column struct {
	Member       string `json:"member"`
	Title        string `json:"title,omitempty"`
	Type         string `json:"type"`
	IsVisible    bool   `json:"isVisible"`
	IsSortable   bool   `json:"isSortable"`
	IsFilterable bool   `json:"isFilterable"`
	IsNullable   bool   `json:"isNullable"`
	Index        int    `json:"index"`
	AllowVisible bool   `json:"allowVisible"`
}

func visible(i int, member, title, typ string, filterable, nullable bool) column {
	return column{Member: member, Title: title, Type: typ, IsVisible: true, IsSortable: true,
		IsFilterable: filterable, IsNullable: nullable, Index: i, AllowVisible: true}
}

func hidden(i int, member string) column {
	return column{Member: member, Type: "UNDEFINED", Index: i}
}

var profileColumns = []column{
	visible(0, "meteringDatetime", "Dátum a čas merania", "DateTime", false, false),
	visible(1, "period", "Perióda", "Int", true, false),
	visible(2, "actualConsumption", "1.5.0 - Činný odber (kW)", "Float", true, true),
	visible(3, "actualConsumptionQualityType", "1.5.0 - Činný odber (kvalita)", "Enumeration", true, true),
	visible(4, "actualSupply", "2.5.0 - Činná dodávka (kW)", "Float", true, true),
	visible(5, "actualSupplyQualityType", "2.5.0 - Činná dodávka (kvalita)", "Enumeration", true, true),
	visible(6, "idleConsumption", "3.5.0 - Jalový odber (kVAR)", "Float", true, true),
	visible(7, "idleConsumptionQualityType", "3.5.0 - Jalový odber (kvalita)", "Enumeration", true, true),
	visible(8, "idleSupply", "4.5.0 - Jalová dodávka (kVAR)", "Float", true, true),
	visible(9, "idleSupplyQualityType", "4.5.0 - Jalová dodávka (kvalita)", "Enumeration", true, true),
	hidden(10, "profileDataId"),
	hidden(11, "pointOfDeliveryId"),
	hidden(12, "startPeriodDatetime"),
	hidden(13, "userId"),
	hidden(14, "changedOn"),
	hidden(15, "deletedOn"),
}

func (c *Client) profileRequest(day time.Time) map[string]any {
	const layout = "2006-01-02T15:04:05.000Z"
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	from := start.Format(layout)
	to := start.AddDate(0, 0, 1).Format(layout)
	return map[string]any{
		"page":    map[string]int{"totalRows": 0, "currentPage": 1, "pageSize": 100},
		"columns": profileColumns,
		"filters": []map[string]any{
			{"member": "pointOfDeliveryId", "operator": "Equals", "type": "Int", "value": c.cfg.PointOfDeliveryID},
			{"member": "meteringDatetime", "operator": "Greater", "type": "DateTimeMilliseconds",
				"value": from, "rangeOperator": "LowerOrEquals", "rangeValue": to},
			{"member": "meteringDatetime", "operator": "LowerOrEquals", "type": "DateTimeMilliseconds",
				"value": to, "rangeOperator": nil, "rangeValue": nil},
		},
		"sort":     []map[string]string{{"member": "meteringDatetime", "sortOrder": "asc"}},
		"isExport": false,
	}
}
