package native

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrCredentialsNotFound = errors.New("native password credentials not found")

// PostgresPasswords owns the inactive Go-native credential rows. These methods
// are internal persistence primitives, not authorization or public endpoints.
// Call SetHash only after independently authorized registration or credential
// change; the database layer does not verify an OTP or a current password.
type PostgresPasswords struct {
	db *sql.DB
}

func NewPostgresPasswords(db *sql.DB) (PostgresPasswords, error) {
	if db == nil {
		return PostgresPasswords{}, ErrSessionStore
	}
	return PostgresPasswords{db: db}, nil
}

func (p PostgresPasswords) ReadHash(ctx context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", ErrCredentialsNotFound
	}
	var encoded string
	err := p.db.QueryRowContext(ctx,
		`SELECT encoded_hash FROM native_password_credentials WHERE user_id=$1`, userID,
	).Scan(&encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrCredentialsNotFound
	}
	return encoded, err
}

func (p PostgresPasswords) SetHash(ctx context.Context, userID uuid.UUID, encoded string) error {
	if userID == uuid.Nil || !strings.HasPrefix(encoded, "$argon2id$v=19$") ||
		len(encoded) < 50 || len(encoded) > 256 {
		return ErrInvalidPasswordHash
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO native_password_credentials (user_id, encoded_hash)
		VALUES ($1,$2)
		ON CONFLICT (user_id) DO UPDATE SET
			encoded_hash=EXCLUDED.encoded_hash,
			updated_at=clock_timestamp()
	`, userID, encoded)
	return err
}

// PostgresSessions implements atomic session creation, digest validation,
// one-time refresh rotation, and server-side revocation.
type PostgresSessions struct {
	db *sql.DB
}

func NewPostgresSessions(db *sql.DB) (PostgresSessions, error) {
	if db == nil {
		return PostgresSessions{}, ErrSessionStore
	}
	return PostgresSessions{db: db}, nil
}

func (p PostgresSessions) InsertSession(ctx context.Context, s SessionRecord) error {
	if s.ID == uuid.Nil || s.UserID == uuid.Nil || s.AccessExpiresAt.IsZero() ||
		!s.AccessExpiresAt.Before(s.RefreshExpiresAt) {
		return ErrInvalidSession
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO native_auth_sessions (
			id,user_id,access_digest,refresh_digest,access_expires_at,refresh_expires_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`, s.ID, s.UserID, s.AccessDigest[:], s.RefreshDigest[:],
		s.AccessExpiresAt, s.RefreshExpiresAt)
	return err
}

func (p PostgresSessions) ResolveAccess(ctx context.Context, id uuid.UUID, digest [32]byte, now time.Time) (uuid.UUID, error) {
	var userID uuid.UUID
	err := p.db.QueryRowContext(ctx, `
		SELECT user_id FROM native_auth_sessions
		WHERE id=$1 AND access_digest=$2 AND revoked_at IS NULL
		  AND access_expires_at>$3 AND refresh_expires_at>$3
	`, id, digest[:], now).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrInvalidSession
	}
	return userID, err
}

func (p PostgresSessions) RotateRefresh(
	ctx context.Context, id uuid.UUID, oldDigest, newAccessDigest, newRefreshDigest [32]byte,
	accessUntil, now time.Time,
) (time.Time, time.Time, error) {
	var accessExpires, refreshExpires time.Time
	// Atomic UPDATE is the compare-and-swap: only one concurrent refresh
	// request can match the old digest. The 30-day absolute expiry does not
	// slide indefinitely, and the access token cannot outlive that expiry.
	err := p.db.QueryRowContext(ctx, `
		UPDATE native_auth_sessions SET
			access_digest=$3,
			refresh_digest=$4,
			access_expires_at=LEAST($5,refresh_expires_at),
			rotated_at=$6
		WHERE id=$1 AND refresh_digest=$2 AND revoked_at IS NULL
		  AND refresh_expires_at>$6
		RETURNING access_expires_at,refresh_expires_at
	`, id, oldDigest[:], newAccessDigest[:], newRefreshDigest[:],
		accessUntil, now).Scan(&accessExpires, &refreshExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, time.Time{}, ErrInvalidSession
	}
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return accessExpires, refreshExpires, nil
}

func (p PostgresSessions) RevokeAccess(ctx context.Context, id uuid.UUID, digest [32]byte, now time.Time) error {
	res, err := p.db.ExecContext(ctx, `
		UPDATE native_auth_sessions SET revoked_at=$3
		WHERE id=$1 AND access_digest=$2 AND revoked_at IS NULL
	`, id, digest[:], now)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrInvalidSession
	}
	return nil
}

func (p PostgresSessions) RevokeUserSessions(ctx context.Context, userID uuid.UUID, now time.Time) error {
	if userID == uuid.Nil {
		return ErrInvalidSession
	}
	_, err := p.db.ExecContext(ctx, `
		UPDATE native_auth_sessions SET revoked_at=$2
		WHERE user_id=$1 AND revoked_at IS NULL
	`, userID, now)
	return err
}
