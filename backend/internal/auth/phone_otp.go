package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

var ErrPhoneOTPConfiguration = errors.New("invalid phone OTP configuration")
var ErrPhoneOTPCode = errors.New("invalid phone OTP code")

const phoneOTPDomain = "higo:phone-otp:v1:"

// PhoneOTPCodec generates OTPs and computes challenge-bound HMAC digests.
// The secret must be a stable, random, server-side key with at least 32 bytes.
// It must not be stored alongside challenges or sent to clients.
type PhoneOTPCodec struct {
	key []byte
}

func NewPhoneOTPCodec(secret []byte) (PhoneOTPCodec, error) {
	if len(secret) < 32 {
		return PhoneOTPCodec{}, ErrPhoneOTPConfiguration
	}
	key := make([]byte, len(secret))
	copy(key, secret)
	return PhoneOTPCodec{key: key}, nil
}

// Generate returns a six-digit code (including leading zeros) and its digest.
// Callers must deliver the code only through an approved out-of-band channel
// and persist only the digest, never the plaintext code.
func (c PhoneOTPCodec) Generate(id uuid.UUID) (string, []byte, error) {
	if id == uuid.Nil || len(c.key) < 32 {
		return "", nil, ErrPhoneOTPConfiguration
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", nil, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	digest, err := c.Digest(id, code)
	if err != nil {
		return "", nil, err
	}
	return code, digest, nil
}

// Digest binds the OTP to the individual challenge to prevent substituting
// a known code/digest from another challenge. A digest is not a bearer token.
func (c PhoneOTPCodec) Digest(id uuid.UUID, code string) ([]byte, error) {
	if id == uuid.Nil || len(c.key) < 32 {
		return nil, ErrPhoneOTPConfiguration
	}
	if len(code) != 6 {
		return nil, ErrPhoneOTPCode
	}
	for i := 0; i < len(code); i++ {
		if code[i] < '0' || code[i] > '9' {
			return nil, ErrPhoneOTPCode
		}
	}
	mac := hmac.New(sha256.New, c.key)
	_, _ = mac.Write([]byte(phoneOTPDomain))
	_, _ = mac.Write([]byte(id.String()))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(code))
	return mac.Sum(nil), nil
}
