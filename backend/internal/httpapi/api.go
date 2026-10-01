package httpapi

import (
	"context"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driveronboarding"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/rideservice"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type Dependencies struct {
	Health                 interface{ PingContext(context.Context) error }
	Users                  user.Service
	Drivers                driver.Service
	DriverOnboarding       driveronboarding.Service
	DriverOnboardingReview driveronboarding.ReviewService
	DriverLocations        driverlocation.Service
	DriverTrips            drivertrip.Service
	RiderLocations         riderlocation.Service
	Rides                  ride.Service
	RideServices           rideservice.Service
	RideStatuses           ridestatus.Service
	Cancellations          cancellation.Service
	Offers                 offer.Service
	MarketplaceAssignments marketplace.AssignmentService
	MarketplacePolicy      marketplace.PolicyService
	Trips                           trip.Service
	LocationSearch                  locationsearch.Searcher
	LocationSearchRequestsPerMinute int
	Identity                        identity.Provider
	Auth                   auth.Handler
	AdminReviewUsername    string
	AdminReviewPassword    string
	AdminReviewOrigin      string
}

type API struct {
	health                 interface{ PingContext(context.Context) error }
	users                  user.Service
	drivers                driver.Service
	driverOnboarding       driveronboarding.Service
	driverOnboardingReview driveronboarding.ReviewService
	driverLocations        driverlocation.Service
	driverTrips            drivertrip.Service
	riderLocations         riderlocation.Service
	rides                  ride.Service
	rideServices           rideservice.Service
	rideStatuses           ridestatus.Service
	cancellations          cancellation.Service
	offers                 offer.Service
	marketplaceAssignments marketplace.AssignmentService
	marketplacePolicy      marketplace.PolicyService
	trips                 trip.Service
	locationSearch        locationsearch.Searcher
	locationSearchLimiter *userWindowLimiter
	identity              identity.Provider
	auth                   auth.Handler
	adminReviewUsername    string
	adminReviewPassword    string
	adminReviewOrigin      string
}

func New(deps Dependencies) *API {
	return &API{
		health:                 deps.Health,
		users:                  deps.Users,
		drivers:                deps.Drivers,
		driverOnboarding:       deps.DriverOnboarding,
		driverOnboardingReview: deps.DriverOnboardingReview,
		driverLocations:        deps.DriverLocations,
		driverTrips:            deps.DriverTrips,
		riderLocations:         deps.RiderLocations,
		rides:                  deps.Rides,
		rideServices:           deps.RideServices,
		rideStatuses:           deps.RideStatuses,
		cancellations:          deps.Cancellations,
		offers:                 deps.Offers,
		marketplaceAssignments: deps.MarketplaceAssignments,
		marketplacePolicy:      deps.MarketplacePolicy,
		trips:                 deps.Trips,
		locationSearch:        deps.LocationSearch,
		locationSearchLimiter: newUserWindowLimiter(deps.LocationSearchRequestsPerMinute),
		identity:              deps.Identity,
		auth:                   deps.Auth,
		adminReviewUsername:    deps.AdminReviewUsername,
		adminReviewPassword:    deps.AdminReviewPassword,
		adminReviewOrigin:      deps.AdminReviewOrigin,
	}
}

func (api *API) Handler() http.Handler {
	return api.routes()
}
