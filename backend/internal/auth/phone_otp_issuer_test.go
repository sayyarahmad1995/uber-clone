package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type issuerStore struct {
	result PhoneChallengeCreation
	err    error
	calls  int
}
func (s *issuerStore) Create(_ context.Context, c NewPhoneChallenge) (PhoneChallengeCreation, error) {
	s.calls++
	if s.err != nil { return PhoneChallengeCreation{}, s.err }
	if s.result.ID != uuid.Nil { return s.result, nil }
	return PhoneChallengeCreation{ID:c.ID,ShouldSend:true,ResendAfter:c.ResendAfter},nil
}

type issuerLimiter struct {
	calls []string
	failOn string
}
func (l *issuerLimiter) Allow(_ context.Context, dimension, _ string, _ int64, _ time.Duration) (time.Duration,error) {
	l.calls=append(l.calls,dimension)
	if dimension==l.failOn { return time.Second,ErrPhoneOTPRateLimited }
	return 0,nil
}
type issuerSender struct {
	calls int
	phone string
	code string
	err error
}
func (s *issuerSender) SendPhoneOTP(_ context.Context, id uuid.UUID, phone, code string) error {
	s.calls++
	s.phone=phone
	s.code=code
	if id==uuid.Nil { return errors.New("missing ID") }
	return s.err
}

func testIssuer(t *testing.T, store *issuerStore, limiter *issuerLimiter, sender *issuerSender) PhoneOTPIssuer {
	t.Helper()
	codec,err:=NewPhoneOTPCodec(bytes.Repeat([]byte{0x55},32))
	if err!=nil { t.Fatal(err) }
	issuer,err:=NewPhoneOTPIssuer(store,limiter,codec,sender)
	if err!=nil { t.Fatal(err) }
	return issuer
}
func validIssueRequest() PhoneChallengeRequest {
	return PhoneChallengeRequest{PhoneNumber:"+923001234567",Purpose:PhoneChallengeSignInOrRegister}
}

func TestPhoneOTPIssuerSendsOnlyForNewChallenge(t *testing.T) {
	store:=&issuerStore{}
	limits:=&issuerLimiter{}
	sender:=&issuerSender{}
	issuer:=testIssuer(t,store,limits,sender)
	out,err:=issuer.Issue(context.Background(),validIssueRequest(),"203.0.113.1",uuid.NewString())
	if err!=nil||!out.ShouldSend||sender.calls!=1||store.calls!=1 { t.Fatalf("issue: %+v %v sender=%d store=%d",out,err,sender.calls,store.calls) }
	if len(sender.code)!=6||sender.phone!="+923001234567" { t.Fatalf("invalid sender arguments") }
	if len(limits.calls)!=2||limits.calls[0]!="issue:source"||limits.calls[1]!="issue:phone" { t.Fatalf("limit order: %v",limits.calls) }
}

func TestPhoneOTPIssuerNoSendOnReplayOrRateLimit(t *testing.T) {
	store:=&issuerStore{result:PhoneChallengeCreation{ID:uuid.New(),ShouldSend:false}}
	limiter:=&issuerLimiter{}
	sender:=&issuerSender{}
	issuer:=testIssuer(t,store,limiter,sender)
	if _,err:=issuer.Issue(context.Background(),validIssueRequest(),"203.0.113.1",uuid.NewString());err!=nil {t.Fatal(err)}
	if sender.calls!=0 {t.Fatal("idempotent replay dispatched SMS")}

	limiter.failOn="issue:source"
	if _,err:=issuer.Issue(context.Background(),validIssueRequest(),"203.0.113.1",uuid.NewString());!errors.Is(err,ErrPhoneOTPRateLimited) {t.Fatalf("rate limit: %v",err)}
	if store.calls!=1||sender.calls!=0 {t.Fatal("rate-limited request reached store or sender")}

	limiter.failOn="issue:phone"
	if _,err:=issuer.Issue(context.Background(),validIssueRequest(),"203.0.113.1",uuid.NewString());!errors.Is(err,ErrPhoneOTPRateLimited) {t.Fatalf("phone rate limit: %v",err)}
	if store.calls!=1||sender.calls!=0 {t.Fatal("phone-limited request reached store or sender")}
}

func TestPhoneOTPIssuerDeliveryErrorFailsClosed(t *testing.T) {
	store:=&issuerStore{}
	sender:=&issuerSender{err:errors.New("provider unavailable")}
	issuer:=testIssuer(t,store,&issuerLimiter{},sender)
	if _,err:=issuer.Issue(context.Background(),validIssueRequest(),"203.0.113.1",uuid.NewString());!errors.Is(err,ErrPhoneOTPDelivery) {t.Fatalf("expected delivery failure: %v",err)}
	if sender.calls!=1 {t.Fatal("unexpected dispatch count")}
}

func TestPhoneOTPIssuerRejectsInvalidBeforeSideEffects(t *testing.T) {
	store:=&issuerStore{}
	limiter:=&issuerLimiter{}
	sender:=&issuerSender{}
	issuer:=testIssuer(t,store,limiter,sender)
	if _,err:=issuer.Issue(context.Background(),validIssueRequest(),"",uuid.NewString());!errors.Is(err,ErrInvalidPhoneChallengeRequest) {t.Fatalf("invalid source: %v",err)}
	if _,err:=issuer.Issue(context.Background(),PhoneChallengeRequest{PhoneNumber:"123"}, "203.0.113.1",uuid.NewString());!errors.Is(err,ErrInvalidPhoneChallengeRequest) {t.Fatalf("invalid phone: %v",err)}
	if len(limiter.calls)!=0||store.calls!=0||sender.calls!=0 {t.Fatal("invalid request caused side effect")}
}
