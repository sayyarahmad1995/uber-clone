package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type failedReserveRepository struct{}

func (failedReserveRepository) ReserveRegistration(context.Context, string) (RegistrationOperation, error) {
	return RegistrationOperation{}, errors.New("postgres unavailable")
}
func (failedReserveRepository) ClaimRegistrationSubmission(context.Context, uuid.UUID) (bool, error) {
	panic("unexpected claim")
}
func (failedReserveRepository) RecordRegistrationIdentity(context.Context, uuid.UUID, string) (RegistrationOperation, error) {
	panic("unexpected record")
}
func (failedReserveRepository) CompleteRegistration(context.Context, uuid.UUID) (uuid.UUID, error) {
	panic("unexpected completion")
}

func TestRegistrationCoordinatorDoesNotConsumeProofWhenReservationFails(t *testing.T) {
	proof := &coordinatorProof{phone: "+923001234595"}
	creator := &coordinatorKratos{}
	svc := PhoneRegistrationCoordinator{Proofs: proof, Identities: creator, Operations: failedReserveRepository{}}
	_, err := svc.Register(context.Background(), "verified-token", "secure-password")
	if err == nil {
		t.Fatal("expected PostgreSQL failure")
	}
	if proof.calls != 0 {
		t.Fatalf("consumed proof before durable reservation: %d", proof.calls)
	}
	if creator.calls != 0 {
		t.Fatal("called Kratos without durable reservation")
	}
}

type failedClaimRepository struct{}

func (failedClaimRepository) ReserveRegistration(_ context.Context, phone string) (RegistrationOperation, error) {
	return RegistrationOperation{ID: uuid.New(), Phone: phone, State: "reserved"}, nil
}
func (failedClaimRepository) ClaimRegistrationSubmission(context.Context, uuid.UUID) (bool, error) {
	return false, errors.New("postgres claim unavailable")
}
func (failedClaimRepository) RecordRegistrationIdentity(context.Context, uuid.UUID, string) (RegistrationOperation, error) {
	panic("unexpected record")
}
func (failedClaimRepository) CompleteRegistration(context.Context, uuid.UUID) (uuid.UUID, error) {
	panic("unexpected completion")
}

func TestRegistrationCoordinatorDoesNotConsumeProofWhenClaimFails(t *testing.T) {
	proof := &coordinatorProof{phone: "+923001234597"}
	creator := &coordinatorKratos{}
	service := PhoneRegistrationCoordinator{Proofs: proof, Identities: creator, Operations: failedClaimRepository{}}
	if _, err := service.Register(context.Background(), "token", "password"); err == nil {
		t.Fatal("expected PostgreSQL claim failure")
	}
	if proof.calls != 0 || creator.calls != 0 {
		t.Fatalf("proof consumed or Kratos called without durable claim: %d %d", proof.calls, creator.calls)
	}
}
