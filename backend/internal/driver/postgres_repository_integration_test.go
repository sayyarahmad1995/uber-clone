package driver

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestPostgresRepositoryListsMultipleVehiclesForDriver(t *testing.T) {
	db := openDriverIntegrationDB(t)
	userID := createDriverIntegrationUser(t, db)
	repository := NewPostgresRepository(db)

	if _, err := db.Exec(`
		INSERT INTO driver_profiles (user_id, display_name, status, is_online)
		VALUES ($1, 'Test Driver', 'approved', FALSE)
	`, userID); err != nil {
		t.Fatalf("insert Driver profile: %v", err)
	}
	firstVehicleID := uuid.New()
	secondVehicleID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO driver_vehicles (id, driver_user_id, make, model, model_year, color, license_plate, created_at, updated_at)
		VALUES
			($1, $3, 'Toyota', 'Corolla', 2024, 'White', 'ABC-123', NOW(), NOW()),
			($2, $3, 'Honda', 'Civic', 2023, 'Black', 'XYZ-789', NOW() + INTERVAL '1 second', NOW())
	`, firstVehicleID, secondVehicleID, userID); err != nil {
		t.Fatalf("insert Driver vehicles: %v", err)
	}

	vehicles, err := repository.ListVehicles(context.Background(), userID)
	if err != nil {
		t.Fatalf("list Driver vehicles: %v", err)
	}
	if len(vehicles) != 2 {
		t.Fatalf("expected two vehicles, got %#v", vehicles)
	}
	if vehicles[0].ID != firstVehicleID || vehicles[1].ID != secondVehicleID {
		t.Fatalf("vehicles were not returned in stable order: %#v", vehicles)
	}
}

func TestPostgresRepositoryLegacyProfileUsesStableVehicle(t *testing.T) {
	db := openDriverIntegrationDB(t)
	userID := createDriverIntegrationUser(t, db)
	repository := NewPostgresRepository(db)

	if _, err := db.Exec(`
		INSERT INTO driver_profiles (user_id, display_name, status, is_online)
		VALUES ($1, 'Test Driver', 'active', FALSE)
	`, userID); err != nil {
		t.Fatalf("insert Driver profile: %v", err)
	}
	firstVehicleID := uuid.New()
	secondVehicleID := uuid.New()
	if _, err := db.Exec(`
		INSERT INTO driver_vehicles (id, driver_user_id, make, model, model_year, color, license_plate, created_at, updated_at)
		VALUES
			($1, $3, 'Toyota', 'Corolla', 2024, 'White', 'ABC-123', NOW(), NOW()),
			($2, $3, 'Honda', 'Civic', 2023, 'Black', 'XYZ-789', NOW() + INTERVAL '1 second', NOW())
	`, firstVehicleID, secondVehicleID, userID); err != nil {
		t.Fatalf("insert Driver vehicles: %v", err)
	}

	profile, err := repository.FindByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("load legacy Driver profile: %v", err)
	}
	if profile.Vehicle.ID != firstVehicleID {
		t.Fatalf("legacy Driver profile did not use stable first vehicle: %#v", profile.Vehicle)
	}

	updated, err := repository.UpsertProfile(context.Background(), userID, OnboardingInput{
		DisplayName: "Updated Driver",
		Vehicle: VehicleInput{
			Make:         "Suzuki",
			Model:        "Swift",
			ModelYear:    2022,
			Color:        "Blue",
			LicensePlate: "NEW-456",
		},
	})
	if err != nil {
		t.Fatalf("update legacy Driver profile: %v", err)
	}
	if updated.Vehicle.ID != firstVehicleID || updated.Vehicle.LicensePlate != "NEW-456" {
		t.Fatalf("legacy update did not update stable first vehicle: %#v", updated.Vehicle)
	}

	vehicles, err := repository.ListVehicles(context.Background(), userID)
	if err != nil {
		t.Fatalf("list Driver vehicles after legacy update: %v", err)
	}
	if len(vehicles) != 2 {
		t.Fatalf("legacy update inserted or removed vehicles: %#v", vehicles)
	}
	if vehicles[1].ID != secondVehicleID || vehicles[1].LicensePlate != "XYZ-789" {
		t.Fatalf("legacy update changed non-selected vehicle: %#v", vehicles[1])
	}
}

func openDriverIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
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
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func createDriverIntegrationUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id) VALUES ($1)`, userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO user_capabilities (user_id, capability)
		VALUES ($1, 'driver')
	`, userID); err != nil {
		t.Fatalf("insert Driver capability: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}
