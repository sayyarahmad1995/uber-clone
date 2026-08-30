package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) CreateWithDefaultRider(ctx context.Context, identity ExternalIdentity) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT user_id
		FROM external_identities
		WHERE issuer = $1 AND subject = $2
	`, identity.Issuer, identity.Subject).Scan(&userID)

	if err == sql.ErrNoRows {
		userID = uuid.New()
		now := time.Now().UTC()

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, created_at) VALUES ($1, $2)
		`, userID, now); err != nil {
			return User{}, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO external_identities (user_id, issuer, subject, created_at)
			VALUES ($1, $2, $3, $4)
		`, userID, identity.Issuer, identity.Subject, now); err != nil {
			return User{}, err
		}
	} else if err != nil {
		return User{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_capabilities (user_id, capability)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, CapabilityRider); err != nil {
		return User{}, err
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return r.find(ctx, userID)
}

func (r PostgresRepository) AddCapability(ctx context.Context, userID uuid.UUID, capability Capability) (User, error) {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO user_capabilities (user_id, capability)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, capability); err != nil {
		return User{}, err
	}
	return r.find(ctx, userID)
}

func (r PostgresRepository) find(ctx context.Context, userID uuid.UUID) (User, error) {
	var u User
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, created_at FROM users WHERE id = $1
	`, userID).Scan(&u.ID, &u.CreatedAt); err != nil {
		return User{}, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT capability FROM user_capabilities
		WHERE user_id = $1
		ORDER BY capability
	`, u.ID)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var capability Capability
		if err := rows.Scan(&capability); err != nil {
			return User{}, err
		}
		u.Capabilities = append(u.Capabilities, capability)
	}
	return u, rows.Err()
}
