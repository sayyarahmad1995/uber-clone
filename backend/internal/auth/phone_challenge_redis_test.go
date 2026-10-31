package auth

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func redisOTPTestStore(t *testing.T) RedisPhoneChallengeStore {
	t.Helper()
	raw := os.Getenv("TEST_REDIS_URL")
	if raw == "" {
		t.Skip("TEST_REDIS_URL is not set")
	}
	opt, err := redis.ParseURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(opt)
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}
	store, err := NewRedisPhoneChallengeStore(client, bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func newRedisOTPChallenge(t *testing.T, id uuid.UUID, phone, idem string) (NewPhoneChallenge, []byte) {
	t.Helper()
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{0x31}, 32))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := codec.Digest(id, "012345")
	if err != nil {
		t.Fatal(err)
	}
	return NewPhoneChallenge{
		ID: id, PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister,
		CodeDigest: digest, IdempotencyKey: idem,
		ExpiresAt: time.Now().Add(90 * time.Second), ResendAfter: time.Now().Add(30 * time.Second),
	}, digest
}

func TestRedisOTPIdempotencyAndCooldown(t *testing.T) {
	store := redisOTPTestStore(t)
	ctx := context.Background()
	phone := "+923" + uuid.NewString()[:9] // replaced below with numeric fixture
	_ = phone
	phone = "+923001234567"
	// Unique idempotency values and a dedicated Redis test DB should isolate
	// these cases; do not run this against production Redis.
	scope := store.scope(phone, PhoneChallengeSignInOrRegister)
	_ = store.client.Del(ctx, scope+":cooldown").Err()
	a, _ := newRedisOTPChallenge(t, uuid.New(), phone, uuid.NewString())
	first, err := store.Create(ctx, a)
	if err != nil || !first.ShouldSend || first.ID != a.ID {
		t.Fatalf("first issuance: %+v %v", first, err)
	}
	replay, err := store.Create(ctx, a)
	if err != nil || replay.ShouldSend || replay.ID != a.ID {
		t.Fatalf("idempotent issuance: %+v %v", replay, err)
	}
	second, _ := newRedisOTPChallenge(t, uuid.New(), phone, uuid.NewString())
	if _, err := store.Create(ctx, second); !errors.Is(err, ErrPhoneChallengeCooldown) {
		t.Fatalf("cooldown bypassed: %v", err)
	}
}

func TestRedisOTPSingleUseAndAttemptLimit(t *testing.T) {
	store := redisOTPTestStore(t)
	ctx := context.Background()
	a, digest := newRedisOTPChallenge(t, uuid.New(), "+923001234568", uuid.NewString())
	_, _ = store.client.Del(ctx, store.scope(a.PhoneNumber, a.Purpose)+":cooldown").Result()
	if _, err := store.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	wrong := bytes.Repeat([]byte{0xab}, 32)
	for i := 0; i < 4; i++ {
		if _, err := store.Consume(ctx, a.ID, wrong); !errors.Is(err, ErrPhoneChallengeInvalid) {
			t.Fatalf("wrong attempt %d: %v", i, err)
		}
	}
	if phone, err := store.Consume(ctx, a.ID, digest); err != nil || phone != a.PhoneNumber {
		t.Fatalf("last allowed attempt: %s %v", phone, err)
	}
	if _, err := store.Consume(ctx, a.ID, digest); !errors.Is(err, ErrPhoneChallengeInvalid) {
		t.Fatalf("replay: %v", err)
	}
	b, correct := newRedisOTPChallenge(t, uuid.New(), "+923001234569", uuid.NewString())
	_, _ = store.client.Del(ctx, store.scope(b.PhoneNumber, b.Purpose)+":cooldown").Result()
	if _, err := store.Create(ctx, b); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		_, _ = store.Consume(ctx, b.ID, wrong)
	}
	if _, err := store.Consume(ctx, b.ID, correct); !errors.Is(err, ErrPhoneChallengeLimited) {
		t.Fatalf("attempt cap: %v", err)
	}
}

func TestRedisOTPConcurrentConsume(t *testing.T) {
	store := redisOTPTestStore(t)
	ctx := context.Background()
	a, digest := newRedisOTPChallenge(t, uuid.New(), "+923001234570", uuid.NewString())
	_, _ = store.client.Del(ctx, store.scope(a.PhoneNumber, a.Purpose)+":cooldown").Result()
	if _, err := store.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	const count = 8
	results := make(chan error, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Consume(ctx, a.ID, digest)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrPhoneChallengeInvalid) {
			t.Fatalf("unexpected result: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected one winner, got %d", successes)
	}
}
