package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrOTPDispatchBusy = errors.New("OTP SMS dispatch already claimed")
	ErrOTPDispatchMissing = errors.New("OTP SMS dispatch missing or expired")
	ErrOTPDispatchOwnership = errors.New("OTP SMS dispatch lease not owned")
)

// RedisOTPDispatch coordinates recoverable SMS sends. The same challenge ID
// MUST be passed as the provider idempotency key on every retry. Without
// provider-side idempotency, delivery after an uncertain outcome cannot be
// retried safely.
type RedisOTPDispatch struct {
	client *redis.Client
	aead cipher.AEAD
}

func NewRedisOTPDispatch(client *redis.Client, encryptionKey []byte) (RedisOTPDispatch, error) {
	if client == nil || len(encryptionKey) != 32 {
		return RedisOTPDispatch{}, ErrPhoneOTPConfiguration
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil { return RedisOTPDispatch{}, err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return RedisOTPDispatch{}, err }
	return RedisOTPDispatch{client:client,aead:aead},nil
}

func dispatchKey(id uuid.UUID) string { return "higo:otp:dispatch:v1:"+id.String() }

var redisDispatchEnqueue = redis.NewScript(`
if redis.call('EXISTS',KEYS[1])==1 then return 0 end
redis.call('HSET',KEYS[1],'payload',ARGV[1],'state','ready','owner','')
redis.call('PEXPIRE',KEYS[1],ARGV[2])
return 1
`)
var redisDispatchClaim = redis.NewScript(`
if redis.call('EXISTS',KEYS[1])==0 then return {0,''} end
local state=redis.call('HGET',KEYS[1],'state')
if state=='done' then return {2,''} end
if state=='leased' and redis.call('PTTL',KEYS[2])>0 then return {3,''} end
redis.call('SET',KEYS[2],ARGV[1],'PX',ARGV[2])
redis.call('HSET',KEYS[1],'state','leased','owner',ARGV[1])
return {1,redis.call('HGET',KEYS[1],'payload')}
`)
var redisDispatchFinish = redis.NewScript(`
if redis.call('HGET',KEYS[1],'owner')~=ARGV[1] then return 0 end
if redis.call('GET',KEYS[2])~=ARGV[1] then return 0 end
redis.call('HSET',KEYS[1],'state',ARGV[2],'owner','')
redis.call('DEL',KEYS[2])
if ARGV[2]=='done' then redis.call('HDEL',KEYS[1],'payload') end
return 1
`)

func (s RedisOTPDispatch) seal(id uuid.UUID, phone, code string) (string,error) {
	nonce:=make([]byte,s.aead.NonceSize())
	if _,err:=rand.Read(nonce);err!=nil{return "",err}
	data:=[]byte(phone+"|"+code)
	ciphertext:=s.aead.Seal(nonce,nonce,data,[]byte(id.String()))
	return base64.RawURLEncoding.EncodeToString(ciphertext),nil
}

func (s RedisOTPDispatch) open(id uuid.UUID, encoded string)(string,string,error) {
	data,err:=base64.RawURLEncoding.DecodeString(encoded)
	if err!=nil{return "","",err}
	n:=s.aead.NonceSize()
	if len(data)<n{return "","",ErrOTPDispatchMissing}
	plain,err:=s.aead.Open(nil,data[:n],data[n:],[]byte(id.String()))
	if err!=nil{return "","",err}
	// Phone strings cannot contain '|'; code is always six digits.
	for i,b:=range plain { if b=='|' && i>0 && len(plain[i+1:])==6{return string(plain[:i]),string(plain[i+1:]),nil} }
	return "","",ErrOTPDispatchMissing
}

// Enqueue stores an encrypted SMS payload only until the challenge expires.
// Idempotent repeats never overwrite an earlier payload.
func(s RedisOTPDispatch)Enqueue(ctx context.Context,id uuid.UUID,phone,code string,ttl time.Duration)(bool,error){
	if id==uuid.Nil||ttl<=0||ttl>10*time.Minute {return false,ErrInvalidPhoneChallengeRequest}
	if _,err:=(PhoneChallengeRequest{PhoneNumber:phone,Purpose:PhoneChallengeSignInOrRegister}).Validate();err!=nil{return false,err}
	if len(code)!=6 {return false,ErrPhoneOTPCode}
	for _,b:=range code {if b<'0'||b>'9'{return false,ErrPhoneOTPCode}}
	sealed,err:=s.seal(id,phone,code)
	if err!=nil{return false,err}
	n,err:=redisDispatchEnqueue.Run(ctx,s.client,[]string{dispatchKey(id)},sealed,ttl.Milliseconds()).Int()
	return n==1,err
}

// Claim returns a lease token plus the decrypted payload. The worker must
// complete before the lease expires; leases are not auto-renewed. A delivery
// that runs longer than the lease requires provider-side deduplication.
func(s RedisOTPDispatch)Claim(ctx context.Context,id uuid.UUID,lease time.Duration)(string,string,string,error){
	if id==uuid.Nil||lease<time.Second||lease>time.Minute{return "","","",ErrInvalidPhoneChallengeRequest}
	owner:=uuid.NewString()
	result,err:=redisDispatchClaim.Run(ctx,s.client,[]string{dispatchKey(id),dispatchKey(id)+":lease"},owner,lease.Milliseconds()).Slice()
	if err!=nil{return "","","",err}
	status,err:=redisInteger(result[0]);if err!=nil{return "","","",err}
	if status==0||status==2{return "","","",ErrOTPDispatchMissing}
	if status==3{return "","","",ErrOTPDispatchBusy}
	phone,code,err:=s.open(id,fmt.Sprint(result[1]))
	if err!=nil{return "","","",err}
	return owner,phone,code,nil
}

// Finish only accepts the currently live lease owner. success=false makes
// the item eligible for a later retry; the challenge TTL is never extended.
func(s RedisOTPDispatch)Finish(ctx context.Context,id uuid.UUID,owner string,success bool)error{
	if id==uuid.Nil||owner==""{return ErrOTPDispatchOwnership}
	state:="ready";if success{state="done"}
	n,err:=redisDispatchFinish.Run(ctx,s.client,[]string{dispatchKey(id),dispatchKey(id)+":lease"},owner,state).Int()
	if err!=nil{return err}
	if n!=1{return ErrOTPDispatchOwnership}
	return nil
}
