package db_test

import (
	"context"
	"testing"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/db"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/testdb"
)

// railsSchema mirrors db/schema.rb as Rails created it in production.
const railsSchema = `
CREATE TABLE schema_migrations (version varchar PRIMARY KEY);
CREATE TABLE ar_internal_metadata (key varchar PRIMARY KEY, value varchar, created_at timestamp(6) NOT NULL, updated_at timestamp(6) NOT NULL);
CREATE TABLE virtual_battery_readings (
  id bigserial PRIMARY KEY,
  created_at timestamp(6) NOT NULL,
  date date NOT NULL,
  exported_to_grid numeric(10,2) DEFAULT 0.0,
  imported_from_grid numeric(10,2) DEFAULT 0.0,
  updated_at timestamp(6) NOT NULL
);
CREATE UNIQUE INDEX index_virtual_battery_readings_on_date ON virtual_battery_readings (date);
INSERT INTO virtual_battery_readings (created_at, date, exported_to_grid, imported_from_grid, updated_at)
VALUES (now(), '2026-01-05', 12.34, 5.67, now());
`

func TestMigrateCreatesReadingsTable(t *testing.T) {
	ctx := context.Background()
	pool := testdb.Empty(t)

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO virtual_battery_readings (date) VALUES ('2026-02-01')`); err != nil {
		t.Fatalf("insert with defaults: %v", err)
	}
	var exported, imported string
	if err := pool.QueryRow(ctx, `SELECT exported_to_grid::text, imported_from_grid::text FROM virtual_battery_readings`).Scan(&exported, &imported); err != nil {
		t.Fatal(err)
	}
	if exported != "0.00" || imported != "0.00" {
		t.Fatalf("defaults = %s/%s", exported, imported)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO virtual_battery_readings (date) VALUES ('2026-02-01')`); err == nil {
		t.Fatal("expected unique violation on date")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	pool := testdb.Empty(t)
	for range 2 {
		if err := db.Migrate(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMigrateAdoptsExistingRailsSchema(t *testing.T) {
	ctx := context.Background()
	pool := testdb.Empty(t)
	if _, err := pool.Exec(ctx, railsSchema); err != nil {
		t.Fatal(err)
	}

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	var count int
	var exported string
	if err := pool.QueryRow(ctx, `SELECT count(*), max(exported_to_grid)::text FROM virtual_battery_readings`).Scan(&count, &exported); err != nil {
		t.Fatal(err)
	}
	if count != 1 || exported != "12.34" {
		t.Fatalf("existing data changed: count=%d exported=%s", count, exported)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO virtual_battery_readings (date) VALUES ('2026-02-01')`); err != nil {
		t.Fatalf("timestamps should default after adoption: %v", err)
	}
}

func TestMigrateRemovesRailsTablesWithoutChangingReadings(t *testing.T) {
	ctx := context.Background()
	pool := testdb.Empty(t)
	if _, err := pool.Exec(ctx, railsSchema+`
		CREATE TABLE solid_queue_jobs (id bigint PRIMARY KEY);
		CREATE TABLE solid_queue_ready_executions (id bigint PRIMARY KEY, job_id bigint REFERENCES solid_queue_jobs(id));
		INSERT INTO solid_queue_jobs VALUES (1);
		INSERT INTO solid_queue_ready_executions VALUES (1, 1);
	`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema_migrations", "ar_internal_metadata", "solid_queue_jobs", "solid_queue_ready_executions"} {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, name).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Errorf("Rails table %s still exists", name)
		}
	}
	var count int
	var exported, imported string
	if err := pool.QueryRow(ctx, `SELECT count(*), max(exported_to_grid)::text, max(imported_from_grid)::text FROM virtual_battery_readings`).Scan(&count, &exported, &imported); err != nil {
		t.Fatal(err)
	}
	if count != 1 || exported != "12.34" || imported != "5.67" {
		t.Fatalf("readings changed: count=%d exported=%s imported=%s", count, exported, imported)
	}
}
