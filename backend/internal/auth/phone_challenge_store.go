package auth

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPhoneChallengeInvalid = errors.New("phone challenge invalid or expired")
	ErrPhoneChallengeLimited = errors.New("phone challenge attempt limit reached")
	ErrPhoneChallengeCooldown = errors.New("phone challenge resend cooldown active")
)

// PhoneChallengeStore persists OTP challenges. Creation is intentionally separate
// from SMS delivery; callers must not send an SMS for an idempotent replay.
type PhoneChallengeStore struct{ db *sql.DB }

func NewPhoneChallengeStore(db *sql.DB) PhoneChallengeStore { return PhoneChallengeStore{db: db} }

type NewPhoneChallenge struct {
	ID             uuid.UUID
	PhoneNumber    string
	Purpose        PhoneChallengePurpose
	CodeDigest     []byte // HMAC-SHA256 with a secret not stored in PostgreSQL
	IdempotencyKey string
	ExpiresAt      time.Time
	ResendAfter    time.Time
}

type PhoneChallengeCreation struct {
	ID          uuid.UUID
	ShouldSend  bool
	ResendAfter time.Time
}

// Create creates a challenge only once for the same idempotency key and returns
// ShouldSend=false for replays. It does not implement per-phone/IP throttling:
// issuance must not be exposed until those controls and SMS delivery are wired.
func (s PhoneChallengeStore) Create(ctx context.Context, c NewPhoneChallenge) (PhoneChallengeCreation, error) {
	normalized, err := (PhoneChallengeRequest{PhoneNumber: c.PhoneNumber, Purpose: c.Purpose}).Validate()
	if err != nil || c.ID == uuid.Nil || len(c.CodeDigest) != 32 ||
		len(c.IdempotencyKey) < 16 || len(c.IdempotencyKey) > 128 ||
		!c.ExpiresAt.After(time.Now()) || !c.ResendAfter.After(time.Now()) ||
		c.ResendAfter.After(c.ExpiresAt) {
		return PhoneChallengeCreation{}, ErrInvalidPhoneChallengeRequest
	}
	var result PhoneChallengeCreation
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO auth_phone_challenges
			(id, phone_number, purpose, code_digest, idempotency_key, expires_at, resend_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (phone_number, purpose, idempotency_key) DO NOTHING
		RETURNING id, resend_after
	`, c.ID, normalized.PhoneNumber, c.Purpose, c.CodeDigest, c.IdempotencyKey,
		c.ExpiresAt, c.ResendAfter).Scan(&result.ID, &result.ResendAfter)
	if err == nil {
		result.ShouldSend = true
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PhoneChallengeCreation{}, err
	}
	// A conflicting insert is already committed when the INSERT returns.
	// Read its stored challenge without modifying it or triggering another SMS.
	err = s.db.QueryRowContext(ctx, `
		SELECT id, resend_after FROM auth_phone_challenges
		WHERE phone_number = $1 AND purpose = $2 AND idempotency_key = $3
	`, normalized.PhoneNumber, c.Purpose, c.IdempotencyKey).Scan(&result.ID, &result.ResendAfter)
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	return result, nil
}

// Consume atomically checks an OTP digest and marks its challenge as used.
// PostgreSQL row locking prevents simultaneous submissions from succeeding.
// Wrong attempts are counted even when the code does not match.
func (s PhoneChallengeStore) Consume(ctx context.Context, id uuid.UUID, candidateDigest []byte) (string, error) {
	if id == uuid.Nil || len(candidateDigest) != 32 {
		return "", ErrPhoneChallengeInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var phone string
	var expected []byte
	var attempts, maximum int
	var expiresAt time.Time
	var consumedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT phone_number, code_digest, attempt_count, max_attempts, expires_at, consumed_at
		FROM auth_phone_challenges WHERE id = $1 FOR UPDATE
	`, id).Scan(&phone, &expected, &attempts, &maximum, &expiresAt, &consumedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPhoneChallengeInvalid
	}
	if err != nil {
		return "", err
	}
	if consumedAt.Valid || !expiresAt.After(time.Now()) {
		return "", ErrPhoneChallengeInvalid
	}
	if attempts >= maximum {
		return "", ErrPhoneChallengeLimited
	}
	match := subtle.ConstantTimeCompare(expected, candidateDigest) == 1
	if match {
		_, err = tx.ExecContext(ctx, `
			UPDATE auth_phone_challenges SET attempt_count = attempt_count + 1,
				consumed_at = clock_timestamp() WHERE id = $1
		`, id)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE auth_phone_challenges SET attempt_count = attempt_count + 1
			WHERE id = $1
		`, id)
	}
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	if !match {
		return "", ErrPhoneChallengeInvalid
	}
	return phone, nil
}
