package driveronboarding

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driver"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

func TestPostgresRepositorySubmitAndRestoreApplication(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	service := NewService(NewPostgresRepository(db))

	input := ApplicationInput{
		DisplayName: " Test Driver ",
		ServiceCode: "comfort",
		Vehicle: VehicleInput{
			Make:         " Toyota ",
			Model:        " Corolla ",
			ModelYear:    2024,
			Color:        " White ",
			LicensePlate: " abc-123 ",
		},
	}

	created, err := service.Submit(context.Background(), userID, input)
	if err != nil {
		t.Fatalf("submit onboarding application: %v", err)
	}
	if created.ID == uuid.Nil || created.Status != StatusPending {
		t.Fatalf("unexpected created application: %#v", created)
	}
	if created.DisplayName != "Test Driver" || created.Service.Code != "comfort" {
		t.Fatalf("unexpected normalized application: %#v", created)
	}
	if created.Vehicle.Make != "Toyota" || created.Vehicle.Model != "Corolla" || created.Vehicle.Color != "White" || created.Vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("unexpected normalized vehicle: %#v", created.Vehicle)
	}

	restored, err := service.Latest(context.Background(), userID)
	if err != nil {
		t.Fatalf("restore onboarding application: %v", err)
	}
	if restored.ID != created.ID || restored.DriverUserID != userID || restored.Status != StatusPending {
		t.Fatalf("restored application does not match submission: created=%#v restored=%#v", created, restored)
	}
	if restored.DisplayName != created.DisplayName || restored.Service.Code != created.Service.Code || restored.Vehicle != created.Vehicle {
		t.Fatalf("restored application changed submitted snapshot: created=%#v restored=%#v", created, restored)
	}
	if delta := restored.SubmittedAt.Sub(created.SubmittedAt); delta < -time.Millisecond || delta > time.Millisecond {
		t.Fatalf("submitted_at changed during restore: created=%v restored=%v", created.SubmittedAt, restored.SubmittedAt)
	}
}

