// Package httpapi serves the JSON API and the embedded frontend.
package httpapi

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/readings"
)

func init() {
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

type dashboard struct {
	Year         int                   `json:"year"`
	MaxCapacity  int                   `json:"max_capacity"`
	HasReadings  bool                  `json:"has_readings"`
	Summary      readings.Summary      `json:"summary"`
	ChartData    []readings.ChartPoint `json:"chart_data"`
	PreviousYear []readings.DailyValue `json:"previous_year"`
}

type handler struct {
	svc    *readings.Service
	pool   *pgxpool.Pool
	static fs.FS
}

func New(svc *readings.Service, pool *pgxpool.Pool, static fs.FS) http.Handler {
	h := &handler{svc: svc, pool: pool, static: static}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/dashboard", h.dashboard)
	mux.HandleFunc("GET /api/readings", h.readings)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /api/", http.NotFound)
	mux.HandleFunc("GET /{$}", h.appShell)
	mux.HandleFunc("GET /dashboard", h.appShell)
	mux.HandleFunc("GET /", h.file)
	return logRequests(securityHeaders(mux))
}

func (h *handler) dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, err := h.svc.SyncMissing(ctx); err != nil {
		slog.Error("failed to sync virtual battery readings", "error", err)
	}

	var out dashboard
	var err error
	out.Year = h.svc.Now().UTC().Year()
	out.MaxCapacity = int(readings.MaxCapacity / 100)
	if out.Summary, err = h.svc.Summary(ctx); err == nil {
		if out.ChartData, err = h.svc.ChartData(ctx); err == nil {
			if out.PreviousYear, err = h.svc.PreviousYear(ctx); err == nil {
				out.HasReadings, err = h.svc.Store.Any(ctx)
			}
		}
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, out)
}

func (h *handler) readings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	out, err := h.svc.Readings(r.Context(), r.URL.Query().Get("tab"), page)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, out)
}

func (h *handler) healthz(w http.ResponseWriter, r *http.Request) {
	if err := h.pool.Ping(r.Context()); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	_, _ = w.Write([]byte("ok"))
}

func (h *handler) appShell(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	h.serve(w, r, "index.html")
}

func (h *handler) file(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	switch {
	case name == "service-worker.js":
		w.Header().Set("Cache-Control", "no-cache")
	case strings.HasPrefix(name, "assets/"):
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	h.serve(w, r, name)
}

func (h *handler) serve(w http.ResponseWriter, r *http.Request, name string) {
	info, err := fs.Stat(h.static, name)
	if err != nil || info.IsDir() {
		w.Header().Del("Cache-Control")
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, h.static, name)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func serverError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration", time.Since(start))
	})
}
