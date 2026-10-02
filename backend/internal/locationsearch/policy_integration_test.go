package locationsearch

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestLocationSearchPolicyPersistsAndDefaultsToFiveMeters(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatalf("TEST_DATABASE_URL must point to a database ending in _test, got %q", parsed.Path)
	}
	db, err := database.Open(databaseURL)
	if err != nil {
		t.Fatalf("open integration database: %v", err)
	}
	defer db.Close()
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`
			UPDATE location_search_policy
			SET named_place_snap_radius_meters = 5,
			    updated_at = NOW(),
			    updated_by = 'test-cleanup'
			WHERE id = 1
		`)
	})

	service := NewPolicyService(db)
	policy, err := service.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if policy.NamedPlaceSnapRadiusMeters != DefaultNamedPlaceSnapRadiusMeters {
		t.Fatalf("expected migration default %d, got %d", DefaultNamedPlaceSnapRadiusMeters, policy.NamedPlaceSnapRadiusMeters)
	}

	updated, err := service.Update(
		t.Context(),
		Policy{NamedPlaceSnapRadiusMeters: 7},
		"admin-test",
	)
	if err != nil {
		t.Fatal(err)
	}
	if updated.NamedPlaceSnapRadiusMeters != 7 || updated.UpdatedBy != "admin-test" {
		t.Fatalf("unexpected updated policy: %#v", updated)
	}

	reloaded, err := service.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.NamedPlaceSnapRadiusMeters != 7 || reloaded.UpdatedBy != "admin-test" {
		t.Fatalf("policy was not persisted: %#v", reloaded)
	}
}
