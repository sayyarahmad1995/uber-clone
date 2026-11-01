package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type recoveryVerifier struct {
	allowed bool
	err     error
	calls   int
}

func (v *recoveryVerifier) VerifyPhoneIdentity(_ context.Context, subject, phone string) (bool, error) {
	v.calls++
	return v.allowed, v.err
}

func TestReconcilePhoneRegistrationRequiresKratosIdentityMatch(t *testing.T) {
	store := phoneChallengeTestStore(t)
	repo, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	phone := "+923001234594"
	op, err := repo.ReserveRegistration(ctx, phone)
	if err != nil {
		t.Fatal(err)
	}
	subject := uuid.NewString()
	verifier := &recoveryVerifier{allowed: true}
	if _, err = ReconcilePhoneRegistration(ctx, repo, verifier, phone, op.ID, subject); !errors.Is(err, ErrPhoneRegistrationReconciliation) {
		t.Fatalf("reserved operation reconciled: %v", err)
	}
	claimed, err := repo.ClaimRegistrationSubmission(ctx, op.ID)
	if err != nil || !claimed {
		t.Fatalf("claim: %v %v", claimed, err)
	}
	verifier.allowed = false
	if _, err = ReconcilePhoneRegistration(ctx, repo, verifier, phone, op.ID, subject); !errors.Is(err, ErrPhoneRegistrationReconciliation) {
		t.Fatalf("mismatched Kratos trait accepted: %v", err)
	}
	verifier.allowed = true
	if _, err = ReconcilePhoneRegistration(ctx, repo, verifier, phone, uuid.New(), subject); !errors.Is(err, ErrPhoneRegistrationReconciliation) {
		t.Fatalf("incorrect operation accepted: %v", err)
	}
	uid, err := ReconcilePhoneRegistration(ctx, repo, verifier, phone, op.ID, subject)
	if err != nil || uid == uuid.Nil {
		t.Fatalf("recover: %v %v", uid, err)
	}
	again, err := ReconcilePhoneRegistration(ctx, repo, verifier, phone, op.ID, subject)
	if err != nil || again != uid {
		t.Fatalf("retry: %v %v", again, err)
	}
	if _, err = ReconcilePhoneRegistration(ctx, repo, verifier, phone, op.ID, uuid.NewString()); !errors.Is(err, ErrPhoneRegistrationReconciliation) {
		t.Fatalf("subject substitution accepted: %v", err)
	}
}
