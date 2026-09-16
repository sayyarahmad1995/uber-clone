package driver

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile = errors.New("invalid driver profile")
	ErrNotFound       = errors.New("driver profile not found")
)

type Status string

const (
	StatusApproved Status = "approved"
	StatusActive   Status = "active"
)

type ServiceEnrollment struct {
	ServiceCode   string
	DisplayName   string
	ApprovedAt    time.Time
	ServiceActive bool
}

type Vehicle struct {
	Enrollments  []ServiceEnrollment
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
	OperatingState(context.Context, uuid.UUID) (OperatingState, error)
	SelectOperation(context.Context, uuid.UUID, uuid.UUID, string) (OperatingState, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (Profile, error)
	ListVehicles(ctx context.Context, userID uuid.UUID) ([]Vehicle, error)
	SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) Get(ctx context.Context, userID uuid.UUID) (Profile, error) {
	return s.repository.FindByUserID(ctx, userID)
}

func (s Service) ListVehicles(ctx context.Context, userID uuid.UUID) ([]Vehicle, error) {
	return s.repository.ListVehicles(ctx, userID)
}

func (s Service) SetOnline(ctx context.Context, userID uuid.UUID, online bool) (Profile, error) {
	return s.repository.SetOnline(ctx, userID, online)
}
