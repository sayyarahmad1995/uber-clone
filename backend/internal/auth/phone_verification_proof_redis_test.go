package auth

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func proofTestCodec(t *testing.T) PhoneOTPCodec {
	t.Helper()
	c,err:=NewPhoneOTPCodec(bytes.Repeat([]byte{0x91},32))
	if err!=nil {t.Fatal(err)}
	return c
}

func TestRedisOTPProofSingleUse(t *testing.T) {
	store:=redisOTPTestStore(t)
	proofs,err:=NewRedisPhoneProofs(store.client)
	if err!=nil {t.Fatal(err)}
	codec:=proofTestCodec(t)
	ctx:=context.Background()
	id:=uuid.New()
	phone:="+923001234581"
	digest,err:=codec.Digest(id,"012345")
	if err!=nil {t.Fatal(err)}
	c:=NewPhoneChallenge{ID:id,PhoneNumber:phone,Purpose:PhoneChallengeSignInOrRegister,CodeDigest:digest,IdempotencyKey:uuid.NewString(),ExpiresAt:time.Now().Add(time.Minute),ResendAfter:time.Now().Add(30*time.Second)}
	_ = store.client.Del(ctx,store.scope(phone,c.Purpose)+":cooldown").Err()
	if _,err:=store.Create(ctx,c);err!=nil{t.Fatal(err)}
	if _,err:=store.VerifyAndIssueProof(ctx,id,"000000",c.Purpose,codec);!errors.Is(err,ErrPhoneChallengeInvalid){t.Fatalf("wrong code: %v",err)}
	if _,err:=store.VerifyAndIssueProof(ctx,id,"012345",PhoneChallengePurpose("change_email"),codec);!errors.Is(err,ErrPhoneVerificationProof){t.Fatalf("purpose mismatch: %v",err)}
	token,err:=store.VerifyAndIssueProof(ctx,id,"012345",c.Purpose,codec)
	if err!=nil || len(token)!=64 {t.Fatalf("issue proof: %v",err)}
	if _,err:=store.VerifyAndIssueProof(ctx,id,"012345",c.Purpose,codec);!errors.Is(err,ErrPhoneChallengeInvalid){t.Fatalf("OTP replay: %v",err)}
	got,err:=proofs.ConsumeProof(ctx,token,c.Purpose)
	if err!=nil || got.PhoneNumber!=phone || got.Purpose!=c.Purpose {t.Fatalf("proof: %+v %v",got,err)}
	if _,err:=proofs.ConsumeProof(ctx,token,c.Purpose);!errors.Is(err,ErrPhoneVerificationProof){t.Fatalf("proof replay: %v",err)}
}

func TestRedisOTPProofConcurrentConsumeAndExpiration(t *testing.T) {
	store:=redisOTPTestStore(t)
	proofs,err:=NewRedisPhoneProofs(store.client)
	if err!=nil{t.Fatal(err)}
	ctx:=context.Background()
	token:=strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", "")
	// Proofs are never written directly by production callers; this fixture
	// tests only the atomic single-use consumption and expiration behavior.
	if err:=store.client.Set(ctx,proofKey(token),"+923001234582|sign_in_or_register",time.Minute).Err();err!=nil{t.Fatal(err)}
	const workers=10
	results:=make(chan error,workers)
	var wg sync.WaitGroup
	for i:=0;i<workers;i++{
		wg.Add(1)
		go func(){defer wg.Done();_,err:=proofs.ConsumeProof(ctx,token,PhoneChallengeSignInOrRegister);results<-err}()
	}
	wg.Wait();close(results)
	success:=0
	for err:=range results{
		if err==nil {success++} else if !errors.Is(err,ErrPhoneVerificationProof) {t.Fatalf("unexpected: %v",err)}
	}
	if success!=1{t.Fatalf("expected one proof consumer, got %d",success)}
	expiring:=strings.ReplaceAll(uuid.NewString()+uuid.NewString(), "-", "")
	if err:=store.client.Set(ctx,proofKey(expiring),"+923001234582|sign_in_or_register",80*time.Millisecond).Err();err!=nil{t.Fatal(err)}
	time.Sleep(120*time.Millisecond)
	if _,err:=proofs.ConsumeProof(ctx,expiring,PhoneChallengeSignInOrRegister);!errors.Is(err,ErrPhoneVerificationProof){t.Fatalf("expired proof accepted: %v",err)}
}
