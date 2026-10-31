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

func TestRedisOTPDispatchLifecycle(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x61}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id := uuid.New()
	ok, err := dispatch.Enqueue(ctx, id, "+923001234567", "012345", 90*time.Second)
	if err != nil || !ok {
		t.Fatalf("enqueue: %v %v", ok, err)
	}
	ok, err = dispatch.Enqueue(ctx, id, "+923001234567", "999999", 90*time.Second)
	if err != nil || ok {
		t.Fatalf("duplicate enqueue: %v %v", ok, err)
	}
	owner, phone, code, err := dispatch.Claim(ctx, id, 10*time.Second)
	if err != nil || phone != "+923001234567" || code != "012345" {
		t.Fatalf("claim: %q %q %v", phone, code, err)
	}
	if _, _, _, err := dispatch.Claim(ctx, id, 10*time.Second); !errors.Is(err, ErrOTPDispatchBusy) {
		t.Fatalf("double claim: %v", err)
	}
	if err := dispatch.Finish(ctx, id, uuid.NewString(), true); !errors.Is(err, ErrOTPDispatchOwnership) {
		t.Fatalf("wrong owner: %v", err)
	}
	if err := dispatch.Finish(ctx, id, owner, false); err != nil {
		t.Fatal(err)
	}
	owner, _, code, err = dispatch.Claim(ctx, id, 10*time.Second)
	if err != nil || code != "012345" {
		t.Fatalf("retry: %q %v", code, err)
	}
	if err := dispatch.Finish(ctx, id, owner, true); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := dispatch.Claim(ctx, id, 10*time.Second); !errors.Is(err, ErrOTPDispatchMissing) {
		t.Fatalf("completed delivery reclaimed: %v", err)
	}
}

func TestRedisOTPDispatchConcurrentClaim(t *testing.T) {
	store := redisOTPTestStore(t)
	d, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x62}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id := uuid.New()
	if _, err := d.Enqueue(ctx, id, "+923001234568", "123456", time.Minute); err != nil {
		t.Fatal(err)
	}
	const count = 12
	var wg sync.WaitGroup
	results := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _, err := d.Claim(ctx, id, 10*time.Second)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	wins := 0
	for err := range results {
		if err == nil {
			wins++
		} else if !errors.Is(err, ErrOTPDispatchBusy) {
			t.Fatalf("unexpected claim result: %v", err)
		}
	}
	if wins != 1 {
		t.Fatalf("expected one claim, got %d", wins)
	}
}

func TestRedisOTPDispatchExpirationAndOutage(t *testing.T) {
	store := redisOTPTestStore(t)
	d, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x63}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id := uuid.New()
	if _, err := d.Enqueue(ctx, id, "+923001234569", "654321", 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(170 * time.Millisecond)
	if _, _, _, err := d.Claim(ctx, id, time.Second); !errors.Is(err, ErrOTPDispatchMissing) {
		t.Fatalf("expired payload available: %v", err)
	}
	closed := store.client
	// A canceled context models an unavailable cache operation without
	// closing shared integration-test connections.
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := d.Enqueue(canceled, uuid.New(), "+923001234570", "111111", time.Minute); err == nil {
		t.Fatal("outage allowed enqueue")
	}
	_ = closed
}
