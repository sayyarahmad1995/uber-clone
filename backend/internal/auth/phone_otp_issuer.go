package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrPhoneOTPDelivery = errors.New("phone OTP delivery failed")

// PhoneOTPSender delivers a code using an approved SMS provider. The provider
// adapter must use challengeID as its idempotency reference where supported.
// Never log a code or include it in HTTP responses.
type PhoneOTPSender interface {
	SendPhoneOTP(ctx context.Context, challengeID uuid.UUID, phone, code string) error
}

type phoneOTPChallengeWriter interface {
	Create(context.Context, NewPhoneChallenge) (PhoneChallengeCreation, error)
}

type phoneOTPLimiter interface {
	Allow(context.Context, string, string, int64, time.Duration) (time.Duration, error)
}

// PhoneOTPIssuer is not registered on an HTTP route until delivery recovery
// and trusted-source extraction are in place. It does not access PostgreSQL.
type PhoneOTPIssuer struct {
	store  phoneOTPChallengeWriter
	limit  phoneOTPLimiter
	codec  PhoneOTPCodec
	sender PhoneOTPSender
}

func NewPhoneOTPIssuer(store phoneOTPChallengeWriter, limit phoneOTPLimiter, codec PhoneOTPCodec, sender PhoneOTPSender) (PhoneOTPIssuer, error) {
	if store == nil || limit == nil || sender == nil || len(codec.key) < 32 {
		return PhoneOTPIssuer{}, ErrPhoneOTPConfiguration
	}
	return PhoneOTPIssuer{store: store, limit: limit, codec: codec, sender: sender}, nil
}

// Issue accepts only a requester identity established by the trusted HTTP
// boundary. Caller-generated idempotency keys must be unpredictable and stable
// across retries; a retry never sends a second SMS.
// The fixed limits are conservative initial defaults, not final abuse policy.
func (s PhoneOTPIssuer) Issue(ctx context.Context, request PhoneChallengeRequest, trustedSource, idempotencyKey string) (PhoneChallengeCreation, error) {
	normalized, err := request.Validate()
	if err != nil || trustedSource == "" || len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return PhoneChallengeCreation{}, ErrInvalidPhoneChallengeRequest
	}
	// Charge the request source first to avoid arbitrary phone-number probes
	// consuming another user's phone quota. Both counters live in Redis.
	if _, err := s.limit.Allow(ctx, "issue:source", trustedSource, 20, time.Hour); err != nil {
		return PhoneChallengeCreation{}, err
	}
	if _, err := s.limit.Allow(ctx, "issue:phone", normalized.PhoneNumber, 5, time.Hour); err != nil {
		return PhoneChallengeCreation{}, err
	}
	id := uuid.New()
	code, digest, err := s.codec.Generate(id)
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	now := time.Now()
	created, err := s.store.Create(ctx, NewPhoneChallenge{
		ID: id, PhoneNumber: normalized.PhoneNumber, Purpose: normalized.Purpose,
		CodeDigest: digest, IdempotencyKey: idempotencyKey,
		ExpiresAt: now.Add(5 * time.Minute), ResendAfter: now.Add(30 * time.Second),
	})
	if err != nil || !created.ShouldSend {
		return created, err
	}
	// Delivery is deliberately at-most-once from this synchronous path.
	// On provider failure, do not automatically retry with a fresh code:
	// a dispatch reservation/recovery protocol is required before activation.
	if err := s.sender.SendPhoneOTP(ctx, created.ID, normalized.PhoneNumber, code); err != nil {
		return PhoneChallengeCreation{}, errors.Join(ErrPhoneOTPDelivery, err)
	}
	return created, nil
}
