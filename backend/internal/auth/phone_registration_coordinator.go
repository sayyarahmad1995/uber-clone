package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrPhoneRegistrationPending = errors.New("phone registration requires reconciliation")

// RegistrationProofConsumer deletes the verified phone-possession proof once.
// A proof is authorization to begin one registration, never a login session.
type RegistrationProofConsumer interface {
	PeekProof(context.Context, string, PhoneChallengePurpose) (PhoneVerificationProof, error)
	ConsumeProof(context.Context, string, PhoneChallengePurpose) (PhoneVerificationProof, error)
}
type RegistrationIdentityCreator interface {
	RegisterPhoneIdentity(context.Context, string, string) (PhoneRegistrationIdentity, error)
}
type PhoneRegistrationIdentity struct{ IdentitySubject string }
type RegistrationRepository interface {
	ReserveRegistration(context.Context, string) (RegistrationOperation, error)
	ClaimRegistrationSubmission(context.Context, uuid.UUID) (bool, error)
	RecordRegistrationIdentity(context.Context, uuid.UUID, string) (RegistrationOperation, error)
	CompleteRegistration(context.Context, uuid.UUID) (uuid.UUID, error)
}

// PhoneRegistrationCoordinator performs only an initial, proof-authorized
// registration. Retrying a potentially dispatched Kratos call is forbidden.
// A separate privileged reconciler must verify orphan identities with Kratos.
type PhoneRegistrationCoordinator struct {
	Proofs     RegistrationProofConsumer
	Identities RegistrationIdentityCreator
	Operations RegistrationRepository
}

func (s PhoneRegistrationCoordinator) Register(ctx context.Context, proofToken, password string) (uuid.UUID, error) {
	if s.Proofs == nil || s.Identities == nil || s.Operations == nil {
		return uuid.Nil, ErrPhoneOTPConfiguration
	}
	if password == "" {
		return uuid.Nil, ErrRegistrationInvalid
	}
	proof, err := s.Proofs.PeekProof(ctx, proofToken, PhoneChallengeSignInOrRegister)
	if err != nil {
		return uuid.Nil, err
	}
	phone, err := (PhoneChallengeRequest{PhoneNumber: proof.PhoneNumber, Purpose: proof.Purpose}).Validate()
	if err != nil || proof.Purpose != PhoneChallengeSignInOrRegister {
		return uuid.Nil, ErrPhoneVerificationProof
	}
	operation, err := s.Operations.ReserveRegistration(ctx, phone.PhoneNumber)
	if err != nil {
		return uuid.Nil, err
	}
	// Persist recovery state before irreversibly consuming the proof.
	// A proof must never permit a second Kratos creation or replay completion.
	if operation.State != "reserved" {
		return uuid.Nil, ErrPhoneRegistrationPending
	}
	consumed, err := s.Proofs.ConsumeProof(ctx, proofToken, PhoneChallengeSignInOrRegister)
	if err != nil {
		return uuid.Nil, err
	}
	if consumed.PhoneNumber != phone.PhoneNumber || consumed.Purpose != proof.Purpose {
		return uuid.Nil, ErrPhoneVerificationProof
	}
	claimed, err := s.Operations.ClaimRegistrationSubmission(ctx, operation.ID)
	if err != nil {
		return uuid.Nil, err
	}
	if !claimed {
		return uuid.Nil, ErrPhoneRegistrationPending
	}
	created, err := s.Identities.RegisterPhoneIdentity(ctx, phone.PhoneNumber, password)
	if err != nil {
		return uuid.Nil, ErrPhoneRegistrationPending
	}
	if created.IdentitySubject == "" {
		return uuid.Nil, ErrPhoneRegistrationPending
	}
	if _, err = s.Operations.RecordRegistrationIdentity(ctx, operation.ID, created.IdentitySubject); err != nil {
		return uuid.Nil, ErrPhoneRegistrationPending
	}
	userID, err := s.Operations.CompleteRegistration(ctx, operation.ID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}
