package driveronboarding

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidApplication    = errors.New("invalid driver onboarding application")
	ErrServiceNotFound       = errors.New("driver service not found")
	ErrPendingApplication    = errors.New("driver onboarding application already pending")
	ErrAlreadyOnboarded      = errors.New("driver onboarding already complete")
	ErrNotFound              = errors.New("driver onboarding application not found")
	ErrIneligible            = errors.New("vehicle is not eligible for selected service")
	ErrInvalidReviewDecision = errors.New("invalid driver onboarding review decision")
	ErrApplicationNotPending = errors.New("driver onboarding application is not pending")
)

type ApplicationStatus string

const (
	StatusPending  ApplicationStatus = "pending"
	StatusApproved ApplicationStatus = "approved"
	StatusRejected ApplicationStatus = "rejected"
)

type ServiceOption struct {
	Code               string
	DisplayName        string
	Description        string
	MinimumModelYear   *int
	ImpliedServiceCode *string
}

type VehicleInput struct {
	Make         string
	Model        string
	ModelYear    int
	Color        string
	LicensePlate string
}

type ApplicationInput struct {
	DisplayName string
	ServiceCode string
	Vehicle     VehicleInput
}

type Application struct {
	ID              uuid.UUID
	DriverUserID    uuid.UUID
	DisplayName     string
	Service         ServiceOption
	Vehicle         VehicleInput
	Status          ApplicationStatus
	RejectionReason string
	SubmittedAt     time.Time
	DecidedAt       *time.Time
	DecidedBy       string
}

type PrecheckResult struct {
	Eligible bool
	Reasons  []string
	Service  ServiceOption
}

type Repository interface {
	ListServices(ctx context.Context) ([]ServiceOption, error)
	FindService(ctx context.Context, code string) (ServiceOption, error)
	LatestApplication(ctx context.Context, userID uuid.UUID) (Application, error)
	CreateApplication(ctx context.Context, userID uuid.UUID, input ApplicationInput, service ServiceOption) (Application, error)
}

type ReviewRepository interface {
	ListPendingApplications(ctx context.Context) ([]Application, error)
	FindApplicationByID(ctx context.Context, applicationID uuid.UUID) (Application, error)
	ApproveApplication(ctx context.Context, applicationID uuid.UUID, reviewer string) (Application, error)
	RejectApplication(ctx context.Context, applicationID uuid.UUID, reviewer, reason string) (Application, error)
}

type Service struct{ repository Repository }

type ReviewService struct{ repository ReviewRepository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func NewReviewService(repository ReviewRepository) ReviewService {
	return ReviewService{repository: repository}
}

func (s Service) ListServices(ctx context.Context) ([]ServiceOption, error) {
	return s.repository.ListServices(ctx)
}

func (s Service) Latest(ctx context.Context, userID uuid.UUID) (Application, error) {
	return s.repository.LatestApplication(ctx, userID)
}

func (s Service) Precheck(ctx context.Context, input ApplicationInput) (PrecheckResult, error) {
	input = normalize(input)
	if err := validateInput(input); err != nil {
		return PrecheckResult{}, err
	}
	service, err := s.repository.FindService(ctx, input.ServiceCode)
	if err != nil {
		return PrecheckResult{}, err
	}
	result := PrecheckResult{Eligible: true, Service: service, Reasons: []string{}}
	if service.MinimumModelYear != nil && input.Vehicle.ModelYear < *service.MinimumModelYear {
		result.Eligible = false
		result.Reasons = append(result.Reasons, "vehicle model year does not meet the selected service requirement")
	}
	return result, nil
}

func (s Service) Submit(ctx context.Context, userID uuid.UUID, input ApplicationInput) (Application, error) {
	input = normalize(input)
	if err := validateInput(input); err != nil {
		return Application{}, err
	}
	precheck, err := s.Precheck(ctx, input)
	if err != nil {
		return Application{}, err
	}
	if !precheck.Eligible {
		return Application{}, ErrIneligible
	}
	return s.repository.CreateApplication(ctx, userID, input, precheck.Service)
}

func (s ReviewService) ListPending(ctx context.Context) ([]Application, error) {
	return s.repository.ListPendingApplications(ctx)
}

func (s ReviewService) Get(ctx context.Context, applicationID uuid.UUID) (Application, error) {
	if applicationID == uuid.Nil {
		return Application{}, ErrNotFound
	}
	return s.repository.FindApplicationByID(ctx, applicationID)
}

func (s ReviewService) Approve(ctx context.Context, applicationID uuid.UUID, reviewer string) (Application, error) {
	reviewer = strings.TrimSpace(reviewer)
	if applicationID == uuid.Nil || reviewer == "" {
		return Application{}, ErrInvalidReviewDecision
	}
	return s.repository.ApproveApplication(ctx, applicationID, reviewer)
}

func (s ReviewService) Reject(ctx context.Context, applicationID uuid.UUID, reviewer, reason string) (Application, error) {
	reviewer = strings.TrimSpace(reviewer)
	reason = strings.TrimSpace(reason)
	if applicationID == uuid.Nil || reviewer == "" || reason == "" {
		return Application{}, ErrInvalidReviewDecision
	}
	return s.repository.RejectApplication(ctx, applicationID, reviewer, reason)
}

func normalize(input ApplicationInput) ApplicationInput {
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.ServiceCode = strings.ToLower(strings.TrimSpace(input.ServiceCode))
	input.Vehicle.Make = strings.TrimSpace(input.Vehicle.Make)
	input.Vehicle.Model = strings.TrimSpace(input.Vehicle.Model)
	input.Vehicle.Color = strings.TrimSpace(input.Vehicle.Color)
	input.Vehicle.LicensePlate = strings.ToUpper(strings.TrimSpace(input.Vehicle.LicensePlate))
	return input
}

func validateInput(input ApplicationInput) error {
	currentYear := time.Now().UTC().Year()
	if input.DisplayName == "" || input.ServiceCode == "" || input.Vehicle.Make == "" || input.Vehicle.Model == "" || input.Vehicle.Color == "" || input.Vehicle.LicensePlate == "" || input.Vehicle.ModelYear < 1886 || input.Vehicle.ModelYear > currentYear+1 {
		return ErrInvalidApplication
	}
	return nil
}
