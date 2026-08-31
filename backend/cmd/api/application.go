package main

import (
	"database/sql"
	"fmt"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	authkratos "github.com/sayyarahmad1995/uber-clone/backend/internal/auth/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	identitykratos "github.com/sayyarahmad1995/uber-clone/backend/internal/identity/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/matching"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

type application struct {
	users    user.Service
	drivers  driver.Service
	rides    ride.Service
	matching matching.Service
	offers   offer.Service
	trips    trip.Service
	db       *sql.DB
	identity identity.Provider
	auth     auth.Handler
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

	return application{
		users:    user.NewService(user.NewPostgresRepository(db)),
		drivers:  driver.NewService(driver.NewPostgresRepository(db)),
		rides:    ride.NewService(ride.NewPostgresRepository(db)),
		matching: matching.NewService(matching.NewPostgresRepository(db)),
		offers:   offer.NewService(offer.NewPostgresRepository(db)),
		trips:    trip.NewService(trip.NewPostgresRepository(db)),
		db:       db,
		identity: identityProvider,
		auth:     auth.NewHandler(auth.NewService(authProvider)),
	}, cleanup, nil
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
