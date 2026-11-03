// Command ssdvb serves the SSD virtual battery dashboard.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/config"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/db"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/httpapi"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/readings"
	"github.com/xkovacikm2/ssd-virtual-battery/internal/ssd"
	"github.com/xkovacikm2/ssd-virtual-battery/web"
)

const usage = "usage: ssdvb serve|migrate|sync"

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return errors.New(usage)
	}
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	svc := &readings.Service{
		Store:      readings.NewStore(pool),
		NewFetcher: func() (readings.Fetcher, error) { return ssd.NewClient(cfg.SSD) },
		Now:        time.Now,
	}

	switch os.Args[1] {
	case "migrate":
		return db.Migrate(ctx, pool)
	case "sync":
		saved, err := svc.SyncMissing(ctx)
		fmt.Printf("Saved %d readings\n", saved)
		return err
	case "serve":
		if err := db.Migrate(ctx, pool); err != nil {
			return err
		}
		return serve(ctx, cfg.Addr, httpapi.New(svc, pool, web.Dist()))
	default:
		return errors.New(usage)
	}
}

func serve(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 10 * time.Second}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", addr)
		errc <- srv.ListenAndServe()
	}()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
