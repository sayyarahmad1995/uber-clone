package auth

import (
	"bytes"
	"errors"
	"regexp"
	"testing"

	"github.com/google/uuid"
)

func TestPhoneOTPCodecGenerateAndDigest(t *testing.T) {
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{0x4a}, 32))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	shape := regexp.MustCompile(`^[0-9]{6}$`)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, digest, err := codec.Generate(id)
		if err != nil {
			t.Fatal(err)
		}
		if !shape.MatchString(code) || len(digest) != sha256DigestSize {
			t.Fatalf("invalid code or digest: code length=%d digest length=%d", len(code), len(digest))
		}
		computed, err := codec.Digest(id, code)
		if err != nil || !bytes.Equal(digest, computed) {
			t.Fatal("generated digest does not match")
		}
		seen[code] = true
	}
	if len(seen) < 2 {
		t.Fatal("generated OTPs were unexpectedly identical")
	}
}

const sha256DigestSize = 32

func TestPhoneOTPCodecDigestIsChallengeAndKeyBound(t *testing.T) {
	id := uuid.New()
	otherID := uuid.New()
	a, _ := NewPhoneOTPCodec(bytes.Repeat([]byte{1}, 32))
	b, _ := NewPhoneOTPCodec(bytes.Repeat([]byte{2}, 32))
	first, _ := a.Digest(id, "000001")
	otherChallenge, _ := a.Digest(otherID, "000001")
	otherSecret, _ := b.Digest(id, "000001")
	otherCode, _ := a.Digest(id, "000002")
	for _, value := range [][]byte{otherChallenge, otherSecret, otherCode} {
		if bytes.Equal(first, value) {
			t.Fatal("digest was not bound to challenge, secret, and code")
		}
	}
}

func TestPhoneOTPCodecRejectsInvalidInputs(t *testing.T) {
	if _, err := NewPhoneOTPCodec(bytes.Repeat([]byte{1}, 31)); !errors.Is(err, ErrPhoneOTPConfiguration) {
		t.Fatalf("short key: %v", err)
	}
	codec, _ := NewPhoneOTPCodec(bytes.Repeat([]byte{1}, 32))
	if _, _, err := codec.Generate(uuid.Nil); !errors.Is(err, ErrPhoneOTPConfiguration) {
		t.Fatalf("nil challenge ID: %v", err)
	}
	for _, code := range []string{"12345", "1234567", "abcdef", "12345x", "۱۲۳۴۵۶", "12 456"} {
		if _, err := codec.Digest(uuid.New(), code); !errors.Is(err, ErrPhoneOTPCode) {
			t.Fatalf("code %q returned %v", code, err)
		}
	}
	if _, err := codec.Digest(uuid.Nil, "123456"); !errors.Is(err, ErrPhoneOTPConfiguration) {
		t.Fatalf("nil challenge ID: %v", err)
	}
}
