package readings

import (
	"context"
	"time"
)

// MaxCapacity is the virtual battery size, 6000 kWh.
const MaxCapacity KWh = 600_000

type Summary struct {
	TotalExported KWh `json:"total_exported_to_grid"`
	TotalImported KWh `json:"total_imported_from_grid"`
	NetGridFlow   KWh `json:"net_grid_flow"`
}

type ChartPoint struct {
	Date               string `json:"date"`
	CumulativeExported KWh    `json:"cumulative_exported"`
	CumulativeImported KWh    `json:"cumulative_imported"`
}

type DailyValue struct {
	Date     string `json:"date"`
	Exported KWh    `json:"exported_to_grid"`
	Imported KWh    `json:"imported_from_grid"`
}

func (s *Service) currentYear(ctx context.Context) ([]Reading, error) {
	y := s.today().Year()
	return s.Store.Range(ctx, time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(y, 12, 31, 0, 0, 0, 0, time.UTC))
}

func (s *Service) Summary(ctx context.Context) (Summary, error) {
	rs, err := s.currentYear(ctx)
	if err != nil {
		return Summary{}, err
	}
	var sum Summary
	for _, r := range rs {
		sum.TotalExported += r.Exported
		sum.TotalImported += r.Imported
	}
	sum.NetGridFlow = min(MaxCapacity, sum.TotalExported) - sum.TotalImported
	return sum, nil
}

func (s *Service) ChartData(ctx context.Context) ([]ChartPoint, error) {
	rs, err := s.currentYear(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ChartPoint, 0, len(rs))
	var e, i KWh
	for _, r := range rs {
		e += r.Exported
		i += r.Imported
		out = append(out, ChartPoint{Date: r.Date.Format(time.DateOnly), CumulativeExported: e, CumulativeImported: i})
	}
	return out, nil
}

// oneYearBefore follows Ruby's Date - 1.year, which clamps Feb 29 to Feb 28.
func oneYearBefore(t time.Time) time.Time {
	y, m, d := t.Date()
	lastDay := time.Date(y-1, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return time.Date(y-1, m, min(d, lastDay), 0, 0, 0, 0, time.UTC)
}

// PreviousYear returns daily values from one year ago to the end of last year, used for projections.
func (s *Service) PreviousYear(ctx context.Context) ([]DailyValue, error) {
	today := s.today()
	rs, err := s.Store.Range(ctx, oneYearBefore(today), time.Date(today.Year()-1, 12, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		return nil, err
	}
	out := make([]DailyValue, 0, len(rs))
	for _, r := range rs {
		out = append(out, DailyValue{Date: r.Date.Format(time.DateOnly), Exported: r.Exported, Imported: r.Imported})
	}
	return out, nil
}
