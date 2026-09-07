package driveronboarding

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	service     ServiceOption
	application Application
	created     ApplicationInput
}

func (r *fakeRepository) ListServices(context.Context) ([]ServiceOption, error) {
	return []ServiceOption{r.service}, nil
}
func (r *fakeRepository) FindService(context.Context, string) (ServiceOption, error) {
	if r.service.Code == "" {
		return ServiceOption{}, ErrServiceNotFound
	}
	return r.service, nil
}
func (r *fakeRepository) LatestApplication(context.Context, uuid.UUID) (Application, error) {
	if r.application.ID == uuid.Nil {
		return Application{}, ErrNotFound
	}
	return r.application, nil
}
func (r *fakeRepository) CreateApplication(_ context.Context, userID uuid.UUID, input ApplicationInput, service ServiceOption) (Application, error) {
	r.created = input
	return Application{DriverUserID: userID, DisplayName: input.DisplayName, Service: service, Vehicle: input.Vehicle, Status: StatusPending}, nil
}

func TestSubmitNormalizesApplication(t *testing.T) {
	repository := &fakeRepository{service: ServiceOption{Code: "comfort", DisplayName: "Comfort"}}
	service := NewService(repository)
	userID := uuid.New()

	application, err := service.Submit(context.Background(), userID, ApplicationInput{
		DisplayName: "  Test Driver  ",
		ServiceCode: " COMFORT ",
		Vehicle: VehicleInput{
			Make:         " Toyota ",
			Model:        " Corolla ",
			ModelYear:    2024,
			Color:        " White ",
			LicensePlate: " abc-123 ",
		},
	})
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if application.Status != StatusPending {
		t.Fatalf("status = %q, want pending", application.Status)
	}
	if repository.created.DisplayName != "Test Driver" || repository.created.ServiceCode != "comfort" {
		t.Fatalf("application was not normalized: %#v", repository.created)
	}
	if repository.created.Vehicle.LicensePlate != "ABC-123" {
		t.Fatalf("license plate = %q, want ABC-123", repository.created.Vehicle.LicensePlate)
	}
}

func TestPrecheckRejectsModelYearBelowServiceRequirement(t *testing.T) {
	minimum := 2022
	repository := &fakeRepository{service: ServiceOption{Code: "comfort", DisplayName: "Comfort", MinimumModelYear: &minimum}}
	service := NewService(repository)

	result, err := service.Precheck(context.Background(), ApplicationInput{
		DisplayName: "Test Driver",
		ServiceCode: "comfort",
		Vehicle: VehicleInput{Make: "Toyota", Model: "Corolla", ModelYear: 2020, Color: "White", LicensePlate: "ABC-123"},
	})
	if err != nil {
		t.Fatalf("Precheck returned error: %v", err)
	}
	if result.Eligible || len(result.Reasons) == 0 {
		t.Fatalf("precheck = %#v, want ineligible with reason", result)
	}
}

func TestSubmitRejectsInvalidVehicleYear(t *testing.T) {
	repository := &fakeRepository{service: ServiceOption{Code: "economy", DisplayName: "Economy"}}
	service := NewService(repository)
	_, err := service.Submit(context.Background(), uuid.New(), ApplicationInput{
		DisplayName: "Test Driver",
		ServiceCode: "economy",
		Vehicle: VehicleInput{Make: "Toyota", Model: "Corolla", ModelYear: 1800, Color: "White", LicensePlate: "ABC-123"},
	})
	if err != ErrInvalidApplication {
		t.Fatalf("err = %v, want %v", err, ErrInvalidApplication)
	}
}
