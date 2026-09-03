package driver

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile = errors.New("invalid driver profile")
	ErrNotFound       = errors.New("driver profile not found")
)

type Status string

const StatusActive Status = "active"

type VehicleInput struct {
	Make         string
	Model        string
	ModelYear    int
	Color        string
	LicensePlate string
}

type OnboardingInput struct {
	DisplayName string
	Vehicle     VehicleInput
}

type Vehicle struct {
	ID           uuid.UUID
	Make         string
	Model        string
	ModelYear    int
	Color        string
	LicensePlate string
}

type Profile struct {
	UserID      uuid.UUID
	DisplayName string
	Status      Status
	IsOnline    bool
	Vehicle     Vehicle
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	UpsertProfile(ctx context.Context, userID uuid.UUID, input OnboardingInput) (Profile, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (Profile, error)
	SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Onboard(ctx context.Context, userID uuid.UUID, input OnboardingInput) (Profile, error) {
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Vehicle = normalizeVehicle(input.Vehicle)
	currentYear := time.Now().UTC().Year()
	if input.DisplayName == "" || input.Vehicle.Make == "" || input.Vehicle.Model == "" || input.Vehicle.Color == "" || input.Vehicle.LicensePlate == "" || input.Vehicle.ModelYear < 1886 || input.Vehicle.ModelYear > currentYear+1 {
		return Profile{}, ErrInvalidProfile
	}
	return s.repository.UpsertProfile(ctx, userID, input)
}

func (s Service) Get(ctx context.Context, userID uuid.UUID) (Profile, error) {
	return s.repository.FindByUserID(ctx, userID)
}

func (s Service) SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error) {
	return s.repository.SetOnline(ctx, userID, online)
}

func normalizeVehicle(vehicle VehicleInput) VehicleInput {
	vehicle.Make = strings.TrimSpace(vehicle.Make)
	vehicle.Model = strings.TrimSpace(vehicle.Model)
	vehicle.Color = strings.TrimSpace(vehicle.Color)
	vehicle.LicensePlate = strings.ToUpper(strings.TrimSpace(vehicle.LicensePlate))
	return vehicle
}
