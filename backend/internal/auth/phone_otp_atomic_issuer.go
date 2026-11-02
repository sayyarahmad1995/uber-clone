package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// atomicPhoneOTPWriter prevents production issuance from using Create followed
// by Enqueue, which would leave a crash window between the two operations.
type atomicPhoneOTPWriter interface {
	CreateWithDispatch(context.Context, NewPhoneChallenge, RedisOTPDispatch, string) (PhoneChallengeCreation, error)
}

// AtomicPhoneOTPIssuer is the preferred issuance path. It persists the OTP
// and encrypted SMS payload atomically; a separate worker performs delivery.
// It is intentionally not registered on any HTTP route.
type AtomicPhoneOTPIssuer struct {
	store    atomicPhoneOTPWriter
	dispatch RedisOTPDispatch
	limit    phoneOTPLimiter
	codec    PhoneOTPCodec
}

func NewAtomicPhoneOTPIssuer(store atomicPhoneOTPWriter, dispatch RedisOTPDispatch, limiter phoneOTPLimiter, codec PhoneOTPCodec) (AtomicPhoneOTPIssuer, error) {
	if store == nil || limiter == nil || dispatch.client == nil || dispatch.aead == nil || len(codec.key) < 32 {
		return AtomicPhoneOTPIssuer{}, ErrPhoneOTPConfiguration
	}
	return AtomicPhoneOTPIssuer{store: store, dispatch: dispatch, limit: limiter, codec: codec}, nil
}

func (s AtomicPhoneOTPIssuer) Issue(ctx context.Context, request PhoneChallengeRequest, trustedSource, idempotencyKey string) (PhoneChallengeCreation, error) {
	normalized, err := request.Validate()
	if err != nil || trustedSource == "" || len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		return PhoneChallengeCreation{}, ErrInvalidPhoneChallengeRequest
	}
	if _, err = s.limit.Allow(ctx, "issue:source", trustedSource, 20, time.Hour); err != nil {
		return PhoneChallengeCreation{}, err
	}
	if _, err = s.limit.Allow(ctx, "issue:phone", normalized.PhoneNumber, 5, time.Hour); err != nil {
		return PhoneChallengeCreation{}, err
	}
	id := uuid.New()
	code, digest, err := s.codec.Generate(id)
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	now := time.Now()
	return s.store.CreateWithDispatch(ctx, NewPhoneChallenge{
		ID: id, PhoneNumber: normalized.PhoneNumber, Purpose: normalized.Purpose,
		CodeDigest: digest, IdempotencyKey: idempotencyKey,
		ExpiresAt: now.Add(5 * time.Minute), ResendAfter: now.Add(30 * time.Second),
	}, s.dispatch, code)
}

// OTPDispatchWorker processes a known challenge ID. The caller must provide
// reliable discovery/scheduling of IDs before enabling automated recovery.
// The sender must enforce challenge-ID idempotency for uncertain outcomes.
type OTPDispatchWorker struct {
	dispatch RedisOTPDispatch
	sender   PhoneOTPSender
	lease    time.Duration
}

func NewOTPDispatchWorker(dispatch RedisOTPDispatch, sender PhoneOTPSender, lease time.Duration) (OTPDispatchWorker, error) {
	if dispatch.client == nil || dispatch.aead == nil || sender == nil || lease < time.Second || lease > time.Minute {
		return OTPDispatchWorker{}, ErrPhoneOTPConfiguration
	}
	return OTPDispatchWorker{dispatch: dispatch, sender: sender, lease: lease}, nil
}

func (w OTPDispatchWorker) Deliver(ctx context.Context, id uuid.UUID) error {
	owner, phone, code, err := w.dispatch.Claim(ctx, id, w.lease)
	if err != nil {
		return err
	}
	// Use the same challenge ID as the provider idempotency reference on
	// every attempt, including attempts after a worker crash.
	// Reserve time for Finish before the non-renewing lease expires.
	// A sender must honor cancellation and deduplicate by challenge UUID.
	sendCtx, sendCancel := context.WithTimeout(ctx, w.lease/2)
	sendErr := w.sender.SendPhoneOTP(sendCtx, id, phone, code)
	sendCancel()
	// A canceled request must not prevent the worker from recording delivery.
	// Bound the acknowledgement independently instead of using Background forever.
	ackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	// An ambiguous provider response must not trigger an automatic retry.
	// Marking it complete sacrifices automatic recovery for this OTP but
	// prevents duplicate messages when Twilio accepted an unseen POST.
	finishErr := w.dispatch.Finish(ackCtx, id, owner, sendErr == nil || errors.Is(sendErr, ErrPhoneOTPSendUncertain))
	if sendErr != nil {
		return errors.Join(ErrPhoneOTPDelivery, sendErr, finishErr)
	}
	return finishErr
}

// Run polls the existing Redis due index without creating a second dispatch
// mechanism. Each batch completes before the next poll; at most concurrency
// sends are in flight. The caller controls lifecycle via ctx. Provider failures
// are retryable queue outcomes, not fatal scheduler errors.
func (w OTPDispatchWorker) Run(ctx context.Context, interval time.Duration, batchSize int64, concurrency int) error {
	if interval <= 0 || batchSize < 1 || batchSize > 100 || concurrency < 1 || concurrency > 100 ||
		w.dispatch.client == nil || w.dispatch.aead == nil || w.sender == nil {
		return ErrPhoneOTPConfiguration
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	poll := func() error {
		ids, err := w.dispatch.Due(ctx, batchSize)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		for _, id := range ids {
			select {
			case <-ctx.Done():
				wg.Wait()
				return ctx.Err()
			case sem <- struct{}{}:
			}
			wg.Add(1)
			go func(id uuid.UUID) {
				defer wg.Done()
				defer func() { <-sem }()
				// Busy/missing leases and provider failures are expected; Redis
				// retains retry state. Never retry directly in the scheduler.
				_ = w.Deliver(ctx, id)
			}(id)
		}
		wg.Wait()
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		if err := poll(); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
