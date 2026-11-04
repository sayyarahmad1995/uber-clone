package native

import (
	"errors"
	"strings"
	"testing"
)

func TestPasswordHashRoundTripAndUniqueSalts(t *testing.T) {
	const password = "a sufficiently long password!"
	first, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "$argon2id$v=19$") {
		t.Fatal("hashes must use Argon2id and fresh salts")
	}
	valid, rehash, err := VerifyPassword(first, password)
	if err != nil || !valid || rehash {
		t.Fatalf("correct password: valid=%v rehash=%v err=%v", valid, rehash, err)
	}
	valid, rehash, err = VerifyPassword(first, "incorrect password")
	if err != nil || valid || rehash {
		t.Fatalf("wrong password: valid=%v rehash=%v err=%v", valid, rehash, err)
	}
}

func TestPasswordPolicy(t *testing.T) {
	for _, password := range []string{"short", strings.Repeat("a", 1025), "invalid\x00character", string([]byte{0xff})} {
		if _, err := HashPassword(password); !errors.Is(err, ErrPasswordPolicy) {
			t.Fatalf("expected password policy rejection, got %v", err)
		}
	}
	if _, err := HashPassword(strings.Repeat("界", 15)); err != nil {
		t.Fatalf("valid Unicode password rejected: %v", err)
	}
}

func TestPasswordStoredHashLimits(t *testing.T) {
	for _, stored := range []string{
		"",
		"plaintext",
		"$argon2id$v=18$m=65536,t=3,p=2$AA$AA",
		"$argon2id$v=19$m=4294967295,t=3,p=2$AA$AA",
		"$argon2id$v=19$m=65536,t=9999,p=2$AA$AA",
		"$argon2id$v=19$m=65536,t=3,p=99$AA$AA",
		"$argon2id$v=19$m=65536,t=3,p=2$invalid*$AA",
		strings.Repeat("x", 300),
	} {
		if valid, _, err := VerifyPassword(stored, "a sufficiently long password!"); valid || !errors.Is(err, ErrInvalidPasswordHash) {
			t.Fatalf("malformed hash was accepted or used: %q err=%v", stored, err)
		}
	}
}
