package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// CreatePhoneAccount atomically establishes a HiGO user, Kratos identity link,
// default rider capability, and verified phone binding. The caller MUST first
// validate phone possession and the Kratos identity; this function does not
// create a Kratos identity or consume a Redis verification proof.
//
// A retry with the same identity and phone is idempotent. Any conflicting
// phone or identity ownership fails closed and rolls back all writes.
func (r PostgresPhoneAccountResolver) CreatePhoneAccount(ctx context.Context, subject, phone string) (uuid.UUID, error) {
	normalized, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate()
	if err != nil || subject == "" {
		return uuid.Nil, ErrPhoneIdentityBinding
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return uuid.Nil, err }
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT user_id FROM external_identities WHERE issuer=$1 AND subject=$2
	`, r.issuer, subject).Scan(&userID)
	switch {
	case err == nil:
		// Do not attach a new phone to an existing identity unless this exact
		// verified-phone binding already exists.
		var boundPhone string
		err = tx.QueryRowContext(ctx, `
			SELECT phone_number FROM user_verified_phones
			WHERE user_id=$1 AND issuer=$2 AND subject=$3
		`, userID, r.issuer, subject).Scan(&boundPhone)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && boundPhone != normalized.PhoneNumber) {
			return uuid.Nil, ErrPhoneIdentityBinding
		}
		if err != nil { return uuid.Nil, err }
	case errors.Is(err, sql.ErrNoRows):
		// Prevent silently binding a phone already owned by another identity.
		var exists bool
		if err = tx.QueryRowContext(ctx,`
			SELECT EXISTS(SELECT 1 FROM user_verified_phones WHERE phone_number=$1)
		`, normalized.PhoneNumber).Scan(&exists); err != nil { return uuid.Nil, err }
		if exists { return uuid.Nil, ErrPhoneIdentityBinding }
		userID = uuid.New()
		if _, err = tx.ExecContext(ctx, `INSERT INTO users(id) VALUES ($1)`,userID);err != nil{return uuid.Nil,err}
		if _, err = tx.ExecContext(ctx,`
			INSERT INTO external_identities(user_id,issuer,subject) VALUES ($1,$2,$3)
		`,userID,r.issuer,subject);err != nil{return uuid.Nil,err}
		if _, err = tx.ExecContext(ctx,`
			INSERT INTO user_verified_phones(phone_number,user_id,issuer,subject) VALUES ($1,$2,$3,$4)
		`,normalized.PhoneNumber,userID,r.issuer,subject);err != nil{return uuid.Nil,err}
	default:
		return uuid.Nil, err
	}
	if _, err = tx.ExecContext(ctx,`
		INSERT INTO user_capabilities(user_id,capability) VALUES ($1,'rider') ON CONFLICT DO NOTHING
	`,userID);err != nil{return uuid.Nil,err}
	if err = tx.Commit();err != nil{return uuid.Nil,err}
	return userID,nil
}
