// Package testdb gives each test an isolated, migrated Postgres schema.
package testdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func URL(t testing.TB) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	return url
}

// Empty returns a pool whose search_path is a fresh schema with no tables.
func Empty(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	buf := make([]byte, 6)
	_, _ = rand.Read(buf)
	schema := "t_" + hex.EncodeToString(buf)

	admin, err := pgxpool.New(ctx, URL(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}

	cfg, err := pgxpool.ParseConfig(URL(t))
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
	})
	return pool
}
