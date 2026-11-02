package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type pollingOTPSender struct {
	mu sync.Mutex
	attempts map[uuid.UUID]int
	failFirst bool
	inFlight int
	maxInFlight int
}

func (s *pollingOTPSender) SendPhoneOTP(ctx context.Context, id uuid.UUID, phone, code string) error {
	s.mu.Lock()
	if s.attempts == nil {
		s.attempts = make(map[uuid.UUID]int)
	}
	s.attempts[id]++
	attempt := s.attempts[id]
	s.inFlight++
	if s.inFlight > s.maxInFlight {
		s.maxInFlight = s.inFlight
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.inFlight--
		s.mu.Unlock()
	}()
	if s.failFirst && attempt == 1 {
		return errors.New("temporary provider failure")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Millisecond):
		return nil
	}
}

func TestOTPDispatchWorkerScheduledRetryAndShutdown(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x91}, 32))
	if err != nil {
		t.Fatal(err)
	}
	// Enqueue alone does not schedule discovery; the atomic issuer does.
	limiter, err := NewOTPRateLimiter(store.client, bytes.Repeat([]byte{0x92}, 32))
	if err != nil {
		t.Fatal(err)
	}
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{0x93}, 32))
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := NewAtomicPhoneOTPIssuer(store, dispatch, limiter, codec)
	if err != nil {
		t.Fatal(err)
	}
	created, err := issuer.Issue(context.Background(), PhoneChallengeRequest{
		PhoneNumber: "+923001234589", Purpose: PhoneChallengeSignInOrRegister,
	}, uuid.NewString(), uuid.NewString())
	if err != nil || !created.ShouldSend {
		t.Fatalf("issue: %+v %v", created, err)
	}
	sender := &pollingOTPSender{failFirst: true}
	worker, err := NewOTPDispatchWorker(dispatch, sender, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx, 20*time.Millisecond, 10, 2) }()
	deadline := time.After(8 * time.Second)
	for {
		sender.mu.Lock()
		attempts := sender.attempts[created.ID]
		sender.mu.Unlock()
		if attempts >= 2 {
			break
		}
		select {
		case err := <-done:
			cancel()
			t.Fatalf("worker ended before retry: %v", err)
		case <-deadline:
			cancel()
			t.Fatal("scheduled retry did not occur")
		case <-time.After(20 * time.Millisecond):
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not shut down")
	}
	sender.mu.Lock()
	attempts := sender.attempts[created.ID]
	sender.mu.Unlock()
	if attempts != 2 {
		t.Fatalf("expected one failure and one retry, got %d", attempts)
	}
	if ids, err := dispatch.Due(context.Background(), 10); err != nil {
		t.Fatal(err)
	} else {
		for _, id := range ids {
			if id == created.ID {
				t.Fatal("successfully delivered challenge remained due")
			}
		}
	}
}

func TestOTPDispatchWorkerRejectsInvalidPollingConfig(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x94}, 32))
	if err != nil {
		t.Fatal(err)
	}
	worker, err := NewOTPDispatchWorker(dispatch, &pollingOTPSender{}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		interval time.Duration
		batch int64
		concurrency int
	}{
		{0, 1, 1}, {time.Millisecond, 0, 1},
		{time.Millisecond, 101, 1}, {time.Millisecond, 1, 0},
		{time.Millisecond, 1, 101},
	} {
		if err := worker.Run(context.Background(), tc.interval, tc.batch, tc.concurrency); !errors.Is(err, ErrPhoneOTPConfiguration) {
			t.Fatalf("expected configuration error for %+v: %v", tc, err)
		}
	}
}

func TestOTPDispatchWorkerConcurrentPollersClaimOnce(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x95}, 32))
	if err != nil {
		t.Fatal(err)
	}
	limiter, err := NewOTPRateLimiter(store.client, bytes.Repeat([]byte{0x96}, 32))
	if err != nil {
		t.Fatal(err)
	}
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{0x97}, 32))
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := NewAtomicPhoneOTPIssuer(store, dispatch, limiter, codec)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]uuid.UUID, 0, 8)
	for i := 0; i < 8; i++ {
		created, err := issuer.Issue(context.Background(), PhoneChallengeRequest{
			PhoneNumber: fmt.Sprintf("+923%09d", 100000000+i),
			Purpose: PhoneChallengeSignInOrRegister,
		}, uuid.NewString(), uuid.NewString())
		if err != nil || !created.ShouldSend {
			t.Fatalf("issue %d: %+v %v", i, created, err)
		}
		ids = append(ids, created.ID)
	}
	sender := &pollingOTPSender{}
	worker, err := NewOTPDispatchWorker(dispatch, sender, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- worker.Run(ctx, 10*time.Millisecond, 8, 2)
		}()
	}
	deadline := time.After(4 * time.Second)
	for {
		sender.mu.Lock()
		completed := len(sender.attempts)
		sender.mu.Unlock()
		if completed == len(ids) {
			break
		}
		select {
		case <-deadline:
			cancel()
			wg.Wait()
			t.Fatalf("only %d of %d deliveries attempted", completed, len(ids))
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("worker: %v", err)
		}
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	for _, id := range ids {
		if sender.attempts[id] != 1 {
			t.Errorf("challenge %s sent %d times", id, sender.attempts[id])
		}
	}
	if sender.maxInFlight > 4 {
		t.Errorf("global in-flight count %d exceeded both workers' combined cap", sender.maxInFlight)
	}
}
