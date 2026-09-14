package driveronboarding

import (
	"context"
	"testing"
)

func TestPostgresRepositoryIdentifiesInitialOnboardingApplication(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	repository := NewPostgresRepository(db)
	service := NewService(repository)
	review := NewReviewService(repository)

	application, err := service.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("comfort", "INITIAL-123"),
	)
	if err != nil {
		t.Fatalf("submit initial application: %v", err)
	}
	if application.ApplicationType != ApplicationTypeInitialOnboarding {
		t.Fatalf("expected initial application type on submit, got %q", application.ApplicationType)
	}

	pending, err := review.ListPending(context.Background())
	if err != nil {
		t.Fatalf("list pending applications: %v", err)
	}
	if len(pending) != 1 || pending[0].ApplicationType != ApplicationTypeInitialOnboarding {
		t.Fatalf("expected one pending initial application, got %#v", pending)
	}

	approved, err := review.Approve(context.Background(), application.ID, "reviewer")
	if err != nil {
		t.Fatalf("approve initial application: %v", err)
	}
	if approved.ApplicationType != ApplicationTypeInitialOnboarding {
		t.Fatalf("expected approved initial application to remain initial, got %q", approved.ApplicationType)
	}
}

func TestPostgresRepositoryIdentifiesAdditionalVehicleServiceApplication(t *testing.T) {
	db := openDriverOnboardingIntegrationDB(t)
	userID := createDriverOnboardingUser(t, db)
	repository := NewPostgresRepository(db)
	service := NewService(repository)
	review := NewReviewService(repository)

	initial, err := service.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("comfort", "FIRST-123"),
	)
	if err != nil {
		t.Fatalf("submit initial application: %v", err)
	}
	if _, err := review.Approve(context.Background(), initial.ID, "reviewer"); err != nil {
		t.Fatalf("approve initial application: %v", err)
	}

	additional, err := service.Submit(
		context.Background(),
		userID,
		validDriverOnboardingInput("economy", "SECOND-456"),
	)
	if err != nil {
		t.Fatalf("submit additional application: %v", err)
	}
	if additional.ApplicationType != ApplicationTypeAdditionalVehicleService {
		t.Fatalf("expected additional application type on submit, got %q", additional.ApplicationType)
	}

	latest, err := service.Latest(context.Background(), userID)
	if err != nil {
		t.Fatalf("restore latest application: %v", err)
	}
	if latest.ID != additional.ID || latest.ApplicationType != ApplicationTypeAdditionalVehicleService {
		t.Fatalf("expected latest additional application, got %#v", latest)
	}

	pending, err := review.ListPending(context.Background())
	if err != nil {
		t.Fatalf("list pending applications: %v", err)
	}
	if len(pending) != 1 || pending[0].ID != additional.ID || pending[0].ApplicationType != ApplicationTypeAdditionalVehicleService {
		t.Fatalf("expected one pending additional application, got %#v", pending)
	}
}
