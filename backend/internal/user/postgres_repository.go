package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) CreateWithDefaultRider(ctx context.Context, subject string) (User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return User{}, err }
	defer tx.Rollback()

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, external_subject, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (external_subject) DO NOTHING
	`, id, subject, now); err != nil { return User{}, err }

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_capabilities (user_id, capability)
		SELECT id, 'rider' FROM users WHERE external_subject = $1
		ON CONFLICT DO NOTHING
	`, subject); err != nil { return User{}, err }

	if err := tx.Commit(); err != nil { return User{}, err }

	return r.find(ctx, subject)
}

func (r PostgresRepository) find(ctx context.Context, subject string) (User, error) {
	var u User
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, external_subject, created_at FROM users WHERE external_subject = $1
	`, subject).Scan(&u.ID, &u.ExternalSubject, &u.CreatedAt); err != nil { return User{}, err }

	rows, err := r.db.QueryContext(ctx, `SELECT capability FROM user_capabilities WHERE user_id = $1 ORDER BY capability`, u.ID)
	if err != nil { return User{}, err }
	defer rows.Close()
	for rows.Next() {
		var capability Capability
		if err := rows.Scan(&capability); err != nil { return User{}, err }
		u.Capabilities = append(u.Capabilities, capability)
	}
	return u, rows.Err()
}
