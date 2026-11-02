package auth

import (
 "context"
 "errors"
 "testing"
 "time"

 "github.com/google/uuid"
)
type fakeDispatchQueue struct{
 id uuid.UUID
 owner string
 claims int
 finishes int
 success bool
 failClaim error
 failFinish error
}
func(q *fakeDispatchQueue)Due(context.Context,int64)([]uuid.UUID,error){return []uuid.UUID{q.id},nil}
func(q *fakeDispatchQueue)Claim(context.Context,uuid.UUID,time.Duration)(string,string,string,error){
 q.claims++
 if q.failClaim!=nil{return "","","",q.failClaim}
 return "lease-owner","+923001234567","012345",nil
}
func(q *fakeDispatchQueue)Finish(_ context.Context,_ uuid.UUID,owner string,success bool)error{
 q.finishes++;q.owner=owner;q.success=success;return q.failFinish
}
type fakeSMSSender struct{calls int; key uuid.UUID; err error}
func(s *fakeSMSSender)SendOTP(_ context.Context,phone,code string,id uuid.UUID)error{
 s.calls++;s.key=id
 if phone!="+923001234567"||code!="012345"{return errors.New("corrupted dispatch payload")}
 return s.err
}
func TestOTPDispatchWorkerSuccessAndIdempotencyKey(t *testing.T){
 q:=&fakeDispatchQueue{id:uuid.New()}; sender:=&fakeSMSSender{}
 n,err:=(OTPDispatchWorker{Queue:q,Sender:sender}).ProcessDue(context.Background(),1)
 if err!=nil||n!=1||sender.calls!=1||sender.key!=q.id||q.owner!="lease-owner"||!q.success||q.finishes!=1{t.Fatalf("dispatch result: processed=%d err=%v queue=%+v sender=%+v",n,err,q,sender)}
}
func TestOTPDispatchWorkerFailureIsRetriedByQueue(t *testing.T){
 q:=&fakeDispatchQueue{id:uuid.New()};sender:=&fakeSMSSender{err:errors.New("provider timeout")}
 n,err:=(OTPDispatchWorker{Queue:q,Sender:sender}).ProcessDue(context.Background(),1)
 if n!=1||err==nil||q.success||q.finishes!=1 {t.Fatalf("failed send incorrectly acknowledged: n=%d err=%v queue=%+v",n,err,q)}
}
func TestOTPDispatchWorkerMissingProviderAndBusyLease(t *testing.T){
 q:=&fakeDispatchQueue{id:uuid.New(),failClaim:ErrOTPDispatchBusy}
 if _,err:=(OTPDispatchWorker{Queue:q}).ProcessDue(context.Background(),1);!errors.Is(err,ErrSMSDeliveryNotConfigured){t.Fatalf("missing provider: %v",err)}
 sender:=&fakeSMSSender{}
 n,err:=(OTPDispatchWorker{Queue:q,Sender:sender}).ProcessDue(context.Background(),1)
 if err!=nil||n!=0||sender.calls!=0||q.finishes!=0{t.Fatalf("leased dispatch unexpectedly sent: n=%d err=%v",n,err)}
}
