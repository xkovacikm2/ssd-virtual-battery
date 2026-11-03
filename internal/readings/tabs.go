package readings

import (
	"context"
	"time"
)

var tabs = map[string]bool{"daily": true, "weekly": true, "monthly": true}

type DailyRow struct {
	Date       string `json:"date"`
	Exported   KWh    `json:"exported"`
	Imported   KWh    `json:"imported"`
	Difference KWh    `json:"difference"`
}

type WeeklyRow struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Exported   KWh    `json:"exported"`
	Imported   KWh    `json:"imported"`
	Difference KWh    `json:"difference"`
}

type MonthlyRow struct {
	Month      int `json:"month"`
	Exported   KWh `json:"exported"`
	Imported   KWh `json:"imported"`
	Difference KWh `json:"difference"`
}

type ReadingsPage struct {
	Tab      string       `json:"tab"`
	Page     int          `json:"page"`
	HasNewer bool         `json:"has_newer"`
	HasOlder bool         `json:"has_older"`
	Daily    []DailyRow   `json:"daily,omitempty"`
	Weekly   []WeeklyRow  `json:"weekly,omitempty"`
	Monthly  []MonthlyRow `json:"monthly,omitempty"`
}

func (s *Service) Readings(ctx context.Context, tab string, page int) (ReadingsPage, error) {
	if !tabs[tab] {
		tab = "daily"
	}
	page = max(page, 0)
	out := ReadingsPage{Tab: tab, Page: page}
	var err error
	switch tab {
	case "daily":
		err = s.daily(ctx, &out)
	case "weekly":
		out.Page = 0
		out.Weekly, err = s.weekly(ctx)
	case "monthly":
		out.Page = 0
		out.Monthly, err = s.monthly(ctx)
	}
	return out, err
}

func (s *Service) daily(ctx context.Context, out *ReadingsPage) error {
	end := s.yesterday().AddDate(0, 0, -7*out.Page)
	start := end.AddDate(0, 0, -6)
	rs, err := s.Store.Range(ctx, start, end)
	if err != nil {
		return err
	}
	byDate := make(map[string]Reading, len(rs))
	for _, r := range rs {
		byDate[r.Date.Format(time.DateOnly)] = r
	}
	for day := end; !day.Before(start); day = day.AddDate(0, 0, -1) {
		key := day.Format(time.DateOnly)
		r := byDate[key]
		out.Daily = append(out.Daily, DailyRow{Date: key, Exported: r.Exported, Imported: r.Imported, Difference: r.Exported - r.Imported})
	}
	out.HasNewer = out.Page > 0
	out.HasOlder, err = s.Store.AnyBefore(ctx, start)
	return err
}

func (s *Service) weekly(ctx context.Context) ([]WeeklyRow, error) {
	rs, err := s.currentYear(ctx)
	if err != nil {
		return nil, err
	}
	type key struct{ year, week int }
	var rows []WeeklyRow
	var last key
	for _, r := range rs {
		y, w := r.Date.ISOWeek()
		k := key{y, w}
		if len(rows) == 0 || k != last {
			rows = append(rows, WeeklyRow{StartDate: r.Date.Format(time.DateOnly)})
			last = k
		}
		row := &rows[len(rows)-1]
		row.EndDate = r.Date.Format(time.DateOnly)
		row.Exported += r.Exported
		row.Imported += r.Imported
		row.Difference = row.Exported - row.Imported
	}
	reverse(rows)
	return rows, nil
}

func (s *Service) monthly(ctx context.Context) ([]MonthlyRow, error) {
	rs, err := s.currentYear(ctx)
	if err != nil {
		return nil, err
	}
	var rows []MonthlyRow
	for _, r := range rs {
		m := int(r.Date.Month())
		if len(rows) == 0 || rows[len(rows)-1].Month != m {
			rows = append(rows, MonthlyRow{Month: m})
		}
		row := &rows[len(rows)-1]
		row.Exported += r.Exported
		row.Imported += r.Imported
		row.Difference = row.Exported - row.Imported
	}
	reverse(rows)
	return rows, nil
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
