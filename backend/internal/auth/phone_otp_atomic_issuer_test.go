package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAtomicPhoneOTPIssuerAndWorker(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x71}, 32))
	if err != nil {
		t.Fatal(err)
	}
	limiter, err := NewOTPRateLimiter(store.client, bytes.Repeat([]byte{0x72}, 32))
	if err != nil {
		t.Fatal(err)
	}
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{0x73}, 32))
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := NewAtomicPhoneOTPIssuer(store, dispatch, limiter, codec)
	if err != nil {
		t.Fatal(err)
	}
	sender := &issuerSender{}
	worker, err := NewOTPDispatchWorker(dispatch, sender, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	phone := fmt.Sprintf("+923%09d", uuid.New().ID()%1000000000)
	ctx := context.Background()
	request := PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}
	idem := uuid.NewString()
	source := uuid.NewString()
	created, err := issuer.Issue(ctx, request, source, idem)
	if err != nil || !created.ShouldSend {
		t.Fatalf("issue: %+v %v", created, err)
	}
	if sender.calls != 0 {
		t.Fatal("issuer performed SMS delivery")
	}
	replayed, err := issuer.Issue(ctx, request, source, idem)
	if err != nil || replayed.ShouldSend || replayed.ID != created.ID {
		t.Fatalf("replay: %+v %v", replayed, err)
	}
	if err := worker.Deliver(ctx, created.ID); err != nil {
		t.Fatalf("delivery: %v", err)
	}
	if sender.calls != 1 || sender.phone != phone || len(sender.code) != 6 {
		t.Fatal("SMS sender arguments")
	}
	if err := worker.Deliver(ctx, created.ID); !errors.Is(err, ErrOTPDispatchMissing) {
		t.Fatalf("duplicate send: %v", err)
	}
	if sender.calls != 1 {
		t.Fatal("duplicate SMS sent")
	}
	digest, err := codec.Digest(created.ID, sender.code)
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := store.Consume(ctx, created.ID, digest); err != nil || verified != phone {
		t.Fatalf("OTP verification: %s %v", verified, err)
	}
}

func TestAtomicPhoneOTPWorkerDeliveryFailure(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x74}, 32))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	if ok, err := dispatch.Enqueue(context.Background(), id, "+923001234579", "123456", time.Minute); err != nil || !ok {
		t.Fatalf("enqueue: %v %v", ok, err)
	}
	sender := &issuerSender{err: errors.New("provider unavailable")}
	worker, err := NewOTPDispatchWorker(dispatch, sender, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Deliver(context.Background(), id); !errors.Is(err, ErrPhoneOTPDelivery) {
		t.Fatalf("failure: %v", err)
	}
	sender.err = nil
	if err := worker.Deliver(context.Background(), id); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if sender.calls != 2 {
		t.Fatalf("expected 2 provider attempts, got %d", sender.calls)
	}
}
