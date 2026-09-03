package driver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	profile Profile
	input   OnboardingInput
	online  *bool
}

func (f *fakeRepository) UpsertProfile(_ context.Context, userID uuid.UUID, input OnboardingInput) (Profile, error) {
	f.input = input
	f.profile = Profile{
		UserID:      userID,
		DisplayName: input.DisplayName,
		Status:      StatusActive,
		Vehicle: Vehicle{
			Make:         input.Vehicle.Make,
			Model:        input.Vehicle.Model,
			ModelYear:    input.Vehicle.ModelYear,
			Color:        input.Vehicle.Color,
			LicensePlate: input.Vehicle.LicensePlate,
		},
	}
	return f.profile, nil
}

func (f *fakeRepository) FindByUserID(_ context.Context, _ uuid.UUID) (Profile, error) {
	if f.profile.UserID == uuid.Nil {
		return Profile{}, ErrNotFound
	}
	return f.profile, nil
}

func (f *fakeRepository) SetOnline(_ context.Context, userID uuid.UUID, online bool) (Profile, error) {
	if f.profile.UserID == uuid.Nil {
		return Profile{}, ErrNotFound
	}
	f.online = &online
	f.profile.UserID = userID
	f.profile.IsOnline = online
	return f.profile, nil
}

func TestOnboardNormalizesPublicProfileAndVehicle(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	userID := uuid.New()

	profile, err := service.Onboard(context.Background(), userID, OnboardingInput{
		DisplayName: " Sayyar Ahmad ",
		Vehicle: VehicleInput{
			Make: " Toyota ", Model: " Corolla ", ModelYear: 2024, Color: " White ", LicensePlate: " abc-123 ",
		},
	})
	if err != nil {
		t.Fatalf("Onboard returned error: %v", err)
	}
	if profile.Status != StatusActive {
		t.Fatalf("expected active status, got %q", profile.Status)
	}
	if profile.IsOnline {
		t.Fatal("newly onboarded driver must start offline")
	}
	if repo.input.DisplayName != "Sayyar Ahmad" {
		t.Fatalf("display name was not normalized: %q", repo.input.DisplayName)
	}
	vehicle := repo.input.Vehicle
	if vehicle.Make != "Toyota" || vehicle.Model != "Corolla" || vehicle.ModelYear != 2024 || vehicle.Color != "White" || vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("vehicle was not normalized: %#v", vehicle)
	}
}

func TestOnboardRejectsIncompletePublicProfile(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Onboard(context.Background(), uuid.New(), OnboardingInput{
		DisplayName: "Driver",
		Vehicle:     VehicleInput{Make: "Toyota"},
	})
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("expected ErrInvalidProfile, got %v", err)
	}
}

func TestOnboardRejectsInvalidModelYear(t *testing.T) {
	service := NewService(&fakeRepository{})
	base := OnboardingInput{
		DisplayName: "Driver",
		Vehicle: VehicleInput{
			Make: "Toyota", Model: "Corolla", Color: "White", LicensePlate: "ABC-123",
		},
	}
	for _, year := range []int{1885, time.Now().UTC().Year() + 2} {
		input := base
		input.Vehicle.ModelYear = year
		if _, err := service.Onboard(context.Background(), uuid.New(), input); !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("expected ErrInvalidProfile for model year %d, got %v", year, err)
		}
	}
}

func TestSetOnlineRequiresExistingDriverProfile(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.SetOnline(context.Background(), uuid.New(), true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
