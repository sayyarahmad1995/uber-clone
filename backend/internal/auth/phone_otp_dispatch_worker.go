package auth

import (
 "context"
 "errors"
 "time"

 "github.com/google/uuid"
)

// SMSSender is provider-neutral. Implementations MUST provide durable
// idempotency keyed by challenge ID. A successful call means accepted by the
// provider, not necessarily delivered to the handset.
type SMSSender interface {
 SendOTP(ctx context.Context, phone, code string, idempotencyKey uuid.UUID) error
}

// OTPDispatchQueue is the existing encrypted Redis queue's worker contract.
type OTPDispatchQueue interface {
 Due(context.Context,int64)([]uuid.UUID,error)
 Claim(context.Context,uuid.UUID,time.Duration)(string,string,string,error)
 Finish(context.Context,uuid.UUID,string,bool)error
}

var ErrSMSDeliveryNotConfigured = errors.New("SMS delivery sender not configured")

type OTPDispatchWorker struct {
 Queue OTPDispatchQueue
 Sender SMSSender
 Lease time.Duration
}

// ProcessDue handles a bounded batch. Provider acceptance is acknowledged
// only after Finish succeeds. Retrying uncertain results requires the sender's
// provider-side idempotency guarantee.
func (w OTPDispatchWorker) ProcessDue(ctx context.Context,limit int64)(int,error){
 if w.Queue==nil||w.Sender==nil{return 0,ErrSMSDeliveryNotConfigured}
 lease:=w.Lease
 if lease==0 {lease=30*time.Second}
 if lease<time.Second||lease>time.Minute{return 0,ErrPhoneOTPConfiguration}
 ids,err:=w.Queue.Due(ctx,limit)
 if err!=nil{return 0,err}
 processed:=0
 var first error
 for _,id:=range ids {
  if err=ctx.Err();err!=nil{return processed,err}
  owner,phone,code,claimErr:=w.Queue.Claim(ctx,id,lease)
  if errors.Is(claimErr,ErrOTPDispatchBusy)||errors.Is(claimErr,ErrOTPDispatchMissing){continue}
  if claimErr!=nil {if first==nil{first=claimErr};continue}
  // Do not log or expose the recipient or OTP.
  sendCtx,cancel:=context.WithTimeout(ctx,lease/2)
  sendErr:=w.Sender.SendOTP(sendCtx,phone,code,id)
  cancel()
  finishErr:=w.Queue.Finish(ctx,id,owner,sendErr==nil)
  if finishErr!=nil&&first==nil{first=finishErr}
  if sendErr!=nil&&first==nil{first=sendErr}
  processed++
 }
 return processed,first
}
