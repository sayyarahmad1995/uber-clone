package driveronboarding

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
	if !restored.SubmittedAt.Equal(created.SubmittedAt) {
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
		SET status = 'rejected', rejection_reason = 'vehicle condition did not meet service requirements', decided_at = NOW(), updated_at = NOW()
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
