package driver

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

func TestPostgresRepositoryVehicleEnrollmentsAreExplicitAndOwnerScoped(t *testing.T) {
	db := openDriverIntegrationDB(t)
	repository := NewPostgresRepository(db)
	owner := createDriverIntegrationUser(t, db)
	other := createDriverIntegrationUser(t, db)
	first, second, foreign := uuid.New(), uuid.New(), uuid.New()
	for _, userID := range []uuid.UUID{owner, other} {
		if _, err := db.Exec(`INSERT INTO driver_profiles (user_id, status) VALUES ($1, 'approved')`, userID); err != nil {
			t.Fatal(err)
		}
	}
	for i, id := range []uuid.UUID{first, second, foreign} {
		userID := owner
		if i == 2 {
			userID = other
		}
		if _, err := db.Exec(`INSERT INTO driver_vehicles (id, driver_user_id, make, model, color, license_plate, created_at)
			VALUES ($1, $2, 'Toyota', 'Corolla', 'White', $3, NOW() + ($4 * INTERVAL '1 second'))`, id, userID, id.String(), i); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO driver_vehicle_service_enrollments (vehicle_id, service_code, approved_at, approved_by)
		VALUES ($1, 'comfort', NOW(), 'reviewer'), ($2, 'economy', NOW(), 'reviewer')`, first, foreign); err != nil {
		t.Fatal(err)
	}
	vehicles, err := repository.ListVehicles(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(vehicles) != 2 || vehicles[0].ID != first || vehicles[1].ID != second {
		t.Fatalf("unexpected owned vehicles: %#v", vehicles)
	}
	if len(vehicles[0].Enrollments) != 1 || vehicles[0].Enrollments[0].ServiceCode != "comfort" {
		t.Fatalf("Comfort must not imply Economy enrollment: %#v", vehicles[0].Enrollments)
	}
	if vehicles[0].Enrollments[0].DisplayName != "Comfort" || vehicles[0].Enrollments[0].ApprovedAt.IsZero() || !vehicles[0].Enrollments[0].ServiceActive {
		t.Fatalf("missing enrollment metadata: %#v", vehicles[0].Enrollments)
	}
	if vehicles[1].Enrollments == nil || len(vehicles[1].Enrollments) != 0 {
		t.Fatalf("vehicle without enrollment must have an empty list: %#v", vehicles[1])
	}
	if _, err := db.Exec(`INSERT INTO driver_vehicle_service_enrollments (vehicle_id, service_code, approved_at, approved_by)
		VALUES ($1, 'economy', NOW(), 'reviewer')`, first); err != nil {
		t.Fatal(err)
	}
	vehicles, err = repository.ListVehicles(context.Background(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(vehicles) != 2 || len(vehicles[0].Enrollments) != 2 || vehicles[0].Enrollments[0].ServiceCode != "economy" {
		t.Fatalf("multiple enrollments must not duplicate vehicles: %#v", vehicles)
	}
	empty, err := repository.ListVehicles(context.Background(), uuid.New())
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("unknown owner should have an empty list: %#v, %v", empty, err)
	}
}

func TestOperatingSelectionAndAvailability(t *testing.T) {
    db:=openDriverIntegrationDB(t)
    id:=createDriverIntegrationUser(t,db)
    repo:=NewPostgresRepository(db)
    vehicle:=uuid.New()
    if _,err:=db.Exec(`INSERT INTO driver_profiles(user_id,status) VALUES ($1,'approved')`,id);err!=nil { t.Fatal(err) }
    if _,err:=db.Exec(`INSERT INTO driver_vehicles(id,driver_user_id,make,model,color,license_plate) VALUES ($1,$2,'Toyota','Corolla','White','READY')`,vehicle,id);err!=nil { t.Fatal(err) }
    if _,err:=repo.SetOnline(context.Background(),id,true);!errors.Is(err,ErrSelectionInvalid) { t.Fatalf("missing selection accepted: %v",err) }
    if _,err:=repo.SelectOperation(context.Background(),id,vehicle,"economy");!errors.Is(err,ErrSelectionInvalid) { t.Fatalf("unapproved selection accepted: %v",err) }
    if _,err:=db.Exec(`INSERT INTO driver_vehicle_service_enrollments(vehicle_id,service_code,approved_at,approved_by) VALUES ($1,'economy',NOW(),'reviewer')`,vehicle);err!=nil { t.Fatal(err) }
    if _,err:=repo.SelectOperation(context.Background(),id,uuid.New(),"economy");!errors.Is(err,ErrSelectionInvalid) { t.Fatalf("foreign vehicle accepted: %v",err) }
    state,err:=repo.SelectOperation(context.Background(),id,vehicle,"economy")
    if err!=nil || state.Selection==nil || !state.Selection.Valid { t.Fatalf("selection failed: %#v %v",state,err) }
    restored,err:=repo.OperatingState(context.Background(),id)
    if err!=nil || restored.Selection==nil || restored.Selection.VehicleID!=vehicle { t.Fatalf("selection not persisted: %#v %v",restored,err) }
    if _,err:=repo.SetOnline(context.Background(),id,true);!errors.Is(err,ErrLocationRequired) { t.Fatalf("missing location accepted: %v",err) }
    if _,err:=db.Exec(`INSERT INTO driver_locations(driver_user_id,latitude,longitude,updated_at) VALUES ($1,24,67,NOW()-INTERVAL '5 minutes')`,id);err!=nil { t.Fatal(err) }
    if _,err:=repo.SetOnline(context.Background(),id,true);!errors.Is(err,ErrLocationRequired) { t.Fatalf("stale location accepted: %v",err) }
    if _,err:=db.Exec(`UPDATE driver_locations SET updated_at=NOW() WHERE driver_user_id=$1`,id);err!=nil { t.Fatal(err) }
    profile,err:=repo.SetOnline(context.Background(),id,true)
    if err!=nil || !profile.IsOnline || profile.Status!=StatusActive { t.Fatalf("approved Driver cannot go online: %#v %v",profile,err) }
    if _,err:=repo.SelectOperation(context.Background(),id,vehicle,"economy");!errors.Is(err,ErrSelectionLocked) { t.Fatalf("online selection allowed: %v",err) }
    if _,err:=repo.SetOnline(context.Background(),id,false);err!=nil { t.Fatal(err) }
    if _,err:=db.Exec(`DELETE FROM driver_vehicle_service_enrollments WHERE vehicle_id=$1`,vehicle);err!=nil { t.Fatal(err) }
    if _,err:=repo.SetOnline(context.Background(),id,true);!errors.Is(err,ErrSelectionInvalid) { t.Fatalf("revoked enrollment accepted: %v",err) }
}
