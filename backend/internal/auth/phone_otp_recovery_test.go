package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Recovery is deliberately caller-driven. Production must only enable it
// when the configured SMS provider deduplicates on challenge ID.
func (w OTPDispatchWorker) RecoverDue(ctx context.Context, limit int64) (int, error) {
	ids, err := w.dispatch.Due(ctx, limit)
	if err != nil { return 0, err }
	processed := 0
	var firstErr error
	for _, id := range ids {
		if err := ctx.Err(); err != nil { return processed, err }
		err := w.Deliver(ctx, id)
		if errors.Is(err, ErrOTPDispatchBusy) || errors.Is(err, ErrOTPDispatchMissing) { continue }
		if err != nil && firstErr == nil { firstErr = err }
		processed++
	}
	return processed, firstErr
}

func TestRedisOTPDispatchDueRecovery(t *testing.T) {
	store:=redisOTPTestStore(t)
	dispatch,err:=NewRedisOTPDispatch(store.client,bytes.Repeat([]byte{0x81},32))
	if err!=nil {t.Fatal(err)}
	codec,err:=NewPhoneOTPCodec(bytes.Repeat([]byte{0x82},32))
	if err!=nil {t.Fatal(err)}
	ctx:=context.Background()
	phone:="+923001234580"
	_ = store.client.Del(ctx,store.scope(phone,PhoneChallengeSignInOrRegister)+":cooldown").Err()
	id:=uuid.New()
	c,digest:=newRedisOTPChallenge(t,id,phone,uuid.NewString())
	result,err:=store.CreateWithDispatch(ctx,c,dispatch,"012345")
	if err!=nil || !result.ShouldSend {t.Fatalf("issue: %+v %v",result,err)}
	_ = codec
	_ = digest
	ids,err:=dispatch.Due(ctx,10)
	if err!=nil{t.Fatal(err)}
	found:=false
	for _,candidate:=range ids{if candidate==id{found=true}}
	if !found{t.Fatal("newly issued dispatch missing from recovery queue")}
	sender:=&issuerSender{}
	worker,err:=NewOTPDispatchWorker(dispatch,sender,10*time.Second)
	if err!=nil{t.Fatal(err)}
	n,err:=worker.RecoverDue(ctx,10)
	if err!=nil || n<1{t.Fatalf("recovery: count=%d err=%v",n,err)}
	if sender.calls!=1 {t.Fatalf("expected exactly one SMS, got %d",sender.calls)}
	ids,err=dispatch.Due(ctx,10)
	if err!=nil{t.Fatal(err)}
	for _,candidate:=range ids{if candidate==id{t.Fatal("completed dispatch still scheduled")}}
}
