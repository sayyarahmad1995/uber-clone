package matching

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeRepository struct {
	result         Result
	candidate      Candidate
	err            error
	rideID         uuid.UUID
	actorUserID    uuid.UUID
	decision       CandidateStatus
	decideInvoked  bool
}

func (f *fakeRepository) Match(_ context.Context, rideRequestID, riderUserID uuid.UUID) (Result, error) {
	f.rideID = rideRequestID
	f.actorUserID = riderUserID
	return f.result, f.err
}

func (f *fakeRepository) Decide(_ context.Context, rideRequestID, driverUserID uuid.UUID, decision CandidateStatus) (Candidate, error) {
	f.rideID = rideRequestID
	f.actorUserID = driverUserID
	f.decision = decision
	f.decideInvoked = true
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

func TestServiceAcceptDelegatesDriverDecision(t *testing.T) {
	rideID := uuid.New()
	driverID := uuid.New()
	repository := &fakeRepository{candidate: Candidate{
		RideRequestID: rideID,
		DriverUserID:  driverID,
		Status:        CandidateStatusAccepted,
	}}

	candidate, err := NewService(repository).Accept(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !repository.decideInvoked || repository.rideID != rideID || repository.actorUserID != driverID {
		t.Fatalf("unexpected delegated ids: ride=%s driver=%s", repository.rideID, repository.actorUserID)
	}
	if repository.decision != CandidateStatusAccepted || candidate.Status != CandidateStatusAccepted {
		t.Fatalf("unexpected decision: delegated=%q candidate=%q", repository.decision, candidate.Status)
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
	if repository.decision != CandidateStatusRejected || candidate.Status != CandidateStatusRejected {
		t.Fatalf("unexpected decision: delegated=%q candidate=%q", repository.decision, candidate.Status)
	}
}

func TestServiceDecisionPreservesRepositoryBusinessError(t *testing.T) {
	repository := &fakeRepository{err: ErrCandidateResolved}
	_, err := NewService(repository).Accept(context.Background(), uuid.New(), uuid.New())
	if err != ErrCandidateResolved {
		t.Fatalf("expected ErrCandidateResolved, got %v", err)
	}
}
