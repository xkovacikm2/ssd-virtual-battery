package readings

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/db"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/ssd"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/testdb"
)

var now = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

func d(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

type fakeFetcher struct {
	requested []string
	respond   func(day time.Time) ([]ssd.Interval, error)
}

func (f *fakeFetcher) FetchDay(_ context.Context, day time.Time) ([]ssd.Interval, error) {
	f.requested = append(f.requested, day.Format(time.DateOnly))
	if f.respond == nil {
		return nil, nil
	}
	return f.respond(day)
}

type fixture struct {
	svc     *Service
	store   *Store
	fetcher *fakeFetcher
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testdb.Empty(t)
	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	fetcher := &fakeFetcher{}
	svc := &Service{
		Store:      store,
		NewFetcher: func() (Fetcher, error) { return fetcher, nil },
		Now:        func() time.Time { return now },
	}
	return fixture{svc: svc, store: store, fetcher: fetcher}
}

func (f fixture) seed(t *testing.T, date string, exported, imported KWh) {
	t.Helper()
	if err := f.store.Upsert(context.Background(), d(date), exported, imported); err != nil {
		t.Fatal(err)
	}
}

func (f fixture) get(t *testing.T, date string) (Reading, bool) {
	t.Helper()
	rs, err := f.store.Range(context.Background(), d(date), d(date))
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) == 0 {
		return Reading{}, false
	}
	return rs[0], true
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var profileRows = []ssd.Interval{{Incoming: 4, Outgoing: 8}, {Incoming: 4, Outgoing: 8}}

// --- store ---

func TestUpsertCreatesThenUpdatesSameDate(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-11-01", 9900, 9900)
	f.seed(t, "2026-11-01", 200, 100)

	all, err := f.store.Range(context.Background(), d("2000-01-01"), d("2100-01-01"))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Exported != 200 || all[0].Imported != 100 {
		t.Fatalf("got %+v", all)
	}
}

func TestUpsertRejectsNegativeValues(t *testing.T) {
	f := setup(t)
	if err := f.store.Upsert(context.Background(), d("2026-11-01"), -1, 0); err == nil {
		t.Fatal("want error for negative exported")
	}
	if err := f.store.Upsert(context.Background(), d("2026-11-01"), 0, -1); err == nil {
		t.Fatal("want error for negative imported")
	}
}

// --- conversion ---

func TestDailyTotalsMapIncomingToImportedAndOutgoingToExported(t *testing.T) {
	in := func(kw float64) ssd.Interval { return ssd.Interval{Incoming: kw} }
	out := func(kw float64) ssd.Interval { return ssd.Interval{Outgoing: kw} }
	exp, imp := DailyTotals([]ssd.Interval{in(10), in(10), in(10), in(10)})
	if exp != 0 || imp != 1000 {
		t.Fatalf("incoming: exp=%d imp=%d", exp, imp)
	}
	exp, imp = DailyTotals([]ssd.Interval{out(40), out(40), out(40), out(40)})
	if exp != 4000 || imp != 0 {
		t.Fatalf("outgoing: exp=%d imp=%d", exp, imp)
	}
}

func TestDailyTotalsRoundToTwoDecimals(t *testing.T) {
	exp, imp := DailyTotals([]ssd.Interval{{Incoming: 0.123, Outgoing: 0.021}, {}})
	if exp != 1 || imp != 3 {
		t.Fatalf("exp=%d imp=%d", exp, imp)
	}
	exp, imp = DailyTotals([]ssd.Interval{{Incoming: 1.5, Outgoing: 0}, {Incoming: 0, Outgoing: 3}})
	if exp != 75 || imp != 38 {
		t.Fatalf("half-up: exp=%d imp=%d", exp, imp)
	}
}

// --- sync ---

func TestSyncFetchesEveryMissingDayUpToYesterday(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-10-30", 100, 100)
	f.fetcher.respond = func(time.Time) ([]ssd.Interval, error) { return profileRows, nil }

	saved, err := f.svc.SyncMissing(context.Background())
	if err != nil || saved != 3 {
		t.Fatalf("saved=%d err=%v", saved, err)
	}
	if want := []string{"2026-10-31", "2026-11-01", "2026-11-02"}; !equalStrings(f.fetcher.requested, want) {
		t.Fatalf("requested %v", f.fetcher.requested)
	}
	r, ok := f.get(t, "2026-11-02")
	if !ok || r.Exported != 400 || r.Imported != 200 {
		t.Fatalf("yesterday = %+v %v", r, ok)
	}
}

func TestSyncDoesNothingWhenYesterdayStored(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-11-02", 100, 100)
	saved, err := f.svc.SyncMissing(context.Background())
	if err != nil || saved != 0 || len(f.fetcher.requested) != 0 {
		t.Fatalf("saved=%d err=%v requested=%v", saved, err, f.fetcher.requested)
	}
}

func TestSyncStopsAtFirstEmptyDay(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-10-31", 100, 100)
	f.fetcher.respond = func(day time.Time) ([]ssd.Interval, error) {
		if day.Equal(d("2026-11-02")) {
			return nil, nil
		}
		return profileRows, nil
	}

	saved, err := f.svc.SyncMissing(context.Background())
	if err != nil || saved != 1 {
		t.Fatalf("saved=%d err=%v", saved, err)
	}
	if _, ok := f.get(t, "2026-11-02"); ok {
		t.Fatal("empty day must not be stored")
	}
	if want := []string{"2026-11-01", "2026-11-02"}; !equalStrings(f.fetcher.requested, want) {
		t.Fatalf("requested %v", f.fetcher.requested)
	}
}

func TestSyncStartsAtBeginningOfYearWithoutReadings(t *testing.T) {
	f := setup(t)
	if _, err := f.svc.SyncMissing(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"2026-01-01"}; !equalStrings(f.fetcher.requested, want) {
		t.Fatalf("requested %v", f.fetcher.requested)
	}
}

func TestSyncPropagatesFetchErrors(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-10-31", 100, 100)
	f.fetcher.respond = func(time.Time) ([]ssd.Interval, error) { return nil, errors.New("SSD down") }
	if _, err := f.svc.SyncMissing(context.Background()); err == nil {
		t.Fatal("want error")
	}
}

func TestSyncPropagatesClientConstructionErrors(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-10-31", 100, 100)
	f.svc.NewFetcher = func() (Fetcher, error) { return nil, errors.New("IMS_SSD_USERNAME environment variable is required") }
	if _, err := f.svc.SyncMissing(context.Background()); err == nil {
		t.Fatal("want error")
	}
}

func TestSyncSkipsWhenAnotherProcessHoldsTheLock(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-10-31", 100, 100)
	f.fetcher.respond = func(time.Time) ([]ssd.Interval, error) { return profileRows, nil }

	ctx := context.Background()
	conn, err := f.store.pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1, hashtext(current_database()))", SyncLockKey); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1, hashtext(current_database()))", SyncLockKey) //nolint:errcheck

	saved, err := f.svc.SyncMissing(ctx)
	if err != nil || saved != 0 || len(f.fetcher.requested) != 0 {
		t.Fatalf("saved=%d err=%v requested=%v", saved, err, f.fetcher.requested)
	}
}

func TestSyncLockKeyMatchesRails(t *testing.T) {
	if SyncLockKey != 74_201_001 {
		t.Fatal("lock key must match Rails so both apps never sync concurrently during cutover")
	}
}
