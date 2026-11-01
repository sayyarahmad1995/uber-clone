package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCreatePhoneAccountTransaction(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	resolver, err := NewPostgresPhoneAccountResolver(store.db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	phone := "+923001234588"
	subject := uuid.NewString()
	userID, err := resolver.CreatePhoneAccount(ctx, subject, phone)
	if err != nil || userID == uuid.Nil {
		t.Fatalf("create: %s %v", userID, err)
	}
	second, err := resolver.CreatePhoneAccount(ctx, subject, phone)
	if err != nil || second != userID {
		t.Fatalf("retry: %s %v", second, err)
	}
	var userCount, capabilityCount int
	if err := store.db.QueryRowContext(ctx, `
		SELECT count(*) FROM users WHERE id=$1
	`, userID).Scan(&userCount); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRowContext(ctx, `
		SELECT count(*) FROM user_capabilities WHERE user_id=$1 AND capability='rider'
	`, userID).Scan(&capabilityCount); err != nil {
		t.Fatal(err)
	}
	if userCount != 1 || capabilityCount != 1 {
		t.Fatalf("records: users=%d riders=%d", userCount, capabilityCount)
	}
	if _, err := resolver.CreatePhoneAccount(ctx, uuid.NewString(), phone); err == nil {
		t.Fatal("phone takeover was allowed")
	}
	if _, err := resolver.CreatePhoneAccount(ctx, subject, "+923001234589"); !errors.Is(err, ErrPhoneIdentityBinding) {
		t.Fatalf("existing identity phone change accepted: %v", err)
	}
	var users int
	if err := store.db.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 1 {
		t.Fatalf("failed registration left an orphan: %d", users)
	}
}
