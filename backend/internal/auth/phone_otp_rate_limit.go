package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrPhoneOTPRateLimited = errors.New("phone OTP request rate limited")

// OTPRateLimiter is a small, atomic fixed-window limiter. Its key material
// is HMAC-protected; pass a stable 32+ byte secret, not a per-process random key.
// Apply independent limits for phone and trusted requester source before SMS.
type OTPRateLimiter struct {
	client *redis.Client
	key    []byte
}

func NewOTPRateLimiter(client *redis.Client, secret []byte) (OTPRateLimiter, error) {
	if client == nil || len(secret) < 32 {
		return OTPRateLimiter{}, ErrPhoneOTPConfiguration
	}
	return OTPRateLimiter{client: client, key: append([]byte(nil), secret...)}, nil
}

var redisTakeOTPRateLimit = redis.NewScript(`
local n=redis.call('INCR',KEYS[1])
if n==1 then redis.call('PEXPIRE',KEYS[1],ARGV[1]) end
local ttl=redis.call('PTTL',KEYS[1])
if ttl<0 then redis.call('PEXPIRE',KEYS[1],ARGV[1]); ttl=tonumber(ARGV[1]) end
if n>tonumber(ARGV[2]) then return {0,ttl} end
return {1,ttl}
`)

// Allow charges an attempt even if rejected, preventing rapid probing.
// A failed Redis operation returns an error, so callers must fail closed.
// Do not derive source from an untrusted X-Forwarded-For header.
func (l OTPRateLimiter) Allow(ctx context.Context, dimension, subject string, max int64, window time.Duration) (time.Duration, error) {
	if dimension == "" || subject == "" || max <= 0 || window < time.Second || window > 24*time.Hour {
		return 0, ErrInvalidPhoneChallengeRequest
	}
	mac := hmac.New(sha256.New, l.key)
	_, _ = mac.Write([]byte("higo:otp-limit:v1:" + dimension + ":" + subject))
	key := "higo:otp:limit:v1:" + hex.EncodeToString(mac.Sum(nil))
	result, err := redisTakeOTPRateLimit.Run(ctx, l.client, []string{key}, window.Milliseconds(), max).Slice()
	if err != nil {
		return 0, fmt.Errorf("OTP rate limiter unavailable: %w", err)
	}
	allowed, err := redisInteger(result[0])
	if err != nil {
		return 0, err
	}
	ms, err := redisInteger(result[1])
	if err != nil {
		return 0, err
	}
	if allowed == 0 {
		return time.Duration(ms) * time.Millisecond, ErrPhoneOTPRateLimited
	}
	return 0, nil
}
