package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPhoneAccountPostgresBinding(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	resolver, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	checker, err := NewPhoneAccountChecker(resolver)
	if err != nil {
		t.Fatal(err)
	}
	phone := "+923001234583"
	decision, err := checker.Decide(ctx, phone)
	if err != nil || decision.NextStep != PhoneAuthRegister {
		t.Fatalf("unbound: %+v %v", decision, err)
	}
	userID := uuid.New()
	subject := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, "INSERT INTO users (id) VALUES ($1)", userID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx,
		"INSERT INTO external_identities (user_id,issuer,subject) VALUES ($1,$2,$3)",
		userID, "primary-identity-v1", subject); err != nil {
		t.Fatal(err)
	}
	if err := resolver.BindVerifiedPhone(ctx, userID, subject, phone); err != nil {
		t.Fatal(err)
	}
	if err := resolver.BindVerifiedPhone(ctx, userID, subject, phone); err != nil {
		t.Fatalf("idempotent bind: %v", err)
	}
	decision, err = checker.Decide(ctx, phone)
	if err != nil || decision.NextStep != PhoneAuthLogin || decision.IdentitySubject != subject {
		t.Fatalf("linked: %+v %v", decision, err)
	}
	anotherUser := uuid.New()
	anotherSubject := uuid.NewString()
	if _, err := store.db.ExecContext(ctx, "INSERT INTO users (id) VALUES ($1)", anotherUser); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx,
		"INSERT INTO external_identities (user_id,issuer,subject) VALUES ($1,$2,$3)",
		anotherUser, "primary-identity-v1", anotherSubject); err != nil {
		t.Fatal(err)
	}
	if err := resolver.BindVerifiedPhone(ctx, anotherUser, anotherSubject, phone); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("phone takeover: %v", err)
	}
	if err := resolver.BindVerifiedPhone(ctx, anotherUser, anotherSubject, "+923001234584"); err != nil {
		t.Fatal(err)
	}
	// An identity cannot own two verified phones in this first-version policy.
	err = resolver.BindVerifiedPhone(ctx, userID, subject, "+923001234585")
	var pgerr *pgconn.PgError
	if !errors.As(err, &pgerr) || pgerr.Code != "23505" {
		t.Fatalf("expected unique user constraint: %v", err)
	}
	// A subject linked to another user cannot be used to create a binding.
	if err := resolver.BindVerifiedPhone(ctx, anotherUser, subject, "+923001234586"); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("unowned identity accepted: %v", err)
	}
}

func TestPhoneAccountPostgresRejectsMissingIdentity(t *testing.T) {
	store := phoneChallengeTestStore(t)
	resolver, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	if err := resolver.BindVerifiedPhone(context.Background(), uuid.New(), "unregistered-subject", "+923001234587"); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("missing identity accepted: %v", err)
	}
}
