package driver

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	profile Profile
	vehicle VehicleInput
	online  *bool
}

func (f *fakeRepository) UpsertProfile(_ context.Context, userID uuid.UUID, vehicle VehicleInput) (Profile, error) {
	f.vehicle = vehicle
	f.profile = Profile{UserID: userID, Status: StatusActive, Vehicle: Vehicle{Make: vehicle.Make, Model: vehicle.Model, Color: vehicle.Color, LicensePlate: vehicle.LicensePlate}}
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

func TestOnboardNormalizesVehicleAndActivatesDriver(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	userID := uuid.New()

	profile, err := service.Onboard(context.Background(), userID, VehicleInput{
		Make: " Toyota ", Model: " Corolla ", Color: " White ", LicensePlate: " abc-123 ",
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
	if repo.vehicle.Make != "Toyota" || repo.vehicle.Model != "Corolla" || repo.vehicle.Color != "White" || repo.vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("vehicle was not normalized: %#v", repo.vehicle)
	}
}

func TestOnboardRejectsIncompleteVehicle(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.Onboard(context.Background(), uuid.New(), VehicleInput{Make: "Toyota"})
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("expected ErrInvalidProfile, got %v", err)
	}
}

func TestSetOnlineRequiresExistingDriverProfile(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.SetOnline(context.Background(), uuid.New(), true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
