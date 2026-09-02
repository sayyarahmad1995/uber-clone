package httpapi

import (
	"database/sql"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type Dependencies struct {
	Users           user.Service
	Drivers         driver.Service
	DriverLocations driverlocation.Service
	DriverTrips     drivertrip.Service
	RiderLocations  riderlocation.Service
	Rides           ride.Service
	RideStatuses    ridestatus.Service
	Cancellations   cancellation.Service
	Matching        matching.Service
	Offers          offer.Service
	Trips           trip.Service
	DB              *sql.DB
	Identity        identity.Provider
	Auth            auth.Handler
}

type API struct {
	users           user.Service
	drivers         driver.Service
	driverLocations driverlocation.Service
	driverTrips     drivertrip.Service
	riderLocations  riderlocation.Service
	rides           ride.Service
	rideStatuses    ridestatus.Service
	cancellations   cancellation.Service
	matching        matching.Service
	offers          offer.Service
	trips           trip.Service
	db              *sql.DB
	identity        identity.Provider
	auth            auth.Handler
}

func New(deps Dependencies) *API {
	return &API{
		users:           deps.Users,
		drivers:         deps.Drivers,
		driverLocations: deps.DriverLocations,
		driverTrips:     deps.DriverTrips,
		riderLocations:  deps.RiderLocations,
		rides:           deps.Rides,
		rideStatuses:    deps.RideStatuses,
		cancellations:   deps.Cancellations,
		matching:        deps.Matching,
		offers:          deps.Offers,
		trips:           deps.Trips,
		db:              deps.DB,
		identity:        deps.Identity,
		auth:            deps.Auth,
	}
}

func (api *API) Handler() http.Handler {
	return api.routes()
}
