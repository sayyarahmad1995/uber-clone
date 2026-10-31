package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhoneChallengeStoreRejectsInvalidCreationWithoutDatabaseAccess(t *testing.T) {
	now := time.Now()
	valid := NewPhoneChallenge{
		ID: uuid.New(), PhoneNumber: "+923001234567",
		Purpose: PhoneChallengeSignInOrRegister,
		CodeDigest: make([]byte, 32),
		IdempotencyKey: "idempotency-key-0001",
		ExpiresAt: now.Add(5 * time.Minute),
		ResendAfter: now.Add(30 * time.Second),
	}
	tests := []struct {
		name string
		edit func(*NewPhoneChallenge)
	}{
		{"bad number", func(c *NewPhoneChallenge) { c.PhoneNumber = "03001234567" }},
		{"wrong purpose", func(c *NewPhoneChallenge) { c.Purpose = "email_change" }},
		{"missing challenge id", func(c *NewPhoneChallenge) { c.ID = uuid.Nil }},
		{"short digest", func(c *NewPhoneChallenge) { c.CodeDigest = make([]byte, 16) }},
		{"missing idempotency key", func(c *NewPhoneChallenge) { c.IdempotencyKey = "" }},
		{"expired challenge", func(c *NewPhoneChallenge) { c.ExpiresAt = now.Add(-time.Second) }},
		{"cooldown after expiry", func(c *NewPhoneChallenge) { c.ResendAfter = now.Add(6 * time.Minute) }},
	}
	store := NewPhoneChallengeStore(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.edit(&c)
			_, err := store.Create(context.Background(), c)
			if !errors.Is(err, ErrInvalidPhoneChallengeRequest) {
				t.Fatalf("got %v, want ErrInvalidPhoneChallengeRequest", err)
			}
		})
	}
}

func TestPhoneChallengeStoreRejectsMalformedCompletionWithoutDatabaseAccess(t *testing.T) {
	store := NewPhoneChallengeStore(nil)
	for _, tt := range []struct {
		id uuid.UUID
		digest []byte
	}{
		{id: uuid.Nil, digest: make([]byte, 32)},
		{id: uuid.New(), digest: nil},
		{id: uuid.New(), digest: make([]byte, 16)},
	} {
		_, err := store.Consume(context.Background(), tt.id, tt.digest)
		if !errors.Is(err, ErrPhoneChallengeInvalid) {
			t.Fatalf("got %v, want ErrPhoneChallengeInvalid", err)
		}
	}
}
