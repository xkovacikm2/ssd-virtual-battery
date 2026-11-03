package readings

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestKWhMarshalsAsTwoDecimalNumber(t *testing.T) {
	cases := map[KWh]string{0: "0.00", 5: "0.05", 1234: "12.34", -350: "-3.50", -5: "-0.05", 600000: "6000.00"}
	for in, want := range cases {
		got, err := json.Marshal(in)
		if err != nil || string(got) != want {
			t.Errorf("%d => %s (%v), want %s", in, got, err, want)
		}
	}
}

func TestSummaryTotalsCurrentYear(t *testing.T) {
	f := setup(t)
	f.seed(t, "2025-12-31", 50000, 50000)
	f.seed(t, "2026-11-01", 1000, 500)
	f.seed(t, "2026-11-02", 1000, 500)
	f.seed(t, "2026-11-03", 1000, 500)

	s, err := f.svc.Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalExported != 3000 || s.TotalImported != 1500 || s.NetGridFlow != 1500 {
		t.Fatalf("%+v", s)
	}
}

func TestSummaryCapsExportAtCapacity(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-11-01", 350000, 40000)
	f.seed(t, "2026-11-02", 320000, 70000)

	s, err := f.svc.Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalExported != 670000 || s.TotalImported != 110000 || s.NetGridFlow != 490000 {
		t.Fatalf("%+v", s)
	}
}

