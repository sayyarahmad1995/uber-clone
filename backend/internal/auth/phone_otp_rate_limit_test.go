package auth

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestRedisOTPRateLimit(t *testing.T) {
	store := redisOTPTestStore(t)
	limiter, err := NewOTPRateLimiter(store.client, bytes.Repeat([]byte{0x43}, 32))
	if err != nil { t.Fatal(err) }
	ctx := context.Background()
	subject := uuid.NewString()
	for i := 0; i < 3; i++ {
		if _, err := limiter.Allow(ctx, "phone", subject, 3, time.Minute); err != nil {
			t.Fatalf("allowed attempt %d: %v", i, err)
		}
	}
	retry, err := limiter.Allow(ctx, "phone", subject, 3, time.Minute)
	if !errors.Is(err, ErrPhoneOTPRateLimited) || retry <= 0 || retry > time.Minute {
		t.Fatalf("expected bounded retry duration, got %s %v", retry, err)
	}
	if _, err := limiter.Allow(ctx, "source", subject, 1, time.Minute); err != nil {
		t.Fatalf("separate dimension unexpectedly limited: %v", err)
	}
}

func TestRedisOTPRateLimitConcurrent(t *testing.T) {
	store := redisOTPTestStore(t)
	limiter, err := NewOTPRateLimiter(store.client, bytes.Repeat([]byte{0x44}, 32))
	if err != nil { t.Fatal(err) }
	const workers = 20
	var wg sync.WaitGroup
	results := make(chan error, workers)
	subject := uuid.NewString()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := limiter.Allow(context.Background(), "source", subject, 5, time.Minute)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil { success++ } else if !errors.Is(err, ErrPhoneOTPRateLimited) { t.Fatalf("unexpected error: %v", err) }
	}
	if success != 5 { t.Fatalf("expected exactly 5 allowed requests, got %d", success) }
}

func TestRedisOTPExpiredAndUnavailable(t *testing.T) {
	store := redisOTPTestStore(t)
	ctx := context.Background()
	id := uuid.New()
	c, digest := newRedisOTPChallenge(t, id, "+923001234573", uuid.NewString())
	c.ExpiresAt = time.Now().Add(180 * time.Millisecond)
	c.ResendAfter = time.Now().Add(100 * time.Millisecond)
	if _, err := store.Create(ctx, c); err != nil { t.Fatal(err) }
	time.Sleep(250 * time.Millisecond)
	if _, err := store.Consume(ctx, id, digest); !errors.Is(err, ErrPhoneChallengeInvalid) {
		t.Fatalf("expired code was accepted or returned wrong error: %v", err)
	}
	// A closed client simulates Redis being unavailable without impacting
	// shared test infrastructure. Do not silently accept a code on outage.
	opts := store.client.Options()
	closed := redis.NewClient(opts)
	_ = closed.Close()
	failStore, err := NewRedisPhoneChallengeStore(closed, bytes.Repeat([]byte{0x42}, 32))
	if err != nil { t.Fatal(err) }
	c2, d2 := newRedisOTPChallenge(t, uuid.New(), "+923001234574", uuid.NewString())
	if _, err := failStore.Create(ctx, c2); err == nil { t.Fatal("issuing during Redis outage must fail") }
	if _, err := failStore.Consume(ctx, id, d2); err == nil { t.Fatal("verification during Redis outage must fail") }
}
