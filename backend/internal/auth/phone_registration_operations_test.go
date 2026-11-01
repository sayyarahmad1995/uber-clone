package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestPhoneRegistrationOperationLifecycle(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	resolver, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	phone := "+923001234590"
	op, err := resolver.ReserveRegistration(ctx, phone)
	if err != nil || op.State != "reserved" {
		t.Fatalf("reserve: %+v %v", op, err)
	}
	retry, err := resolver.ReserveRegistration(ctx, phone)
	if err != nil || retry.ID != op.ID || retry.State != "reserved" {
		t.Fatalf("retry reservation: %+v %v", retry, err)
	}
	if _, err := resolver.CompleteRegistration(ctx, op.ID); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("completed without Kratos subject: %v", err)
	}
	claimed, claimErr := resolver.ClaimRegistrationSubmission(ctx,op.ID)
 if claimErr!=nil || !claimed { t.Fatalf("claim: %v %v",claimed,claimErr) }
 claimed, claimErr=resolver.ClaimRegistrationSubmission(ctx,op.ID)
 if claimErr!=nil || claimed { t.Fatalf("duplicate claim: %v %v",claimed,claimErr) }
 subject := uuid.NewString()
	confirmed, err := resolver.RecordRegistrationIdentity(ctx, op.ID, subject)
	if err != nil || confirmed.State != "identity_created" || confirmed.Subject != subject {
		t.Fatalf("record: %+v %v", confirmed, err)
	}
	confirmed, err = resolver.RecordRegistrationIdentity(ctx, op.ID, subject)
	if err != nil || confirmed.Subject != subject {
		t.Fatalf("record idempotency: %+v %v", confirmed, err)
	}
	if _, err = resolver.RecordRegistrationIdentity(ctx, op.ID, uuid.NewString()); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("subject replaced: %v", err)
	}
	userID, err := resolver.CompleteRegistration(ctx, op.ID)
	if err != nil || userID == uuid.Nil {
		t.Fatalf("complete: %s %v", userID, err)
	}
	again, err := resolver.CompleteRegistration(ctx, op.ID)
	if err != nil || again != userID {
		t.Fatalf("complete retry: %s %v", again, err)
	}
	current, err := resolver.GetRegistration(ctx, phone)
	if err != nil || current.State != "completed" || current.UserID != userID {
		t.Fatalf("current: %+v %v", current, err)
	}
	binding, err := resolver.ResolveVerifiedPhone(ctx, phone)
	if err != nil || !binding.Exists || binding.IdentitySubject != subject {
		t.Fatalf("binding: %+v %v", binding, err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("unexpected users count: %d %v", count, err)
	}
}

func TestPhoneRegistrationOperationDoesNotReplacePhoneOwner(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	resolver, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	phone := "+923001234591"
	owner := uuid.NewString()
	if _, err := resolver.CreatePhoneAccount(ctx, owner, phone); err != nil {
		t.Fatal(err)
	}
	op, err := resolver.ReserveRegistration(ctx, phone)
	if err != nil {
		t.Fatal(err)
	}
	if claimed, err := resolver.ClaimRegistrationSubmission(ctx,op.ID); err != nil || !claimed {t.Fatalf("claim: %v %v",claimed,err)}
 if _, err := resolver.RecordRegistrationIdentity(ctx, op.ID, uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.CompleteRegistration(ctx, op.ID); err == nil {
		t.Fatal("phone takeover unexpectedly committed")
	}
	current, err := resolver.GetRegistration(ctx, phone)
	if err != nil || current.State != "identity_created" {
		t.Fatalf("operation unexpectedly completed: %+v %v", current, err)
	}
}
