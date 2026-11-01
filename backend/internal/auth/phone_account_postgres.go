package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// PostgresPhoneAccountResolver reads durable phone ownership. The Kratos
// identity source string must be supplied from trusted backend configuration.
type PostgresPhoneAccountResolver struct {
	db     *sql.DB
	issuer string
}

func NewPostgresPhoneAccountResolver(db *sql.DB, issuer string) (PostgresPhoneAccountResolver, error) {
	if db == nil || issuer == "" {
		return PostgresPhoneAccountResolver{}, ErrPhoneOTPConfiguration
	}
	return PostgresPhoneAccountResolver{db: db, issuer: issuer}, nil
}

func (r PostgresPhoneAccountResolver) ResolveVerifiedPhone(ctx context.Context, phone string) (PhoneAccountBinding, error) {
	normalized, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate()
	if err != nil {
		return PhoneAccountBinding{}, ErrPhoneVerificationProof
	}
	var ownerID uuid.UUID
	var subject, storedPhone string
	err = r.db.QueryRowContext(ctx, `
		SELECT p.user_id, p.subject, p.phone_number
		FROM user_verified_phones p
		JOIN external_identities e ON e.user_id=p.user_id
			AND e.issuer=p.issuer AND e.subject=p.subject
		WHERE p.phone_number=$1 AND p.issuer=$2
	`, normalized.PhoneNumber, r.issuer).Scan(&ownerID, &subject, &storedPhone)
	if errors.Is(err, sql.ErrNoRows) {
		return PhoneAccountBinding{}, nil
	}
	if err != nil {
		return PhoneAccountBinding{}, err
	}
	if ownerID == uuid.Nil || subject == "" || storedPhone != normalized.PhoneNumber {
		return PhoneAccountBinding{}, ErrPhoneIdentityBinding
	}
	return PhoneAccountBinding{Exists: true, IdentitySubject: subject, VerifiedPhone: storedPhone}, nil
}

// BindVerifiedPhone is exclusively for the post-registration account-linking
// transaction. Only call it with a verified phone and an independently checked
// Kratos identity. No public HTTP handler must expose it directly.
// Conflicting ownership returns a SQL unique-constraint error; it is not
// acceptable to overwrite an existing phone link.
func (r PostgresPhoneAccountResolver) BindVerifiedPhone(ctx context.Context, userID uuid.UUID, subject, phone string) error {
	normalized, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate()
	if err != nil || userID == uuid.Nil || subject == "" {
		return ErrPhoneIdentityBinding
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_verified_phones (phone_number,user_id,issuer,subject)
		SELECT $1,e.user_id,e.issuer,e.subject
		FROM external_identities e
		WHERE e.user_id=$2 AND e.issuer=$3 AND e.subject=$4
		ON CONFLICT (phone_number) DO NOTHING
	`, normalized.PhoneNumber, userID, r.issuer, subject)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		return nil
	}
	// Idempotent only for an identical binding.
	var existingUser uuid.UUID
	var existingSubject string
	err = r.db.QueryRowContext(ctx, `
		SELECT user_id,subject FROM user_verified_phones WHERE phone_number=$1 AND issuer=$2
	`, normalized.PhoneNumber, r.issuer).Scan(&existingUser, &existingSubject)
	if err != nil {
		return ErrPhoneIdentityBinding
	}
	if existingUser != userID || existingSubject != subject {
		return ErrPhoneIdentityBinding
	}
	return nil
}

// MatchesVerifiedPhone authorizes an already authenticated Kratos subject.
// Unlike a phone-only existence check, this requires an exact three-way match.
func (r PostgresPhoneAccountResolver) MatchesVerifiedPhone(ctx context.Context, issuer, subject, phone string) (bool, error) {
	if issuer != r.issuer || subject == "" {
		return false, nil
	}
	normalized, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate()
	if err != nil { return false, nil }
	var exists bool
	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_verified_phones p
			JOIN external_identities e ON e.user_id=p.user_id
				AND e.issuer=p.issuer AND e.subject=p.subject
			WHERE p.phone_number=$1 AND p.issuer=$2 AND p.subject=$3
		)
	`, normalized.PhoneNumber, issuer, subject).Scan(&exists)
	return exists, err
}
