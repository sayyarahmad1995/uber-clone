package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrPhoneVerificationProof = errors.New("phone verification proof invalid, expired, or consumed")

// PhoneVerificationProof is proof of phone possession for one purpose.
// It is NOT a Kratos session, a login token, or a user authorization.
type PhoneVerificationProof struct {
	PhoneNumber string
	Purpose     PhoneChallengePurpose
}

type RedisPhoneProofs struct {
	client *redis.Client
}

func NewRedisPhoneProofs(client *redis.Client) (RedisPhoneProofs, error) {
	if client == nil {
		return RedisPhoneProofs{}, ErrPhoneOTPConfiguration
	}
	return RedisPhoneProofs{client: client}, nil
}

func proofKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "higo:otp:proof:v1:" + hex.EncodeToString(sum[:])
}

// Verification and issuance of proof are performed in a single Redis script:
// an acknowledged correct code cannot be lost between two writes.
var redisVerifyAndIssueProof = redis.NewScript(`
if redis.call('EXISTS',KEYS[1])==0 then return {-1,''} end
if redis.call('HGET',KEYS[1],'state')~='pending' then return {-1,''} end
if redis.call('HGET',KEYS[1],'purpose')~=ARGV[2] then return {-1,''} end
local attempts=tonumber(redis.call('HGET',KEYS[1],'attempts'))
local maximum=tonumber(redis.call('HGET',KEYS[1],'max_attempts'))
if attempts>=maximum then return {-2,''} end
local phone=redis.call('HGET',KEYS[1],'phone')
if not phone then return {-1,''} end
redis.call('HINCRBY',KEYS[1],'attempts',1)
if redis.call('HGET',KEYS[1],'digest')~=ARGV[1] then return {-1,''} end
if redis.call('EXISTS',KEYS[2])~=0 then return {-1,''} end
redis.call('SET',KEYS[2],phone..'|'..ARGV[2],'PX',ARGV[3])
redis.call('HSET',KEYS[1],'state','consumed')
return {1,phone}
`)

// VerifyAndIssueProof validates the OTP, consumes it, and issues a 32-byte
// unpredictable opaque proof, all within one atomic Redis transaction.
// The proof expires in two minutes; never return it in logs or telemetry.
func (s RedisPhoneChallengeStore) VerifyAndIssueProof(ctx context.Context, id uuid.UUID, code string, purpose PhoneChallengePurpose, codec PhoneOTPCodec) (string, error) {
	if id == uuid.Nil || purpose != PhoneChallengeSignInOrRegister {
		return "", ErrPhoneVerificationProof
	}
	digest, err := codec.Digest(id, code)
	if err != nil {
		return "", ErrPhoneChallengeInvalid
	}
	lookup, err := s.client.Get(ctx, "higo:otp:v1:lookup:"+id.String()).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrPhoneChallengeInvalid
	}
	if err != nil {
		return "", err
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	token := hex.EncodeToString(random)
	result, err := redisVerifyAndIssueProof.Run(ctx, s.client, []string{lookup, proofKey(token)}, hex.EncodeToString(digest), string(purpose), (2 * time.Minute).Milliseconds()).Slice()
	if err != nil {
		return "", err
	}
	status, err := redisInteger(result[0])
	if err != nil {
		return "", err
	}
	if status == -2 {
		return "", ErrPhoneChallengeLimited
	}
	if status != 1 {
		return "", ErrPhoneChallengeInvalid
	}
	return token, nil
}

// ConsumeProof deletes the proof in the same Redis operation as reading it.
// Call only after validating the intended operation; never use the proof
// by itself to issue an authenticated session.
func (s RedisPhoneProofs) ConsumeProof(ctx context.Context, token string, purpose PhoneChallengePurpose) (PhoneVerificationProof, error) {
	if len(token) != 64 || purpose != PhoneChallengeSignInOrRegister {
		return PhoneVerificationProof{}, ErrPhoneVerificationProof
	}
	for i := 0; i < len(token); i++ {
		b := token[i]
		if (b < '0' || b > '9') && (b < 'a' || b > 'f') {
			return PhoneVerificationProof{}, ErrPhoneVerificationProof
		}
	}
	v, err := s.client.GetDel(ctx, proofKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return PhoneVerificationProof{}, ErrPhoneVerificationProof
	}
	if err != nil {
		return PhoneVerificationProof{}, err
	}
	// The proof value is server-generated and has a tightly defined shape.
	// A mismatch fails closed, although consumption is irreversible.
	parts := []byte(v)
	for i, b := range parts {
		if b == '|' {
			phone := string(parts[:i])
			actual := PhoneChallengePurpose(parts[i+1:])
			if actual != purpose {
				return PhoneVerificationProof{}, ErrPhoneVerificationProof
			}
			normalized, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: purpose}).Validate()
			if err != nil {
				return PhoneVerificationProof{}, ErrPhoneVerificationProof
			}
			return PhoneVerificationProof{PhoneNumber: normalized.PhoneNumber, Purpose: actual}, nil
		}
	}
	return PhoneVerificationProof{}, fmt.Errorf("%w: corrupted proof", ErrPhoneVerificationProof)
}
