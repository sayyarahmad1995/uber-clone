package main

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func (app application) routes() http.Handler {
	mux := http.NewServeMux()

	app.registerSystemRoutes(mux)
	app.registerAuthRoutes(mux)
	app.registerUserRoutes(mux)
	app.registerDriverRoutes(mux)
	app.registerRideRoutes(mux)

	return mux
}

func (app application) registerSystemRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", app.ready)
}

func (app application) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/register", app.auth.Register)
	mux.HandleFunc("POST /v1/auth/login", app.auth.Login)
	mux.HandleFunc("POST /v1/auth/verify", app.auth.Verify)
	mux.HandleFunc("POST /v1/auth/verify/complete", app.auth.CompleteVerification)
	mux.HandleFunc("POST /v1/auth/session/extend", app.auth.ExtendSession)
	mux.HandleFunc("POST /v1/auth/logout", app.auth.Logout)
}

func (app application) registerUserRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v1/me", app.authenticated(app.me))
	mux.Handle("PUT /v1/me/capabilities/driver", app.authenticated(app.enableDriverCapability))
}

func (app application) registerDriverRoutes(mux *http.ServeMux) {
	mux.Handle("PUT /v1/driver", app.authenticated(app.onboardDriver))
	mux.Handle("GET /v1/driver", app.authenticated(app.getDriver))
	mux.Handle("PUT /v1/driver/availability", app.authenticated(app.setDriverAvailability))
}

func (app application) registerRideRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v1/ride-requests", app.authenticated(app.createRideRequest))
	mux.Handle("POST /v1/ride-requests/{ride_request_id}/match", app.authenticated(app.matchRideRequest))
}

func (app application) authenticated(handler http.HandlerFunc) http.Handler {
	return identity.Middleware(app.identity, handler)
}
