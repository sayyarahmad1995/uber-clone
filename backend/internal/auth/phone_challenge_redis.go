package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisPhoneChallengeStore stores only ephemeral phone-verification state.
// No public route calls this until SMS delivery and abuse limits are ready.
type RedisPhoneChallengeStore struct {
	client *redis.Client
	key    []byte
}

func NewRedisPhoneChallengeStore(client *redis.Client, namespaceSecret []byte) (RedisPhoneChallengeStore, error) {
	if client == nil || len(namespaceSecret) < 32 {
		return RedisPhoneChallengeStore{}, ErrPhoneOTPConfiguration
	}
	key := append([]byte(nil), namespaceSecret...)
	return RedisPhoneChallengeStore{client: client, key: key}, nil
}

// Keys are keyed HMACs rather than plaintext phone numbers or unkeyed hashes.
func (s RedisPhoneChallengeStore) scope(phone string, purpose PhoneChallengePurpose) string {
	h := hmac.New(sha256.New, s.key)
	_, _ = h.Write([]byte("higo:otp-scope:v1:" + phone + ":" + string(purpose)))
	return "higo:otp:v1:" + hex.EncodeToString(h.Sum(nil))
}

// Each script runs atomically on a single Redis server. All four issuance keys
// are written together; a retry with the same idempotency key never reissues.
var redisIssueOTP = redis.NewScript(`
local prior=redis.call('GET',KEYS[3])
if prior then return {0,prior,redis.call('PTTL',KEYS[2])} end
local remaining=redis.call('PTTL',KEYS[2])
if remaining>0 then return {-1,'',remaining} end
redis.call('HSET',KEYS[1],'phone',ARGV[1],'digest',ARGV[2],'attempts',0,'max_attempts',5,'state','pending')
redis.call('PEXPIRE',KEYS[1],ARGV[3])
redis.call('SET',KEYS[2],'1','PX',ARGV[4])
redis.call('SET',KEYS[3],ARGV[5],'PX',ARGV[3])
redis.call('SET',KEYS[4],KEYS[1],'PX',ARGV[3])
return {1,ARGV[5],tonumber(ARGV[4])}
`)

var redisVerifyOTP = redis.NewScript(`
if redis.call('EXISTS',KEYS[1])==0 then return {-1,''} end
if redis.call('HGET',KEYS[1],'state')~='pending' then return {-1,''} end
local attempts=tonumber(redis.call('HGET',KEYS[1],'attempts'))
local maxAttempts=tonumber(redis.call('HGET',KEYS[1],'max_attempts'))
if attempts>=maxAttempts then return {-2,''} end
redis.call('HINCRBY',KEYS[1],'attempts',1)
if redis.call('HGET',KEYS[1],'digest')==ARGV[1] then
 redis.call('HSET',KEYS[1],'state','consumed')
 return {1,redis.call('HGET',KEYS[1],'phone')}
end
return {0,''}
`)

func redisInteger(v interface{}) (int64, error) {
	n, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected Redis script result %T", v)
	}
	return n, nil
}

// Create enforces the per-phone resend cooldown and request idempotency.
// ShouldSend does not imply delivery: an SMS outbox/delivery policy is required.
func (s RedisPhoneChallengeStore) Create(ctx context.Context, c NewPhoneChallenge) (PhoneChallengeCreation, error) {
	normalized, err := (PhoneChallengeRequest{PhoneNumber: c.PhoneNumber, Purpose: c.Purpose}).Validate()
	now := time.Now()
	if err != nil || c.ID == uuid.Nil || len(c.CodeDigest) != 32 ||
		len(c.IdempotencyKey) < 16 || len(c.IdempotencyKey) > 128 ||
		!c.ExpiresAt.After(now) || !c.ResendAfter.After(now) || c.ResendAfter.After(c.ExpiresAt) {
		return PhoneChallengeCreation{}, ErrInvalidPhoneChallengeRequest
	}
	ttl, cooldown := time.Until(c.ExpiresAt).Milliseconds(), time.Until(c.ResendAfter).Milliseconds()
	if ttl <= 0 || cooldown <= 0 {
		return PhoneChallengeCreation{}, ErrInvalidPhoneChallengeRequest
	}
	scope := s.scope(normalized.PhoneNumber, c.Purpose)
	lookup := "higo:otp:v1:lookup:" + c.ID.String()
	// An idempotency key can contain user-controlled text, so hash it.
	idemHash := sha256.Sum256([]byte(c.IdempotencyKey))
	keys := []string{scope + ":challenge:" + c.ID.String(), scope + ":cooldown", scope + ":idem:" + hex.EncodeToString(idemHash[:]), lookup}
	output, err := redisIssueOTP.Run(ctx, s.client, keys, normalized.PhoneNumber, hex.EncodeToString(c.CodeDigest), ttl, cooldown, c.ID.String()).Slice()
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	status, err := redisInteger(output[0])
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	if status == -1 {
		remaining, err := redisInteger(output[2])
		if err != nil {
			return PhoneChallengeCreation{}, err
		}
		return PhoneChallengeCreation{ResendAfter: time.Now().Add(time.Duration(remaining) * time.Millisecond)}, ErrPhoneChallengeCooldown
	}
	id, err := uuid.Parse(fmt.Sprint(output[1]))
	if err != nil {
		return PhoneChallengeCreation{}, err
	}
	return PhoneChallengeCreation{ID: id, ShouldSend: status == 1, ResendAfter: c.ResendAfter}, nil
}

// Consume atomically counts attempts, verifies the keyed digest and makes a
// challenge single-use. The lookup is an additional Redis request; no database
// access is involved. A valid OTP yields only a phone proof, not a session.
func (s RedisPhoneChallengeStore) Consume(ctx context.Context, id uuid.UUID, candidateDigest []byte) (string, error) {
	if id == uuid.Nil || len(candidateDigest) != 32 {
		return "", ErrPhoneChallengeInvalid
	}
	key, err := s.client.Get(ctx, "higo:otp:v1:lookup:" + id.String()).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrPhoneChallengeInvalid
	}
	if err != nil {
		return "", err
	}
	output, err := redisVerifyOTP.Run(ctx, s.client, []string{key}, hex.EncodeToString(candidateDigest)).Slice()
	if err != nil {
		return "", err
	}
	status, err := redisInteger(output[0])
	if err != nil {
		return "", err
	}
	switch status {
	case 1:
		return fmt.Sprint(output[1]), nil
	case -2:
		return "", ErrPhoneChallengeLimited
	default:
		return "", ErrPhoneChallengeInvalid
	}
}
