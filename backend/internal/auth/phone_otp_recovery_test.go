package auth

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRedisOTPDispatchDueRecovery(t *testing.T) {
	store := redisOTPTestStore(t)
	dispatch, err := NewRedisOTPDispatch(store.client, bytes.Repeat([]byte{0x81}, 32))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Dedicated TEST_REDIS_URL database only: clear earlier integration-test jobs.
	if err := store.client.Del(ctx, redisOTPDispatchQueueKey).Err(); err != nil {
		t.Fatal(err)
	}
	phone := "+923001234580"
	_ = store.client.Del(ctx, store.scope(phone, PhoneChallengeSignInOrRegister)+":cooldown").Err()
	id := uuid.New()
	c, _ := newRedisOTPChallenge(t, id, phone, uuid.NewString())
	result, err := store.CreateWithDispatch(ctx, c, dispatch, "012345")
	if err != nil || !result.ShouldSend {
		t.Fatalf("issue: %+v %v", result, err)
	}
	ids, err := dispatch.Due(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, candidate := range ids {
		if candidate == id {
			found = true
		}
	}
	if !found {
		t.Fatal("newly issued dispatch missing from recovery queue")
	}
	sender := &issuerSender{}
	worker, err := NewOTPDispatchWorker(dispatch, sender, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	n, err := worker.RecoverDue(ctx, 100)
	if err != nil || n < 1 {
		t.Fatalf("recovery: count=%d err=%v", n, err)
	}
	if sender.calls != 1 {
		t.Fatalf("expected exactly one SMS, got %d", sender.calls)
	}
	ids, err = dispatch.Due(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range ids {
		if candidate == id {
			t.Fatal("completed dispatch still scheduled")
		}
	}
}