func TestPostgresRepositoryRejectsSecondPendingApplication(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	service := NewService(NewPostgresRepository(db))

	first := validDriverOnboardingInput("economy", "ONE-123")
	if _, err := service.Submit(context.Background(), userID, first); err != nil {
		t.Fatalf("submit first application: %v", err)
	}

	second := validDriverOnboardingInput("comfort", "TWO-456")
	_, err := service.Submit(context.Background(), userID, second)
	if !errors.Is(err, ErrPendingApplication) {
		t.Fatalf("expected ErrPendingApplication, got %v", err)
	}

	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM driver_onboarding_applications
		WHERE driver_user_id = $1 AND status = 'pending'
	`, userID).Scan(&count); err != nil {
		t.Fatalf("count pending applications: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one pending application, got %d", count)
	}
}

func TestPostgresRepositoryAllowsNewSubmissionAfterPreviousDecision(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	service := NewService(NewPostgresRepository(db))

	first, err := service.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("comfort", "OLD-123"),
	)
	if err != nil {
		t.Fatalf("submit first application: %v", err)
	}
	if _, err := db.Exec(`
		UPDATE driver_onboarding_applications
		SET status = 'rejected', rejection_reason = 'vehicle condition did not meet service requirements', decided_at = NOW(), decided_by = 'test-reviewer', updated_at = NOW()
		WHERE id = $1
	`, first.ID); err != nil {
		t.Fatalf("mark first application rejected: %v", err)
	}

	second, err := service.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("economy", "NEW-456"),
	)
	if err != nil {
		t.Fatalf("submit second application after decision: %v", err)
	}
	if second.ID == first.ID || second.Status != StatusPending {
		t.Fatalf("unexpected second application: first=%#v second=%#v", first, second)
	}

	latest, err := service.Latest(context.Background(), userID)
	if err != nil {
		t.Fatalf("restore latest application: %v", err)
	}
	if latest.ID != second.ID {
		t.Fatalf("expected latest application %s, got %s", second.ID, latest.ID)
	}
}

func TestPostgresReviewApprovePromotesExactSnapshotAndServiceEnrollment(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	repository := NewPostgresRepository(db)
	submission := NewService(repository)
	review := NewReviewService(repository)

	application, err := submission.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("comfort", "APP-123"),
	)
	if err != nil {
		t.Fatalf("submit application: %v", err)
	}

	approved, err := review.Approve(context.Background(), application.ID, "reviewer")
	if err != nil {
		t.Fatalf("approve application: %v", err)
	}
	if approved.Status != StatusApproved || approved.DecidedAt == nil || approved.DecidedBy != "reviewer" {
		t.Fatalf("unexpected approved application: %#v", approved)
	}

	var profileStatus string
	var isOnline bool
	var displayName string
	if err := db.QueryRow(`
		SELECT status, is_online, display_name
		FROM driver_profiles
		WHERE user_id = $1
	`, userID).Scan(&profileStatus, &isOnline, &displayName); err != nil {
		t.Fatalf("load promoted Driver profile: %v", err)
	}
	if profileStatus != "approved" || isOnline || displayName != application.DisplayName {
		t.Fatalf("unexpected promoted Driver profile: status=%s online=%v display_name=%q", profileStatus, isOnline, displayName)
	}
	if _, err := driver.NewPostgresRepository(db).SetOnline(context.Background(), userID, true); !errors.Is(err, driver.ErrSelectionInvalid) {
		t.Fatalf("approved Driver bypassed active-only availability guard: %v", err)
	}

	var vehicleID uuid.UUID
	var make, model, color, plate string
	var modelYear int
	if err := db.QueryRow(`
		SELECT id, make, model, model_year, color, license_plate
		FROM driver_vehicles
		WHERE driver_user_id = $1
	`, userID).Scan(&vehicleID, &make, &model, &modelYear, &color, &plate); err != nil {
		t.Fatalf("load promoted vehicle: %v", err)
	}
	if make != application.Vehicle.Make || model != application.Vehicle.Model || modelYear != application.Vehicle.ModelYear || color != application.Vehicle.Color || plate != application.Vehicle.LicensePlate {
		t.Fatalf("promoted vehicle changed submitted snapshot: application=%#v persisted=%s %s %d %s %s", application.Vehicle, make, model, modelYear, color, plate)
	}

	var serviceCode, approvedBy string
	if err := db.QueryRow(`
		SELECT service_code, approved_by
		FROM driver_vehicle_service_enrollments
		WHERE vehicle_id = $1
	`, vehicleID).Scan(&serviceCode, &approvedBy); err != nil {
		t.Fatalf("load service enrollment: %v", err)
	}
	if serviceCode != "comfort" || approvedBy != "reviewer" {
		t.Fatalf("unexpected service enrollment: service=%s reviewer=%s", serviceCode, approvedBy)
	}
	var economyEnrollmentCount int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM driver_vehicle_service_enrollments
		WHERE vehicle_id = $1 AND service_code = 'economy'
	`, vehicleID).Scan(&economyEnrollmentCount); err != nil {
		t.Fatalf("count implied Economy enrollment: %v", err)
	}
	if economyEnrollmentCount != 0 {
		t.Fatalf("Comfort approval silently enrolled Economy: count=%d", economyEnrollmentCount)
	}

	if _, err := review.Reject(context.Background(), application.ID, "reviewer", "second decision"); !errors.Is(err, ErrApplicationNotPending) {
		t.Fatalf("expected immutable decision, got %v", err)
	}
}

