package native

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSession = errors.New("invalid or expired session")
	ErrSessionStore   = errors.New("session store is required")
)

const (
	accessLifetime  = 15 * time.Minute
	refreshLifetime = 30 * 24 * time.Hour
)

// SessionRecord contains only SHA-256 digests of bearer secrets. It must not
// include the plaintext access or refresh tokens in persistent storage.
type SessionRecord struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	AccessDigest     [32]byte
	RefreshDigest    [32]byte
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// SessionStore is the atomic persistence boundary for server-revocable,
// rotating native sessions. A refresh token may rotate only once.
type SessionStore interface {
	InsertSession(context.Context, SessionRecord) error
	ResolveAccess(context.Context, uuid.UUID, [32]byte, time.Time) (uuid.UUID, error)
	RotateRefresh(context.Context, uuid.UUID, [32]byte, [32]byte, [32]byte, time.Time, time.Time) (time.Time, time.Time, error)
	RevokeAccess(context.Context, uuid.UUID, [32]byte, time.Time) error
	RevokeUserSessions(context.Context, uuid.UUID, time.Time) error
}

// SessionTokens are bearer credentials. Only return them over TLS to an
// authorized caller; the refresh token is never logged or sent to business APIs.
type SessionTokens struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// SessionManager issues opaque random bearer secrets. A session ID prefix
// permits indexed lookup but is not itself an authentication credential.
// Database queries always also require the digest of the full secret.
type SessionManager struct {
	store SessionStore
	now   func() time.Time
}

func NewSessionManager(store SessionStore) (SessionManager, error) {
	if store == nil {
		return SessionManager{}, ErrSessionStore
	}
	return SessionManager{store: store, now: time.Now}, nil
}

func (m SessionManager) Issue(ctx context.Context, userID uuid.UUID) (SessionTokens, error) {
	if userID == uuid.Nil {
		return SessionTokens{}, ErrInvalidSession
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return SessionTokens{}, err
	}
	access, accessDigest, err := randomBearer(id)
	if err != nil {
		return SessionTokens{}, err
	}
	refresh, refreshDigest, err := randomBearer(id)
	if err != nil {
		return SessionTokens{}, err
	}
	now := m.now().UTC()
	tokens := SessionTokens{
		AccessToken: access, RefreshToken: refresh,
		AccessExpiresAt: now.Add(accessLifetime),
		RefreshExpiresAt: now.Add(refreshLifetime),
	}
	if err := m.store.InsertSession(ctx, SessionRecord{
		ID: id, UserID: userID,
		AccessDigest: accessDigest, RefreshDigest: refreshDigest,
		AccessExpiresAt: tokens.AccessExpiresAt, RefreshExpiresAt: tokens.RefreshExpiresAt,
	}); err != nil {
		return SessionTokens{}, err
	}
	return tokens, nil
}

func (m SessionManager) Authenticate(ctx context.Context, accessToken string) (uuid.UUID, error) {
	id, digest, err := digestBearer(accessToken)
	if err != nil {
		return uuid.Nil, ErrInvalidSession
	}
	return m.store.ResolveAccess(ctx, id, digest, m.now().UTC())
}

// Refresh atomically consumes the supplied refresh secret. Concurrent calls
// with the same token cannot both succeed; only the winning call receives the
// next credentials. Replay is rejected, not silently retried.
func (m SessionManager) Refresh(ctx context.Context, refreshToken string) (SessionTokens, error) {
	id, oldDigest, err := digestBearer(refreshToken)
	if err != nil {
		return SessionTokens{}, ErrInvalidSession
	}
	access, accessDigest, err := randomBearer(id)
	if err != nil {
		return SessionTokens{}, err
	}
	refresh, refreshDigest, err := randomBearer(id)
	if err != nil {
		return SessionTokens{}, err
	}
	accessExpiry, refreshExpiry, err := m.store.RotateRefresh(
		ctx, id, oldDigest, accessDigest, refreshDigest,
		m.now().UTC().Add(accessLifetime), m.now().UTC())
	if err != nil {
		return SessionTokens{}, err
	}
	return SessionTokens{
		AccessToken: access, RefreshToken: refresh,
		AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry,
	}, nil
}

func (m SessionManager) Logout(ctx context.Context, accessToken string) error {
	id, digest, err := digestBearer(accessToken)
	if err != nil {
		return ErrInvalidSession
	}
	return m.store.RevokeAccess(ctx, id, digest, m.now().UTC())
}

// RevokeUser is intended for authenticated credential resets, account
// suspension and security incidents; untrusted callers must not invoke it.
func (m SessionManager) RevokeUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidSession
	}
	return m.store.RevokeUserSessions(ctx, userID, m.now().UTC())
}

func randomBearer(id uuid.UUID) (string, [32]byte, error) {
	var nonce [32]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", [32]byte{}, err
	}
	token := id.String() + "." + base64.RawURLEncoding.EncodeToString(nonce[:])
	return token, sha256.Sum256([]byte(token)), nil
}

func digestBearer(token string) (uuid.UUID, [32]byte, error) {
	if len(token) != 80 {
		return uuid.Nil, [32]byte{}, ErrInvalidSession
	}
	prefix, secret, ok := strings.Cut(token, ".")
	if !ok {
		return uuid.Nil, [32]byte{}, ErrInvalidSession
	}
	id, err := uuid.Parse(prefix)
	if err != nil || id == uuid.Nil || id.String() != prefix {
		return uuid.Nil, [32]byte{}, ErrInvalidSession
	}
	decoded, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != secret {
		return uuid.Nil, [32]byte{}, ErrInvalidSession
	}
	return id, sha256.Sum256([]byte(token)), nil
}
