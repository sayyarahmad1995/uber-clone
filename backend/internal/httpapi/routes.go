package httpapi

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func (api *API) routes() http.Handler {
	mux := http.NewServeMux()
	api.registerSystemRoutes(mux)
	api.registerAuthRoutes(mux)
	api.registerUserRoutes(mux)
	api.registerDriverRoutes(mux)
	api.registerRideRoutes(mux)
	return mux
}

func (api *API) registerSystemRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /ready", api.ready)
}

func (api *API) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/register", api.auth.Register)
	mux.HandleFunc("POST /v1/auth/login", api.auth.Login)
	mux.HandleFunc("POST /v1/auth/verify", api.auth.Verify)
	mux.HandleFunc("POST /v1/auth/verify/complete", api.auth.CompleteVerification)
	mux.HandleFunc("POST /v1/auth/session/extend", api.auth.ExtendSession)
	mux.HandleFunc("POST /v1/auth/logout", api.auth.Logout)
}

func (api *API) registerUserRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v1/me", api.authenticated(api.me))
	mux.Handle("PUT /v1/me/capabilities/driver", api.authenticated(api.enableDriverCapability))
}

func (api *API) registerDriverRoutes(mux *http.ServeMux) {
	mux.Handle("PUT /v1/driver", api.authenticated(api.onboardDriver))
	mux.Handle("GET /v1/driver", api.authenticated(api.getDriver))
	mux.Handle("PUT /v1/driver/availability", api.authenticated(api.setDriverAvailability))
	mux.Handle("PUT /v1/driver/location", api.authenticated(api.setDriverLocation))
	mux.Handle("GET /v1/driver/trip", api.authenticated(api.getDriverCurrentTrip))
	mux.Handle("GET /v1/driver/trips", api.authenticated(api.listDriverTripHistory))
	mux.Handle("GET /v1/driver/marketplace/ride-requests", api.authenticated(api.discoverDriverMarketplace))
	mux.Handle("PUT /v1/driver/ride-requests/{ride_request_id}/offer", api.authenticated(api.submitRideOffer))
	mux.Handle("POST /v1/driver/ride-requests/{ride_request_id}/accept", api.authenticated(api.acceptRideRequestCandidate))
	mux.Handle("POST /v1/driver/ride-requests/{ride_request_id}/reject", api.authenticated(api.rejectRideRequestCandidate))
	mux.Handle("POST /v1/driver/ride-requests/{ride_request_id}/start", api.authenticated(api.startTrip))
	mux.Handle("POST /v1/driver/ride-requests/{ride_request_id}/complete", api.authenticated(api.completeTrip))
	mux.Handle("POST /v1/driver/ride-requests/{ride_request_id}/cancel", api.authenticated(api.cancelDriverRideRequest))
}

func (api *API) registerRideRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v1/ride-requests", api.authenticated(api.createRideRequest))
	mux.Handle("GET /v1/ride-requests", api.authenticated(api.listRideRequests))
	mux.Handle("GET /v1/ride-requests/{ride_request_id}", api.authenticated(api.getRideRequestStatus))
	mux.Handle("GET /v1/ride-requests/active/driver-location", api.authenticated(api.getRiderActiveTripDriverLocation))
	mux.Handle("POST /v1/ride-requests/{ride_request_id}/cancel", api.authenticated(api.cancelRideRequest))
	mux.Handle("POST /v1/ride-requests/{ride_request_id}/match", api.authenticated(api.matchRideRequest))
	mux.Handle("GET /v1/ride-requests/{ride_request_id}/offers", api.authenticated(api.listRideOffers))
	mux.Handle("POST /v1/ride-requests/{ride_request_id}/offers/{driver_user_id}/accept", api.authenticated(api.acceptRideOffer))
	mux.Handle("POST /v1/ride-requests/{ride_request_id}/offers/{driver_user_id}/reject", api.authenticated(api.rejectRideOffer))
}

func (api *API) authenticated(handler http.HandlerFunc) http.Handler {
	return identity.Middleware(api.identity, handler)
}
