package auth

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRedisOTPAtomicDispatchCreation(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x67}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	phone := "+923001234575"
	_ = store.client.Del(ctx, store.scope(phone, PhoneChallengeSignInOrRegister)+":cooldown").Err()
	c, digest := newRedisOTPChallenge(t, uuid.New(), phone, uuid.NewString())
	out, err := store.CreateWithDispatch(ctx, c, dispatch, "012345")
	if err != nil || !out.ShouldSend || out.ID != c.ID {
		t.Fatalf("atomic issue: %+v %v", out, err)
	}
	replay, err := store.CreateWithDispatch(ctx, c, dispatch, "999999")
	if err != nil || replay.ShouldSend || replay.ID != c.ID {
		t.Fatalf("idempotent issue: %+v %v", replay, err)
	}
	owner, gotPhone, code, err := dispatch.Claim(ctx, c.ID, time.Second*10)
	if err != nil || gotPhone != phone || code != "012345" {
		t.Fatalf("dispatch: %q %q %v", gotPhone, code, err)
	}
	if err := dispatch.Finish(ctx, c.ID, owner, true); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Consume(ctx, c.ID, digest); err != nil || got != phone {
		t.Fatalf("verify: %s %v", got, err)
	}
	second, _ := newRedisOTPChallenge(t, uuid.New(), phone, uuid.NewString())
	if _, err := store.CreateWithDispatch(ctx, second, dispatch, "012345"); !errors.Is(err, ErrPhoneChallengeCooldown) {
		t.Fatalf("cooldown: %v", err)
	}
	if _, _, _, err := dispatch.Claim(ctx, second.ID, time.Second); !errors.Is(err, ErrOTPDispatchMissing) {
		t.Fatalf("orphan dispatch: %v", err)
	}
}

func TestRedisOTPAtomicDispatchConcurrentIssue(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x68}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	phone := "+923001234576"
	_ = store.client.Del(ctx, store.scope(phone, PhoneChallengeSignInOrRegister)+":cooldown").Err()
	const workers = 8
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, _ := newRedisOTPChallenge(t, uuid.New(), phone, uuid.NewString())
			_, err := store.CreateWithDispatch(ctx, c, dispatch, "012345")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrPhoneChallengeCooldown) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 {
		t.Fatalf("expected one issuance, got %d", success)
	}
}

func TestRedisOTPAtomicDispatchExpiresTogether(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x69}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id := uuid.New()
	c, digest := newRedisOTPChallenge(t, id, "+923001234577", uuid.NewString())
	c.ExpiresAt = time.Now().Add(250 * time.Millisecond)
	c.ResendAfter = time.Now().Add(100 * time.Millisecond)
	if _, err := store.CreateWithDispatch(ctx, c, dispatch, "012345"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(350 * time.Millisecond)
	if _, err := store.Consume(ctx, id, digest); !errors.Is(err, ErrPhoneChallengeInvalid) {
		t.Fatalf("expired challenge: %v", err)
	}
	if _, _, _, err := dispatch.Claim(ctx, id, time.Second); !errors.Is(err, ErrOTPDispatchMissing) {
		t.Fatalf("expired dispatch: %v", err)
	}
}
