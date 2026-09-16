package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

var (
	ErrNotOpen            = errors.New("ride request marketplace is not open")
	ErrOfferNotActionable = errors.New("ride offer is not actionable")
	ErrDriverUnavailable  = errors.New("driver is no longer available")
)

type AssignmentRepository interface {
	SelectOffer(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
		time.Time,
	) (trip.Trip, error)
}

type AssignmentService interface {
	SelectOffer(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
		time.Time,
	) (trip.Trip, error)
}

type assignmentService struct {
	repository AssignmentRepository
}

type PostgresAssignmentRepository struct {
	db *sql.DB
}

func NewPostgresAssignmentRepository(db *sql.DB) PostgresAssignmentRepository {
	return PostgresAssignmentRepository{db: db}
}

func NewAssignmentService(repository AssignmentRepository) AssignmentService {
	return assignmentService{repository: repository}
}

func (s assignmentService) SelectOffer(
	ctx context.Context,
	rideRequestID uuid.UUID,
	riderUserID uuid.UUID,
	driverUserID uuid.UUID,
	expectedVersion time.Time,
) (trip.Trip, error) {
	return s.repository.SelectOffer(
		ctx,
		rideRequestID,
		riderUserID,
		driverUserID,
		expectedVersion,
	)
}
