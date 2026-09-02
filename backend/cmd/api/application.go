package main

import (
	"fmt"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	authkratos "github.com/sayyarahmad1995/uber-clone/backend/internal/auth/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/httpapi"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	identitykratos "github.com/sayyarahmad1995/uber-clone/backend/internal/identity/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type application struct {
	handler http.Handler
}

func newApplication(cfg config) (application, func(), error) {
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return application{}, nil, fmt.Errorf("database connection failed: %w", err)
	}
	cleanup := func() { _ = db.Close() }
	if err := migrations.Apply(db); err != nil {
		cleanup()
		return application{}, nil, fmt.Errorf("database migration failed: %w", err)
	}
	authProvider, identityProvider, err := buildIdentityProviders(cfg)
	if err != nil {
		cleanup()
		return application{}, nil, fmt.Errorf("identity infrastructure initialization failed: %w", err)
	}
	tripService := trip.NewService(trip.NewPostgresRepository(db))
	api := httpapi.New(httpapi.Dependencies{
		Users:           user.NewService(user.NewPostgresRepository(db)),
		Drivers:         driver.NewService(driver.NewPostgresRepository(db)),
		DriverLocations: driverlocation.NewService(driverlocation.NewPostgresRepository(db)),
		DriverTrips:     drivertrip.NewService(drivertrip.NewPostgresRepository(db)),
		RiderLocations:  riderlocation.NewService(riderlocation.NewPostgresRepository(db)),
		Rides:           ride.NewService(ride.NewPostgresRepository(db)),
		RideStatuses:    ridestatus.NewService(ridestatus.NewPostgresRepository(db)),
		Cancellations:   cancellation.NewService(cancellation.NewPostgresRepository(db)),
		Matching:        matching.NewService(matching.NewPostgresRepository(db)),
		Offers:          offer.NewService(offer.NewPostgresRepository(db), tripService),
		Trips:           tripService,
		DB:              db,
		Identity:        identityProvider,
		Auth:            auth.NewHandler(auth.NewService(authProvider)),
	})
	return application{handler: api.Handler()}, cleanup, nil
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
