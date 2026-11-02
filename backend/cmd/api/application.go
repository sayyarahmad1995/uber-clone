package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
 "github.com/redis/go-redis/v9"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
	authkratos "github.com/sayyarahmad1995/uber-clone/backend/internal/auth/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/cancellation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driveronboarding"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/httpapi"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	identitykratos "github.com/sayyarahmad1995/uber-clone/backend/internal/identity/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
	locationsearchgoogle "github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch/google"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/riderlocation"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/rideservice"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
	routinggoogle "github.com/sayyarahmad1995/uber-clone/backend/internal/routing/google"
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
	// Only the verified PostgreSQL owner of the Kratos identity may use
	// the phone-only protected-route authorization path.
	phoneBindings, err := auth.NewPostgresPhoneAccountResolver(db, cfg.IdentitySource)
	if err != nil {
		cleanup()
		return application{}, nil, fmt.Errorf("phone identity resolver initialization failed: %w", err)
	}
	if kratosIdentity, ok := identityProvider.(*identitykratos.Provider); ok {
		kratosIdentity.WithVerifiedPhoneBindings(phoneBindings)
	}
	if kratosAuth, ok := authProvider.(*authkratos.Provider); ok {
		kratosAuth.WithVerifiedPhoneBindings(cfg.IdentitySource, phoneBindings)
	}
	var phoneRegistration *auth.PhoneRegistrationCoordinator
	var closePhoneRedis func()
	if cfg.PhoneRegistrationEnabled {
		// Explicit opt-in only. No OTP issuance or SMS route is enabled yet.
		if cfg.AuthProvider != "kratos" {
			cleanup()
			return application{}, nil, fmt.Errorf("phone registration requires Kratos")
		}
		options, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			cleanup()
			return application{}, nil, fmt.Errorf("invalid phone registration Redis configuration: %w", err)
		}
		redisClient := redis.NewClient(options)
		closePhoneRedis = func() { _ = redisClient.Close() }
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			closePhoneRedis()
			cleanup()
			return application{}, nil, fmt.Errorf("phone registration Redis unavailable: %w", err)
		}
		proofs, err := auth.NewRedisPhoneProofs(redisClient)
		if err != nil {
			closePhoneRedis()
			cleanup()
			return application{}, nil, err
		}
		kratosProvider := authProvider.(*authkratos.Provider)
		phoneRegistration = &auth.PhoneRegistrationCoordinator{
			Proofs: proofs, Identities: kratosProvider, Operations: phoneBindings,
		}
	}
	tripService := trip.NewService(trip.NewPostgresRepository(db))
	marketplaceAssignmentService := marketplace.NewAssignmentService(
		marketplace.NewPostgresAssignmentRepository(db),
	)
	marketplacePolicyService := marketplace.NewPolicyService(db)
	driverOnboardingRepository := driveronboarding.NewPostgresRepository(db)
	locationSearchPolicyService := locationsearch.NewSearchPolicyService(db)
	locationSearchService := locationsearch.NewServiceWithPolicy(
		locationsearchgoogle.New(cfg.GoogleMapsServerAPIKey),
		locationSearchPolicyService,
	)
	routingService := routing.NewService(routinggoogle.New(cfg.GoogleMapsServerAPIKey))
	pricingRepository := pricing.NewPostgresRepository(db)
	api := httpapi.New(httpapi.Dependencies{
		Pricing:                         pricingRepository,
		SuggestedFaresEnabled:           cfg.SuggestedFaresEnabled,
		Health:                          db,
		Users:                           user.NewService(user.NewPostgresRepository(db)),
		Drivers:                         driver.NewService(driver.NewPostgresRepository(db)),
		DriverOnboarding:                driveronboarding.NewService(driverOnboardingRepository),
		DriverOnboardingReview:          driveronboarding.NewReviewService(driverOnboardingRepository),
		DriverLocations:                 driverlocation.NewService(driverlocation.NewPostgresRepository(db)),
		DriverTrips:                     drivertrip.NewService(drivertrip.NewPostgresRepository(db)),
		RiderLocations:                  riderlocation.NewService(riderlocation.NewPostgresRepository(db)),
		Rides:                           ride.NewService(ride.NewPostgresRepositoryWithPricing(db, pricingRepository, cfg.SuggestedFaresEnabled)),
		RideServices:                    rideservice.NewService(rideservice.NewPostgresRepositoryWithPricing(db, cfg.SuggestedFaresEnabled)),
		RideStatuses:                    ridestatus.NewService(ridestatus.NewPostgresRepository(db)),
		Cancellations:                   cancellation.NewService(cancellation.NewPostgresRepository(db)),
		Offers:                          offer.NewService(offer.NewPostgresRepository(db)),
		MarketplaceAssignments:          marketplaceAssignmentService,
		MarketplacePolicy:               marketplacePolicyService,
		Trips:                           tripService,
		LocationSearch:                  locationSearchService,
		LocationSearchPolicy:            locationSearchPolicyService,
		Routing:                         routingService,
		LocationSearchRequestsPerMinute: cfg.LocationSearchRequestsPerMinute,
		Identity:                        identityProvider,
		Auth:                            auth.NewHandler(auth.NewService(authProvider)),
 PhoneRegistrationEnabled: cfg.PhoneRegistrationEnabled,
 PhoneRegistration: phoneRegistration,
		AdminReviewUsername:             cfg.AdminReviewUsername,
		AdminReviewPassword:             cfg.AdminReviewPassword,
		AdminReviewOrigin:               cfg.AdminReviewOrigin,
	})

	presenceCtx, cancelPresence := context.WithCancel(context.Background())
	presenceDone := make(chan struct{})
	go func() {
		defer close(presenceDone)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			ctx, cancel := context.WithTimeout(presenceCtx, 5*time.Second)
			err := driver.NewPostgresRepository(db).ExpirePresence(ctx)
			cancel()
			if err != nil && presenceCtx.Err() == nil {
				slog.Error("expire Driver presence", "error", err)
			}
			select {
			case <-presenceCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	expiryService := marketplace.NewExpiryService(db)
	expiryCtx, cancelExpiry := context.WithCancel(context.Background())
	expiryDone := make(chan struct{})
	go func() {
		defer close(expiryDone)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			ctx, cancel := context.WithTimeout(expiryCtx, 5*time.Second)
			err := expiryService.Sweep(ctx)
			cancel()
			if err != nil && expiryCtx.Err() == nil {
				slog.Error("sweep marketplace expiry", "error", err)
			}
			select {
			case <-expiryCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	return application{handler: api.Handler()}, func() {
		cancelExpiry()
		cancelPresence()
		<-expiryDone
		<-presenceDone
		if closePhoneRedis != nil { closePhoneRedis() }
		cleanup()
	}, nil
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
