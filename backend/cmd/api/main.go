package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	authkratos "github.com/sayyarahmad1995/uber-clone/backend/internal/auth/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	identitykratos "github.com/sayyarahmad1995/uber-clone/backend/internal/identity/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type config struct {
	Port             string
	DatabaseURL      string
	AuthProvider     string
	IdentitySource   string
	KratosPublicURL  string
	KratosAdminURL   string
}

type application struct {
	users    user.Service
	db       *sql.DB
	identity identity.Provider
	auth     auth.Handler
}

func main() {
	cfg := config{
		Port:            getenv("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AuthProvider:    getenv("AUTH_PROVIDER", "kratos"),
		IdentitySource:  getenv("AUTH_IDENTITY_SOURCE", "primary-identity-v1"),
		KratosPublicURL: getenv("KRATOS_PUBLIC_URL", "http://kratos:4433"),
		KratosAdminURL:  getenv("KRATOS_ADMIN_URL", "http://kratos:4434"),
	}
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

	authProvider, identityProvider, err := buildIdentityProviders(cfg)
	if err != nil {
		logger.Error("identity infrastructure initialization failed", "error", err)
		os.Exit(1)
	}

	app := application{
		users:    user.NewService(user.NewPostgresRepository(db)),
		db:       db,
		identity: identityProvider,
		auth:     auth.NewHandler(auth.NewService(authProvider)),
	}
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

func buildIdentityProviders(cfg config) (auth.Provider, identity.Provider, error) {
	switch cfg.AuthProvider {
	case "kratos":
		authProvider, err := authkratos.New(cfg.KratosPublicURL, cfg.KratosAdminURL)
		if err != nil {
			return nil, nil, err
		}
		identityProvider, err := identitykratos.New(cfg.KratosPublicURL, cfg.IdentitySource)
		if err != nil {
			return nil, nil, err
		}
		return authProvider, identityProvider, nil
	default:
		return nil, nil, fmt.Errorf("unsupported AUTH_PROVIDER %q", cfg.AuthProvider)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func (app application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", app.ready)
	mux.HandleFunc("POST /v1/auth/register", app.auth.Register)
	mux.HandleFunc("POST /v1/auth/login", app.auth.Login)
	mux.HandleFunc("POST /v1/auth/verify", app.auth.Verify)
	mux.HandleFunc("POST /v1/auth/verify/complete", app.auth.CompleteVerification)
	mux.HandleFunc("POST /v1/auth/session/extend", app.auth.ExtendSession)
	mux.HandleFunc("POST /v1/auth/logout", app.auth.Logout)
	mux.Handle("GET /v1/me", identity.Middleware(app.identity, http.HandlerFunc(app.me)))
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
	p, ok := identity.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	u, err := app.users.GetOrCreate(r.Context(), user.ExternalIdentity{Issuer: p.Issuer, Subject: p.Subject})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load user"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "capabilities": u.Capabilities})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
