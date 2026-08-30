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
	Color        string
	LicensePlate string
}

type Vehicle struct {
	ID           uuid.UUID
	Make         string
	Model        string
	Color        string
	LicensePlate string
}

type Profile struct {
	UserID    uuid.UUID
	Status    Status
	IsOnline  bool
	Vehicle   Vehicle
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	UpsertProfile(ctx context.Context, userID uuid.UUID, vehicle VehicleInput) (Profile, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (Profile, error)
	SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Onboard(ctx context.Context, userID uuid.UUID, vehicle VehicleInput) (Profile, error) {
	vehicle = normalizeVehicle(vehicle)
	if vehicle.Make == "" || vehicle.Model == "" || vehicle.Color == "" || vehicle.LicensePlate == "" {
		return Profile{}, ErrInvalidProfile
	}
	return s.repository.UpsertProfile(ctx, userID, vehicle)
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