func TestSummaryZeroWithoutReadings(t *testing.T) {
	f := setup(t)
	s, err := f.svc.Summary(context.Background())
	if err != nil || s != (Summary{}) {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestChartDataIsCumulativeForCurrentYear(t *testing.T) {
	f := setup(t)
	f.seed(t, "2025-12-31", 99900, 99900)
	f.seed(t, "2026-11-01", 1000, 400)
	f.seed(t, "2026-11-02", 2000, 600)
	f.seed(t, "2026-11-03", 3000, 800)

	got, err := f.svc.ChartData(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []ChartPoint{
		{"2026-11-01", 1000, 400},
		{"2026-11-02", 3000, 1000},
		{"2026-11-03", 6000, 1800},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	}
}

func TestChartDataEmpty(t *testing.T) {
	f := setup(t)
	got, err := f.svc.ChartData(context.Background())
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v %v", got, err)
	}
}

func TestPreviousYearFromOneYearAgoToEndOfLastYear(t *testing.T) {
	f := setup(t)
	f.seed(t, "2025-11-02", 500, 200) // one day before the cutoff
	f.seed(t, "2025-11-03", 1500, 700)
	f.seed(t, "2025-12-31", 2000, 800)
	f.seed(t, "2026-01-01", 9900, 9900)

	got, err := f.svc.PreviousYear(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []DailyValue{{"2025-11-03", 1500, 700}, {"2025-12-31", 2000, 800}}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v", got)
	}
}

func TestPreviousYearCutoffOnLeapDay(t *testing.T) {
	f := setup(t)
	f.svc.Now = func() time.Time { return d("2028-02-29").Add(time.Hour) }
	f.seed(t, "2027-02-27", 100, 100)
	f.seed(t, "2027-02-28", 200, 200)

	got, err := f.svc.PreviousYear(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Date != "2027-02-28" {
		t.Fatalf("Ruby's Date - 1.year clamps Feb 29 to Feb 28, got %+v", got)
	}
}

// --- readings tabs (ported from DashboardControllerTest) ---

func TestDailyShowsSevenDaysEndingYesterdayWithZeroFill(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-11-02", 2550, 875)
	f.seed(t, "2026-10-30", 100, 300)

	page, err := f.svc.Readings(context.Background(), "daily", 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.Tab != "daily" || page.Page != 0 || page.HasNewer || page.HasOlder {
		t.Fatalf("meta %+v", page)
	}
	rows := page.Daily
	if len(rows) != 7 || rows[0].Date != "2026-11-02" || rows[6].Date != "2026-10-27" {
		t.Fatalf("rows %+v", rows)
	}
	if rows[0].Exported != 2550 || rows[0].Imported != 875 || rows[0].Difference != 1675 {
		t.Fatalf("first row %+v", rows[0])
	}
	if rows[3] != (DailyRow{Date: "2026-10-30", Exported: 100, Imported: 300, Difference: -200}) {
		t.Fatalf("negative diff row %+v", rows[3])
	}
	if rows[1] != (DailyRow{Date: "2026-11-01"}) {
		t.Fatalf("gap row %+v", rows[1])
	}
}

func TestDailyPaginationFlags(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-11-03", 500, 200)
	f.seed(t, "2026-10-19", 500, 200)

	page, _ := f.svc.Readings(context.Background(), "daily", 1)
	if !page.HasNewer || !page.HasOlder {
		t.Fatalf("page 1 %+v", page)
	}
	if page.Daily[0].Date != "2026-10-26" || page.Daily[6].Date != "2026-10-20" {
		t.Fatalf("window %s..%s", page.Daily[0].Date, page.Daily[6].Date)
	}
	page, _ = f.svc.Readings(context.Background(), "daily", 2)
	if page.HasOlder {
		t.Fatal("no data before page 2 window")
	}
}

func TestReadingsClampsInvalidInput(t *testing.T) {
	f := setup(t)
	page, err := f.svc.Readings(context.Background(), "bogus", -3)
	if err != nil || page.Tab != "daily" || page.Page != 0 || page.HasNewer {
		t.Fatalf("%+v %v", page, err)
	}
}

func TestWeeklyAggregatesByISOWeekNewestFirst(t *testing.T) {
	f := setup(t)
	f.seed(t, "2025-12-15", 7700, 100)
	f.seed(t, "2026-10-19", 1100, 100) // Monday, week 43
	f.seed(t, "2026-10-26", 1000, 300) // Monday, week 44
	f.seed(t, "2026-10-29", 1500, 500) // Thursday, week 44
	f.seed(t, "2026-11-02", 2200, 100) // Monday, week 45

	page, err := f.svc.Readings(context.Background(), "weekly", 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []WeeklyRow{
		{StartDate: "2026-11-02", EndDate: "2026-11-02", Exported: 2200, Imported: 100, Difference: 2100},
		{StartDate: "2026-10-26", EndDate: "2026-10-29", Exported: 2500, Imported: 800, Difference: 1700},
		{StartDate: "2026-10-19", EndDate: "2026-10-19", Exported: 1100, Imported: 100, Difference: 1000},
	}
	if len(page.Weekly) != len(want) {
		t.Fatalf("got %+v", page.Weekly)
	}
	for i := range want {
		if page.Weekly[i] != want[i] {
			t.Fatalf("row %d = %+v, want %+v", i, page.Weekly[i], want[i])
		}
	}
	if page.HasNewer || page.HasOlder {
		t.Fatal("weekly has no pagination")
	}
}

// Rails grouped by cweek alone, which merged late-December days of ISO week 1 into January's week 1.
func TestWeeklyKeepsYearBoundaryWeeksApart(t *testing.T) {
	f := setup(t)
	f.svc.Now = func() time.Time { return d("2025-12-31").Add(time.Hour) }
	f.seed(t, "2025-01-02", 100, 0) // ISO 2025-W01
	f.seed(t, "2025-12-30", 200, 0) // ISO 2026-W01

	page, _ := f.svc.Readings(context.Background(), "weekly", 0)
	if len(page.Weekly) != 2 || page.Weekly[0].StartDate != "2025-12-30" {
		t.Fatalf("got %+v", page.Weekly)
	}
}

func TestMonthlyAggregatesByMonthNewestFirst(t *testing.T) {
	f := setup(t)
	f.seed(t, "2026-01-15", 1000, 400)
	f.seed(t, "2026-02-05", 1000, 300)
	f.seed(t, "2026-02-20", 1200, 400)
	f.seed(t, "2026-11-02", 500, 900)

	page, err := f.svc.Readings(context.Background(), "monthly", 0)
	if err != nil {
		t.Fatal(err)
	}
	want := []MonthlyRow{
		{Month: 11, Exported: 500, Imported: 900, Difference: -400},
		{Month: 2, Exported: 2200, Imported: 700, Difference: 1500},
		{Month: 1, Exported: 1000, Imported: 400, Difference: 600},
	}
	if len(page.Monthly) != len(want) {
		t.Fatalf("got %+v", page.Monthly)
	}
	for i := range want {
		if page.Monthly[i] != want[i] {
			t.Fatalf("row %d = %+v, want %+v", i, page.Monthly[i], want[i])
		}
	}
}

func TestHasReadings(t *testing.T) {
	f := setup(t)
	if ok, _ := f.store.Any(context.Background()); ok {
		t.Fatal("empty")
	}
	f.seed(t, "2020-01-01", 1, 1)
	if ok, _ := f.store.Any(context.Background()); !ok {
		t.Fatal("any reading counts, even from earlier years")
	}
}
