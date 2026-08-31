package matching

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	result        Result
	candidate     Candidate
	err           error
	rideID        uuid.UUID
	actorUserID   uuid.UUID
	rejectInvoked bool
}

func (f *fakeRepository) Match(_ context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	f.rideID = rideRequestID
	f.actorUserID = riderUserID
	return f.result, f.err
}

func (f *fakeRepository) Reject(_ context.Context, rideRequestID, driverUserID uuid.UUID) (Candidate, error) {
	f.rideID = rideRequestID
	f.actorUserID = driverUserID
	f.rejectInvoked = true
	return f.candidate, f.err
}

func TestServiceMatchDelegatesApplicationOwnedIDs(t *testing.T) {
	rideID := uuid.New()
	riderID := uuid.New()
	driverID := uuid.New()
	createdAt := time.Now().UTC()
	repository := &fakeRepository{result: Result{
		Candidate: Candidate{
			RideRequestID: rideID,
			DriverUserID:  driverID,
			Status:        CandidateStatusPending,
			CreatedAt:     createdAt,
		},
		Created: true,
	}}

	result, err := NewService(repository).Match(context.Background(), rideID, riderID)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if repository.rideID != rideID || repository.actorUserID != riderID {
		t.Fatalf("unexpected delegated ids: ride=%s rider=%s", repository.rideID, repository.actorUserID)
	}
	if !result.Created || result.Candidate.DriverUserID != driverID {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestServiceMatchPreservesRepositoryBusinessError(t *testing.T) {
	repository := &fakeRepository{err: ErrNoEligibleDriver}
	_, err := NewService(repository).Match(context.Background(), uuid.New(), uuid.New())
	if err != ErrNoEligibleDriver {
		t.Fatalf("expected ErrNoEligibleDriver, got %v", err)
	}
}

func TestServiceRejectDelegatesDriverDecision(t *testing.T) {
	rideID := uuid.New()
	driverID := uuid.New()
	repository := &fakeRepository{candidate: Candidate{
		RideRequestID: rideID,
		DriverUserID:  driverID,
		Status:        CandidateStatusRejected,
	}}

	candidate, err := NewService(repository).Reject(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if !repository.rejectInvoked || repository.rideID != rideID || repository.actorUserID != driverID {
		t.Fatalf("unexpected delegated ids: ride=%s driver=%s", repository.rideID, repository.actorUserID)
	}
	if candidate.Status != CandidateStatusRejected {
		t.Fatalf("unexpected candidate status: %q", candidate.Status)
	}
}

func TestServiceRejectPreservesRepositoryBusinessError(t *testing.T) {
	repository := &fakeRepository{err: ErrCandidateResolved}
	_, err := NewService(repository).Reject(context.Background(), uuid.New(), uuid.New())
	if err != ErrCandidateResolved {
		t.Fatalf("expected ErrCandidateResolved, got %v", err)
	}
}
