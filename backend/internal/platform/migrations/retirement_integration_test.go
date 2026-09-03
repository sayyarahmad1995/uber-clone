package migrations

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
)

// Each migration test owns a schema so it can exercise the historical schema
// without changing the schema used by application integration tests.
func retirementDB(t *testing.T, legacy bool) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a database ending in _test")
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	schema := "retirement_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	execRetirement(t, db, "CREATE SCHEMA "+schema)
	t.Cleanup(func() {
		if _, err := db.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("clean schema: %v", err)
		}
	})
	execRetirement(t, db, "SET search_path TO "+schema)
	if legacy {
		execRetirement(t, db, `CREATE TABLE schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`)
		entries, err := files.ReadDir("sql")
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.Name() >= "016_" {
				break
			}
			body, err := files.ReadFile("sql/" + entry.Name())
			if err != nil {
				t.Fatal(err)
			}
			execRetirement(t, db, string(body))
			execRetirement(t, db, `INSERT INTO schema_migrations (version) VALUES ($1)`, entry.Name())
		}
	}
	return db
}

func execRetirement(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("fixture/migration SQL: %v", err)
	}
}

func legacyRetirementRide(t *testing.T, db *sql.DB, rider uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execRetirement(t, db, `INSERT INTO ride_requests (id, rider_user_id, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude, status) VALUES ($1,$2,24,67,25,68,'requested')`, id, rider)
	return id
}

func assertRetired(t *testing.T, db *sql.DB) {
	t.Helper()
	var exists bool
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'ride_requests' AND column_name = 'booking_mode')`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("booking_mode still exists")
	}
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'ride_driver_candidates')`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("candidate table still exists")
	}
	if err := Apply(db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
}

func TestRetirementFreshDatabaseAndFareConstraints(t *testing.T) {
	db := retirementDB(t, false)
	if err := Apply(db); err != nil {
		t.Fatal(err)
	}
	assertRetired(t, db)
	rider := uuid.New()
	execRetirement(t, db, `INSERT INTO users (id) VALUES ($1)`, rider)
	for _, tc := range []struct {
		name             string
		amount, currency any
		valid            bool
	}{
		{"legacy_null_pair", nil, nil, true},
		{"minimum", int64(1), "PKR", true},
		{"maximum", int64(1000000000000), "PKR", true},
		{"missing_amount", nil, "PKR", false},
		{"missing_currency", int64(100), nil, false},
		{"zero", int64(0), "PKR", false},
		{"negative", int64(-1), "PKR", false},
		{"above_maximum", int64(1000000000001), "PKR", false},
		{"lowercase", int64(100), "pkr", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec(`INSERT INTO ride_requests (id, rider_user_id, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude, status, proposed_fare_minor, currency) VALUES ($1,$2,24,67,25,68,'requested',$3,$4)`, uuid.New(), rider, tc.amount, tc.currency)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, got %v", tc.valid, err)
			}
		})
	}
}

func TestRetirementPreservesLegacyRidesTripsAndOffers(t *testing.T) {
	db := retirementDB(t, true)
	rider, driver := uuid.New(), uuid.New()
	execRetirement(t, db, `INSERT INTO users (id) VALUES ($1),($2)`, rider, driver)
	execRetirement(t, db, `INSERT INTO driver_profiles (user_id, status) VALUES ($1,'active')`, driver)
	active := legacyRetirementRide(t, db, rider)
	completed := legacyRetirementRide(t, db, rider)
	pending := legacyRetirementRide(t, db, rider)
	priced := legacyRetirementRide(t, db, rider)
	execRetirement(t, db, `UPDATE ride_requests SET booking_mode='offers', proposed_fare_minor=100000, currency='PKR' WHERE id=$1`, priced)
	execRetirement(t, db, `INSERT INTO ride_offers (ride_request_id,driver_user_id,amount_minor,currency) VALUES ($1,$2,110000,'PKR')`, priced, driver)
	execRetirement(t, db, `INSERT INTO trips (ride_request_id,rider_user_id,driver_user_id,assigned_at) VALUES ($1,$2,$3,NOW())`, active, rider, driver)
	execRetirement(t, db, `INSERT INTO trips (ride_request_id,rider_user_id,driver_user_id,status,assigned_at,started_at,completed_at) VALUES ($1,$2,$3,'completed',NOW(),NOW(),NOW())`, completed, rider, driver)
	execRetirement(t, db, `INSERT INTO ride_driver_candidates (ride_request_id,driver_user_id,status,decided_at) VALUES ($1,$2,'accepted',NOW())`, active, driver)
	otherDriver := uuid.New()
	execRetirement(t, db, `INSERT INTO users (id) VALUES ($1)`, otherDriver)
	execRetirement(t, db, `INSERT INTO driver_profiles (user_id,status) VALUES ($1,'active')`, otherDriver)
	execRetirement(t, db, `INSERT INTO ride_driver_candidates (ride_request_id,driver_user_id,status) VALUES ($1,$2,'pending')`, pending, otherDriver)
	var before string
	const snapshot = `SELECT json_build_array((SELECT json_agg(t ORDER BY ride_request_id) FROM trips t),(SELECT json_agg(o ORDER BY ride_request_id) FROM ride_offers o))::text`
	if err := db.QueryRow(snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := Apply(db); err != nil {
		t.Fatal(err)
	}
	assertRetired(t, db)
	var after string
	if err := db.QueryRow(snapshot).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatal("migration changed existing Trips or offers")
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM ride_requests WHERE proposed_fare_minor IS NULL AND currency IS NULL`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected 3 preserved rides without fares, got %d", count)
	}
	if err := db.QueryRow(`SELECT count(*) FROM trips WHERE ride_request_id=$1`, pending).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("pending candidate became assigned")
	}
}

func TestRetirementRejectsUnrepresentedAcceptedCommitment(t *testing.T) {
	db := retirementDB(t, true)
	rider, driver, wrongDriver := uuid.New(), uuid.New(), uuid.New()
	execRetirement(t, db, `INSERT INTO users (id) VALUES ($1),($2),($3)`, rider, driver, wrongDriver)
	execRetirement(t, db, `INSERT INTO driver_profiles (user_id,status) VALUES ($1,'active')`, driver)
	ride := legacyRetirementRide(t, db, rider)
	execRetirement(t, db, `INSERT INTO ride_driver_candidates (ride_request_id,driver_user_id,status,decided_at) VALUES ($1,$2,'accepted',NOW())`, ride, driver)
	for _, wrongTrip := range []bool{false, true} {
		if wrongTrip {
			execRetirement(t, db, `INSERT INTO trips (ride_request_id,rider_user_id,driver_user_id,assigned_at) VALUES ($1,$2,$3,NOW())`, ride, rider, wrongDriver)
		}
		if err := Apply(db); err == nil || !strings.Contains(err.Error(), "reconcile accepted unreleased candidates") {
			t.Fatalf("expected reconciliation error, got %v", err)
		}
		var count int
		if err := db.QueryRow(`SELECT count(*) FROM ride_driver_candidates`).Scan(&count); err != nil || count != 1 {
			t.Fatalf("candidate lost on failed migration: count=%d err=%v", count, err)
		}
		if err := db.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version='016_retire_candidate_schema.sql'`).Scan(&count); err != nil || count != 0 {
			t.Fatalf("failed migration recorded: count=%d err=%v", count, err)
		}
	}
	execRetirement(t, db, `UPDATE trips SET driver_user_id=$1 WHERE ride_request_id=$2`, driver, ride)
	if err := Apply(db); err != nil {
		t.Fatalf("migration after reconciliation: %v", err)
	}
	assertRetired(t, db)
}
