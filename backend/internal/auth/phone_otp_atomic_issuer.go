package auth

import (
	"context"
	"errors"
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
	store atomicPhoneOTPWriter
	dispatch RedisOTPDispatch
	limit phoneOTPLimiter
	codec PhoneOTPCodec
}

func NewAtomicPhoneOTPIssuer(store atomicPhoneOTPWriter, dispatch RedisOTPDispatch, limiter phoneOTPLimiter, codec PhoneOTPCodec) (AtomicPhoneOTPIssuer, error) {
	if store == nil || limiter == nil || dispatch.client == nil || dispatch.aead == nil || len(codec.key) < 32 {
		return AtomicPhoneOTPIssuer{}, ErrPhoneOTPConfiguration
	}
	return AtomicPhoneOTPIssuer{store:store,dispatch:dispatch,limit:limiter,codec:codec},nil
}

func (s AtomicPhoneOTPIssuer) Issue(ctx context.Context, request PhoneChallengeRequest, trustedSource, idempotencyKey string) (PhoneChallengeCreation,error) {
	normalized,err:=request.Validate()
	if err!=nil || trustedSource=="" || len(idempotencyKey)<16 || len(idempotencyKey)>128 {
		return PhoneChallengeCreation{},ErrInvalidPhoneChallengeRequest
	}
	if _,err=s.limit.Allow(ctx,"issue:source",trustedSource,20,time.Hour);err!=nil{return PhoneChallengeCreation{},err}
	if _,err=s.limit.Allow(ctx,"issue:phone",normalized.PhoneNumber,5,time.Hour);err!=nil{return PhoneChallengeCreation{},err}
	id:=uuid.New()
	code,digest,err:=s.codec.Generate(id)
	if err!=nil{return PhoneChallengeCreation{},err}
	now:=time.Now()
	return s.store.CreateWithDispatch(ctx,NewPhoneChallenge{
		ID:id,PhoneNumber:normalized.PhoneNumber,Purpose:normalized.Purpose,
		CodeDigest:digest,IdempotencyKey:idempotencyKey,
		ExpiresAt:now.Add(5*time.Minute),ResendAfter:now.Add(30*time.Second),
	},s.dispatch,code)
}

// OTPDispatchWorker processes a known challenge ID. The caller must provide
// reliable discovery/scheduling of IDs before enabling automated recovery.
// The sender must enforce challenge-ID idempotency for uncertain outcomes.
type OTPDispatchWorker struct {
	dispatch RedisOTPDispatch
	sender PhoneOTPSender
	lease time.Duration
}

func NewOTPDispatchWorker(dispatch RedisOTPDispatch, sender PhoneOTPSender, lease time.Duration) (OTPDispatchWorker,error) {
	if dispatch.client==nil || dispatch.aead==nil || sender==nil || lease<time.Second || lease>time.Minute {
		return OTPDispatchWorker{},ErrPhoneOTPConfiguration
	}
	return OTPDispatchWorker{dispatch:dispatch,sender:sender,lease:lease},nil
}

func(w OTPDispatchWorker) Deliver(ctx context.Context,id uuid.UUID) error {
	owner,phone,code,err:=w.dispatch.Claim(ctx,id,w.lease)
	if err!=nil{return err}
	// Use the same challenge ID as the provider idempotency reference on
	// every attempt, including attempts after a worker crash.
	sendErr:=w.sender.SendPhoneOTP(ctx,id,phone,code)
	// A canceled request must not prevent the worker from recording delivery.
	// Bound the acknowledgement independently instead of using Background forever.
	ackCtx,cancel:=context.WithTimeout(context.WithoutCancel(ctx),2*time.Second)
	defer cancel()
	finishErr:=w.dispatch.Finish(ackCtx,id,owner,sendErr==nil)
	if sendErr!=nil {
		return errors.Join(ErrPhoneOTPDelivery,sendErr,finishErr)
	}
	return finishErr
}
