package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// RegistrationOperation is durable coordination metadata, never a credential.
// A reserved operation is not permission to retry an uncertain Kratos request.
type RegistrationOperation struct {
	ID        uuid.UUID
	Phone     string
	Issuer    string
	State     string
	Subject   string
	UserID    uuid.UUID
	UpdatedAt time.Time
}

// ReserveRegistration is only for the trusted, proof-authorized registration
// coordinator. Replays for the same phone return the existing operation;
// they must not trigger another Kratos identity creation automatically.
func (r PostgresPhoneAccountResolver) ReserveRegistration(ctx context.Context, phone string) (RegistrationOperation, error) {
	n, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate()
	if err != nil {
		return RegistrationOperation{}, ErrPhoneIdentityBinding
	}
	id := uuid.New()
	_, err = r.db.ExecContext(ctx, `
 INSERT INTO phone_registration_operations(id,phone_number,issuer)
 VALUES($1,$2,$3) ON CONFLICT(phone_number,issuer) DO NOTHING
 `, id, n.PhoneNumber, r.issuer)
	if err != nil {
		return RegistrationOperation{}, err
	}
	return r.GetRegistration(ctx, n.PhoneNumber)
}

func (r PostgresPhoneAccountResolver) GetRegistration(ctx context.Context, phone string) (RegistrationOperation, error) {
	var op RegistrationOperation
	var subject sql.NullString
	var userID uuid.NullUUID
	err := r.db.QueryRowContext(ctx, `
 SELECT id,phone_number,issuer,state,identity_subject,user_id,updated_at
 FROM phone_registration_operations WHERE phone_number=$1 AND issuer=$2
 `, phone, r.issuer).Scan(&op.ID, &op.Phone, &op.Issuer, &op.State, &subject, &userID, &op.UpdatedAt)
	if err != nil {
		return RegistrationOperation{}, err
	}
	if subject.Valid {
		op.Subject = subject.String
	}
	if userID.Valid {
		op.UserID = userID.UUID
	}
	return op, nil
}

// RecordRegistrationIdentity persists the Kratos identity after its confirmed
// creation. An unknown Kratos outcome requires manual reconciliation; never
// create a second identity merely because the operation remains reserved.
func (r PostgresPhoneAccountResolver) RecordRegistrationIdentity(ctx context.Context, id uuid.UUID, subject string) (RegistrationOperation, error) {
	if id == uuid.Nil || subject == "" {
		return RegistrationOperation{}, ErrPhoneIdentityBinding
	}
	result, err := r.db.ExecContext(ctx, `
 UPDATE phone_registration_operations SET state='identity_created',
 identity_subject=$2,updated_at=clock_timestamp()
 WHERE id=$1 AND state='submitting'
 `, id, subject)
	if err != nil {
		return RegistrationOperation{}, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return RegistrationOperation{}, err
	}
	var op RegistrationOperation
	var nullable sql.NullString
	var uid uuid.NullUUID
	err = r.db.QueryRowContext(ctx, `
 SELECT id,phone_number,issuer,state,identity_subject,user_id,updated_at
 FROM phone_registration_operations WHERE id=$1
 `, id).Scan(&op.ID, &op.Phone, &op.Issuer, &op.State, &nullable, &uid, &op.UpdatedAt)
	if err != nil {
		return RegistrationOperation{}, err
	}
	if nullable.Valid {
		op.Subject = nullable.String
	}
	if uid.Valid {
		op.UserID = uid.UUID
	}
	if n != 1 && (op.Subject != subject || op.State != "identity_created" && op.State != "completed") {
		return RegistrationOperation{}, ErrPhoneIdentityBinding
	}
	return op, nil
}

// CompleteRegistration writes HiGO identity ownership, rider capability,
// phone binding and operation completion in a single PostgreSQL transaction.
// The Kratos identity must have been independently checked by the coordinator.
func (r PostgresPhoneAccountResolver) CompleteRegistration(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()
	var phone, state string
	var subject sql.NullString
	var existing uuid.NullUUID
	err = tx.QueryRowContext(ctx, `
 SELECT phone_number,identity_subject,state,user_id
 FROM phone_registration_operations WHERE id=$1 FOR UPDATE
 `, id).Scan(&phone, &subject, &state, &existing)
	if err != nil {
		return uuid.Nil, err
	}
	if state == "completed" && existing.Valid {
		return existing.UUID, nil
	}
	if state != "identity_created" || !subject.Valid || subject.String == "" {
		return uuid.Nil, ErrPhoneIdentityBinding
	}
	var userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
 SELECT user_id FROM external_identities WHERE issuer=$1 AND subject=$2
 `, r.issuer, subject.String).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		userID = uuid.New()
		if _, err = tx.ExecContext(ctx, `INSERT INTO users(id) VALUES($1)`, userID); err != nil {
			return uuid.Nil, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO external_identities(user_id,issuer,subject) VALUES($1,$2,$3)`, userID, r.issuer, subject.String); err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	}
	var owner uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM user_verified_phones WHERE phone_number=$1`, phone).Scan(&owner)
	if err == nil && owner != userID {
		return uuid.Nil, ErrPhoneIdentityBinding
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, err
	}
	if _, err = tx.ExecContext(ctx, `
 INSERT INTO user_verified_phones(phone_number,user_id,issuer,subject)
 VALUES($1,$2,$3,$4) ON CONFLICT(phone_number) DO NOTHING
 `, phone, userID, r.issuer, subject.String); err != nil {
		return uuid.Nil, err
	}
	// Verify exact ownership, including subject and issuer, after any conflict.
	var matches bool
	if err = tx.QueryRowContext(ctx, `
 SELECT EXISTS(SELECT 1 FROM user_verified_phones WHERE
 phone_number=$1 AND user_id=$2 AND issuer=$3 AND subject=$4)
 `, phone, userID, r.issuer, subject.String).Scan(&matches); err != nil {
		return uuid.Nil, err
	}
	if !matches {
		return uuid.Nil, ErrPhoneIdentityBinding
	}
	if _, err = tx.ExecContext(ctx, `
 INSERT INTO user_capabilities(user_id,capability) VALUES($1,'rider') ON CONFLICT DO NOTHING
 `, userID); err != nil {
		return uuid.Nil, err
	}
	if _, err = tx.ExecContext(ctx, `
 UPDATE phone_registration_operations SET state='completed',user_id=$2,updated_at=clock_timestamp()
 WHERE id=$1
 `, id, userID); err != nil {
		return uuid.Nil, err
	}
	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

// ClaimRegistrationSubmission is an atomic, one-way claim before any external
// Kratos call. An ambiguous result must never be submitted again automatically.
func (r PostgresPhoneAccountResolver) ClaimRegistrationSubmission(ctx context.Context, id uuid.UUID) (bool, error) {
 result, err := r.db.ExecContext(ctx, `
 UPDATE phone_registration_operations
 SET state='submitting',updated_at=clock_timestamp()
 WHERE id=$1 AND issuer=$2 AND state='reserved'
 `, id, r.issuer)
 if err != nil { return false, err }
 count, err := result.RowsAffected()
 return count==1,err
}
