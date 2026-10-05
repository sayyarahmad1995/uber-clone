package pricing

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func openPricingTestDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatalf("integration database must end with _test: %s", parsed.Path)
	}
	db, err := database.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPostgresRepositoryPublishesImmutablePricingVersions(t *testing.T) {
	db := openPricingTestDB(t)
	code := "price_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := db.Exec(`
		INSERT INTO driver_service_catalog
			(code, display_name, description, is_active, sort_order, rider_visible, presentation_token)
		VALUES ($1, 'Pricing test', 'Synthetic pricing service', TRUE, 99, TRUE, 'car')
	`, code); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM ride_pricing_policies WHERE service_code=$1`, code)
		_, _ = db.Exec(`DELETE FROM driver_service_catalog WHERE code=$1`, code)
	})

	repository := NewPostgresRepository(db)
	first, err := repository.Publish(context.Background(), Draft{
		ServiceCode:        code,
		Currency:           "PKR",
		BaseFareMinor:      10000,
		RateMinorPerKM:     5000,
		RateMinorPerMinute: 1000,
		MinimumFareMinor:   15000,
		RoundingIncrement:  500,
	}, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != 1 || !first.Active || !first.Approved {
		t.Fatalf("unexpected first policy: %#v", first)
	}

	second, err := repository.Publish(context.Background(), Draft{
		ServiceCode:        code,
		Currency:           "PKR",
		BaseFareMinor:      12000,
		RateMinorPerKM:     5500,
		RateMinorPerMinute: 1200,
		MinimumFareMinor:   18000,
		RoundingIncrement:  500,
	}, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if second.Version != 2 {
		t.Fatalf("expected version 2, got %#v", second)
	}

	active, err := repository.ActivePolicy(
		context.Background(),
		code,
		"PKR",
		second.EffectiveFrom,
	)
	if err != nil {
		t.Fatal(err)
	}
	if active.Version != 2 || active.BaseFareMinor != 12000 {
		t.Fatalf("unexpected active policy: %#v", active)
	}

	var oldActive bool
	var oldUntil sql.NullTime
	if err := db.QueryRow(`
		SELECT is_active, effective_until
		FROM ride_pricing_policies
		WHERE service_code=$1 AND currency='PKR' AND version=1
	`, code).Scan(&oldActive, &oldUntil); err != nil {
		t.Fatal(err)
	}
	if oldActive || !oldUntil.Valid {
		t.Fatalf("historical policy was not closed: active=%v until=%v", oldActive, oldUntil)
	}

	policies, err := repository.ListActive(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, policy := range policies {
		if policy.ServiceCode == code {
			found++
			if policy.Version != 2 {
				t.Fatalf("unexpected active version: %#v", policy)
			}
		}
	}
	if found != 1 {
		t.Fatalf("expected one active policy for test service, got %d", found)
	}

	if _, err := db.Exec(`
		UPDATE driver_service_catalog
		SET rider_visible=FALSE
		WHERE code=$1
	`, code); err != nil {
		t.Fatal(err)
	}
	_, err = repository.Publish(context.Background(), Draft{
		ServiceCode:        code,
		Currency:           "PKR",
		BaseFareMinor:      13000,
		RateMinorPerKM:     6000,
		RateMinorPerMinute: 1300,
		MinimumFareMinor:   19000,
		RoundingIncrement:  500,
	}, "owner")
	if !errors.Is(err, ErrPolicyInvalid) {
		t.Fatalf("expected hidden service to reject pricing publish, got %v", err)
	}
}
