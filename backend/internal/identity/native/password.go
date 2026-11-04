// Package native implements Go-owned credential and session primitives.
//
// This package is intentionally not wired into public authentication routes.
// Existing Kratos authentication remains authoritative until a reviewed
// migration, abuse controls, and end-to-end integration are complete.
package native

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

var (
	ErrPasswordPolicy     = errors.New("password does not meet requirements")
	ErrInvalidPasswordHash = errors.New("invalid stored password hash")
)

const (
	minPasswordRunes = 15
	maxPasswordBytes = 1024
	argonMemory      = uint32(64 * 1024) // KiB
	argonIterations  = uint32(3)
	argonParallelism = uint8(2)
	argonKeyBytes    = uint32(32)
	argonSaltBytes   = 16
)

// HashPassword produces an Argon2id PHC string using a new random salt.
// Passwords must have at least 15 Unicode characters and at most 1024 bytes.
// There is no arbitrary character-class rule or silently truncated input.
func HashPassword(password string) (string, error) {
	if !utf8.ValidString(password) || strings.ContainsRune(password, 0) ||
		len(password) > maxPasswordBytes || utf8.RuneCountInString(password) < minPasswordRunes {
		return "", ErrPasswordPolicy
	}
	salt := make([]byte, argonSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyBytes)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonIterations, argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword checks a stored Argon2id PHC value in constant time.
// The second result signals an authenticated credential that should be
// rehashed using the current parameters after a successful login.
//
// Stored parameters are bounded before allocating memory: corrupted or
// attacker-modified database content must not trigger unbounded Argon2 work.
func VerifyPassword(encoded, password string) (valid, needsRehash bool, err error) {
	if len(encoded) > 256 || len(password) > maxPasswordBytes || !utf8.ValidString(password) {
		return false, false, ErrInvalidPasswordHash
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, false, ErrInvalidPasswordHash
	}
	parameters := strings.Split(parts[3], ",")
	if len(parameters) != 3 {
		return false, false, ErrInvalidPasswordHash
	}
	memory, err := parseParameter(parameters[0], "m=")
	if err != nil {
		return false, false, ErrInvalidPasswordHash
	}
	iterations, err := parseParameter(parameters[1], "t=")
	if err != nil {
		return false, false, ErrInvalidPasswordHash
	}
	parallelism, err := parseParameter(parameters[2], "p=")
	if err != nil {
		return false, false, ErrInvalidPasswordHash
	}
	// Constrain all factors that influence CPU/memory allocation.
	if memory < 8*parallelism || memory > 256*1024 || iterations < 1 || iterations > 8 ||
		parallelism < 1 || parallelism > 8 {
		return false, false, ErrInvalidPasswordHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 32 {
		return false, false, ErrInvalidPasswordHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) < 16 || len(expected) > 64 {
		return false, false, ErrInvalidPasswordHash
	}
	calculated := argon2.IDKey([]byte(password), salt, uint32(iterations),
		uint32(memory), uint8(parallelism), uint32(len(expected)))
	ok := subtle.ConstantTimeCompare(calculated, expected) == 1
	return ok, ok && (uint32(memory) != argonMemory ||
		uint32(iterations) != argonIterations || uint8(parallelism) != argonParallelism ||
		len(expected) != int(argonKeyBytes)), nil
}

func parseParameter(part, prefix string) (uint64, error) {
	if !strings.HasPrefix(part, prefix) {
		return 0, ErrInvalidPasswordHash
	}
	value := strings.TrimPrefix(part, prefix)
	if value == "" {
		return 0, ErrInvalidPasswordHash
	}
	return strconv.ParseUint(value, 10, 32)
}