func TestPostgresReviewRejectKeepsOperationalRecordsAbsent(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	repository := NewPostgresRepository(db)
	submission := NewService(repository)
	review := NewReviewService(repository)

	application, err := submission.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("economy", "REJ-123"),
	)
	if err != nil {
		t.Fatalf("submit application: %v", err)
	}
	const reason = "vehicle condition does not meet Economy requirements"
	rejected, err := review.Reject(context.Background(), application.ID, "reviewer", reason)
	if err != nil {
		t.Fatalf("reject application: %v", err)
	}
	if rejected.Status != StatusRejected || rejected.RejectionReason != reason || rejected.DecidedAt == nil || rejected.DecidedBy != "reviewer" {
		t.Fatalf("unexpected rejected application: %#v", rejected)
	}

	var profileCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM driver_profiles WHERE user_id = $1`, userID).Scan(&profileCount); err != nil {
		t.Fatalf("count Driver profiles: %v", err)
	}
	if profileCount != 0 {
		t.Fatalf("rejection created operational Driver records: count=%d", profileCount)
	}

	restored, err := submission.Latest(context.Background(), userID)
	if err != nil {
		t.Fatalf("restore rejected application: %v", err)
	}
	if restored.Status != StatusRejected || restored.RejectionReason != reason || restored.DecidedBy != "reviewer" {
		t.Fatalf("rejected decision did not restore: %#v", restored)
	}
}

func TestPostgresSubmissionCannotRaceApprovalIntoNewPendingApplication(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	repository := NewPostgresRepository(db)
	submission := NewService(repository)
	review := NewReviewService(repository)

	application, err := submission.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("comfort", "RACE-123"),
	)
	if err != nil {
		t.Fatalf("submit initial application: %v", err)
	}

	if _, err := db.Exec(`
		CREATE OR REPLACE FUNCTION driver_onboarding_test_delay_race_insert()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			IF NEW.vehicle_license_plate = 'RACE-456' THEN
				PERFORM pg_sleep(2);
			END IF;
			RETURN NEW;
		END;
		$$
	`); err != nil {
		t.Fatalf("create race-test trigger function: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TRIGGER driver_onboarding_test_delay_race_insert
		BEFORE INSERT ON driver_onboarding_applications
		FOR EACH ROW
		EXECUTE FUNCTION driver_onboarding_test_delay_race_insert()
	`); err != nil {
		t.Fatalf("create race-test trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP TRIGGER IF EXISTS driver_onboarding_test_delay_race_insert ON driver_onboarding_applications`)
		_, _ = db.Exec(`DROP FUNCTION IF EXISTS driver_onboarding_test_delay_race_insert()`)
	})

	submitDone := make(chan error, 1)
	go func() {
		_, err := submission.Submit(
			context.Background(),
			userID,
			validDriverOnboardingInput("economy", "RACE-456"),
		)
		submitDone <- err
	}()

	deadline := time.Now().Add(time.Second)
	for {
		var sleeping bool
		if err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM pg_stat_activity
				WHERE datname = current_database()
				  AND query LIKE '%INSERT INTO driver_onboarding_applications%'
				  AND wait_event = 'PgSleep'
			)
		`).Scan(&sleeping); err != nil {
			t.Fatalf("inspect race-test submission state: %v", err)
		}
		if sleeping {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for the delayed submission insert")
		}
		time.Sleep(10 * time.Millisecond)
	}

	approveDone := make(chan error, 1)
	go func() {
		_, err := review.Approve(context.Background(), application.ID, "reviewer")
		approveDone <- err
	}()

	if err := <-submitDone; !errors.Is(err, ErrPendingApplication) {
		t.Fatalf("expected coordinated submission to remain blocked by the pending application, got %v", err)
	}
	if err := <-approveDone; err != nil {
		t.Fatalf("approve application after coordinated submission: %v", err)
	}

	var pendingCount int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM driver_onboarding_applications
		WHERE driver_user_id = $1 AND status = 'pending'
	`, userID).Scan(&pendingCount); err != nil {
		t.Fatalf("count pending applications after race: %v", err)
	}
	if pendingCount != 0 {
		t.Fatalf("approval race left a new pending application: count=%d", pendingCount)
	}
}

func validDriverOnboardingInput(serviceCode, licensePlate string) ApplicationInput {
	return ApplicationInput{
		DisplayName: "Test Driver",
		ServiceCode: serviceCode,
		Vehicle: VehicleInput{
			Make:         "Toyota",
			Model:        "Corolla",
			ModelYear:    2024,
			Color:        "White",
			LicensePlate: licensePlate,
		},
	}
}

func openDriverOnboardingIntegrationDB(t *testing.T) *sql.DB {
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

func createDriverOnboardingUser(t *testing.T, db *sql.DB) uuid.UUID {
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
