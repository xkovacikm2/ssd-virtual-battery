package readings

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/ssd"
)

// SyncLockKey is shared with the Rails app's advisory lock.
const SyncLockKey = 74_201_001

type Fetcher interface {
	FetchDay(ctx context.Context, day time.Time) ([]ssd.Interval, error)
}

type Service struct {
	Store *Store
	// NewFetcher is called lazily so missing SSD credentials only fail the sync, not the page.
	NewFetcher func() (Fetcher, error)
	Now        func() time.Time
}

func (s *Service) today() time.Time {
	n := s.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) yesterday() time.Time {
	return s.today().AddDate(0, 0, -1)
}

// DailyTotals converts 15-minute kW intervals into daily kWh, rounded to hundredths.
func DailyTotals(intervals []ssd.Interval) (exported, imported KWh) {
	var out, in float64
	for _, i := range intervals {
		out += i.Outgoing / 4
		in += i.Incoming / 4
	}
	return KWh(math.Round(out * 100)), KWh(math.Round(in * 100))
}

// SyncMissing fetches every day after the latest stored reading up to yesterday.
func (s *Service) SyncMissing(ctx context.Context) (int, error) {
	yesterday := s.yesterday()
	if upToDate, err := s.Store.AnyOnOrAfter(ctx, yesterday); err != nil || upToDate {
		return 0, err
	}

	conn, err := s.Store.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	var locked bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1, hashtext(current_database()))`, SyncLockKey).Scan(&locked); err != nil {
		return 0, err
	}
	if !locked {
		return 0, nil
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1, hashtext(current_database()))`, SyncLockKey)
	}()

	start := time.Date(s.today().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	if last, ok, err := s.Store.MaxDate(ctx); err != nil {
		return 0, err
	} else if ok {
		start = last.AddDate(0, 0, 1)
	}
	if start.After(yesterday) {
		return 0, nil
	}

	fetcher, err := s.NewFetcher()
	if err != nil {
		return 0, err
	}
	saved := 0
	for day := start; !day.After(yesterday); day = day.AddDate(0, 0, 1) {
		intervals, err := fetcher.FetchDay(ctx, day)
		if err != nil {
			return saved, fmt.Errorf("fetch %s: %w", day.Format(time.DateOnly), err)
		}
		// SSD has not published this day yet; retry on a later visit.
		if len(intervals) == 0 {
			break
		}
		exported, imported := DailyTotals(intervals)
		if err := s.Store.Upsert(ctx, day, exported, imported); err != nil {
			return saved, err
		}
		slog.Info("saved reading", "date", day.Format(time.DateOnly), "exported_grid", exported, "imported_grid", imported)
		saved++
	}
	return saved, nil
}
