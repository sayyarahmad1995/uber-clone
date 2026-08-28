package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type config struct {
	Port        string
	DatabaseURL string
}

type application struct {
	users user.Service
	db    *sql.DB
}

func main() {
	cfg := config{Port: getenv("APP_PORT", "8080"), DatabaseURL: os.Getenv("DATABASE_URL")}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Apply(db); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	app := application{users: user.NewService(user.NewPostgresRepository(db)), db: db}
	server := &http.Server{Addr: ":" + cfg.Port, Handler: app.routes(), ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("API server starting", "port", cfg.Port)
		errCh <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" { return value }
	return fallback
}

func (app application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", app.ready)
	// Authentication middleware is intentionally not added until the OIDC provider configuration is implemented.
	mux.HandleFunc("GET /v1/me", app.me)
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app application) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := app.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (app application) me(w http.ResponseWriter, r *http.Request) {
	subject := strings.TrimSpace(r.Header.Get("X-External-Subject"))
	if subject == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	u, err := app.users.GetOrCreate(r.Context(), subject)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load user"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": u.ID, "capabilities": u.Capabilities,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
