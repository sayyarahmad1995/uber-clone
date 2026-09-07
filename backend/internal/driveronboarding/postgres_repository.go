package driveronboarding

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (r PostgresRepository) ListServices(ctx context.Context) ([]ServiceOption, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, display_name, description, minimum_model_year, implied_service_code
		FROM driver_service_catalog
		WHERE is_active = TRUE
		ORDER BY sort_order, display_name, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []ServiceOption{}
	for rows.Next() {
		service, err := scanService(rows.Scan)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return services, nil
}

func (r PostgresRepository) FindService(ctx context.Context, code string) (ServiceOption, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT code, display_name, description, minimum_model_year, implied_service_code
		FROM driver_service_catalog
		WHERE code = $1 AND is_active = TRUE
	`, strings.ToLower(strings.TrimSpace(code)))
	service, err := scanService(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceOption{}, ErrServiceNotFound
	}
	if err != nil {
		return ServiceOption{}, err
	}
	return service, nil
}

func (r PostgresRepository) LatestApplication(ctx context.Context, userID uuid.UUID) (Application, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT a.id, a.driver_user_id, a.display_name,
		       s.code, s.display_name, s.description, s.minimum_model_year, s.implied_service_code,
		       a.vehicle_make, a.vehicle_model, a.vehicle_model_year, a.vehicle_color, a.vehicle_license_plate,
		       a.status, COALESCE(a.rejection_reason, ''), a.submitted_at, a.decided_at,
		       COALESCE(a.decided_by, '')
		FROM driver_onboarding_applications a
		JOIN driver_service_catalog s ON s.code = a.service_code
		WHERE a.driver_user_id = $1
		ORDER BY a.submitted_at DESC, a.id DESC
		LIMIT 1
	`, userID)
	application, err := scanApplication(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	return application, nil
}

func (r PostgresRepository) CreateApplication(ctx context.Context, userID uuid.UUID, input ApplicationInput, service ServiceOption) (Application, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback()

	if err := lockDriverReviewScope(ctx, tx, userID); err != nil {
		return Application{}, err
	}

	var alreadyOnboarded bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM driver_profiles WHERE user_id = $1
			UNION ALL
			SELECT 1 FROM driver_onboarding_applications
			WHERE driver_user_id = $1 AND status = 'approved'
		)
	`, userID).Scan(&alreadyOnboarded); err != nil {
		return Application{}, err
	}
	if alreadyOnboarded {
		return Application{}, ErrAlreadyOnboarded
	}

	now := time.Now().UTC()
	applicationID := uuid.New()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO driver_onboarding_applications (
			id, driver_user_id, display_name, service_code,
			vehicle_make, vehicle_model, vehicle_model_year, vehicle_color, vehicle_license_plate,
			status, submitted_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11, $11)
	`, applicationID, userID, input.DisplayName, service.Code,
		input.Vehicle.Make, input.Vehicle.Model, input.Vehicle.ModelYear, input.Vehicle.Color, input.Vehicle.LicensePlate,
		StatusPending, now)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName == "driver_onboarding_one_pending_per_driver" {
			return Application{}, ErrPendingApplication
		}
		return Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return Application{}, err
	}
	return Application{
		ID:           applicationID,
		DriverUserID: userID,
		DisplayName:  input.DisplayName,
		Service:      service,
		Vehicle:      input.Vehicle,
		Status:       StatusPending,
		SubmittedAt:  now,
	}, nil
}

func (r PostgresRepository) ListPendingApplications(ctx context.Context) ([]Application, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.driver_user_id, a.display_name,
		       s.code, s.display_name, s.description, s.minimum_model_year, s.implied_service_code,
		       a.vehicle_make, a.vehicle_model, a.vehicle_model_year, a.vehicle_color, a.vehicle_license_plate,
		       a.status, COALESCE(a.rejection_reason, ''), a.submitted_at, a.decided_at,
		       COALESCE(a.decided_by, '')
		FROM driver_onboarding_applications a
		JOIN driver_service_catalog s ON s.code = a.service_code
		WHERE a.status = 'pending'
		ORDER BY a.submitted_at ASC, a.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applications := []Application{}
	for rows.Next() {
		application, err := scanApplication(rows.Scan)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return applications, nil
}

func (r PostgresRepository) FindApplicationByID(ctx context.Context, applicationID uuid.UUID) (Application, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT a.id, a.driver_user_id, a.display_name,
		       s.code, s.display_name, s.description, s.minimum_model_year, s.implied_service_code,
		       a.vehicle_make, a.vehicle_model, a.vehicle_model_year, a.vehicle_color, a.vehicle_license_plate,
		       a.status, COALESCE(a.rejection_reason, ''), a.submitted_at, a.decided_at,
		       COALESCE(a.decided_by, '')
		FROM driver_onboarding_applications a
		JOIN driver_service_catalog s ON s.code = a.service_code
		WHERE a.id = $1
	`, applicationID)
	application, err := scanApplication(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	return application, nil
}

func (r PostgresRepository) ApproveApplication(ctx context.Context, applicationID uuid.UUID, reviewer string) (Application, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback()

	driverUserID, err := loadApplicationDriverUserID(ctx, tx, applicationID)
	if err != nil {
		return Application{}, err
	}
	if err := lockDriverReviewScope(ctx, tx, driverUserID); err != nil {
		return Application{}, err
	}

	application, err := loadApplicationForUpdate(ctx, tx, applicationID)
	if err != nil {
		return Application{}, err
	}
	if application.Status != StatusPending {
		return Application{}, ErrApplicationNotPending
	}

	var profileExists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM driver_profiles WHERE user_id = $1)`, application.DriverUserID).Scan(&profileExists); err != nil {
		return Application{}, err
	}
	if profileExists {
		return Application{}, ErrAlreadyOnboarded
	}

	now := time.Now().UTC()
	vehicleID := uuid.New()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_profiles (user_id, display_name, status, is_online, created_at, updated_at)
		VALUES ($1, $2, 'approved', FALSE, $3, $3)
	`, application.DriverUserID, application.DisplayName, now); err != nil {
		return Application{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_vehicles (
			id, driver_user_id, make, model, model_year, color, license_plate, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, vehicleID, application.DriverUserID, application.Vehicle.Make, application.Vehicle.Model,
		application.Vehicle.ModelYear, application.Vehicle.Color, application.Vehicle.LicensePlate, now); err != nil {
		return Application{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO driver_vehicle_service_enrollments (
			vehicle_id, service_code, approved_at, approved_by, created_at
		) VALUES ($1, $2, $3, $4, $3)
	`, vehicleID, application.Service.Code, now, reviewer); err != nil {
		return Application{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE driver_onboarding_applications
		SET status = 'approved', rejection_reason = NULL, decided_at = $2, decided_by = $3, updated_at = $2
		WHERE id = $1
	`, application.ID, now, reviewer); err != nil {
		return Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return Application{}, err
	}

	application.Status = StatusApproved
	application.RejectionReason = ""
	application.DecidedAt = &now
	application.DecidedBy = reviewer
	return application, nil
}

func (r PostgresRepository) RejectApplication(ctx context.Context, applicationID uuid.UUID, reviewer, reason string) (Application, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Application{}, err
	}
	defer tx.Rollback()

	driverUserID, err := loadApplicationDriverUserID(ctx, tx, applicationID)
	if err != nil {
		return Application{}, err
	}
	if err := lockDriverReviewScope(ctx, tx, driverUserID); err != nil {
		return Application{}, err
	}

	application, err := loadApplicationForUpdate(ctx, tx, applicationID)
	if err != nil {
		return Application{}, err
	}
	if application.Status != StatusPending {
		return Application{}, ErrApplicationNotPending
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		UPDATE driver_onboarding_applications
		SET status = 'rejected', rejection_reason = $2, decided_at = $3, decided_by = $4, updated_at = $3
		WHERE id = $1
	`, application.ID, reason, now, reviewer); err != nil {
		return Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return Application{}, err
	}

	application.Status = StatusRejected
	application.RejectionReason = reason
	application.DecidedAt = &now
	application.DecidedBy = reviewer
	return application, nil
}

func loadApplicationDriverUserID(ctx context.Context, tx *sql.Tx, applicationID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	if err := tx.QueryRowContext(ctx, `
		SELECT driver_user_id
		FROM driver_onboarding_applications
		WHERE id = $1
	`, applicationID).Scan(&userID); errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	} else if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func lockDriverReviewScope(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	var lockedUserID uuid.UUID
	return tx.QueryRowContext(ctx, `
		SELECT id
		FROM users
		WHERE id = $1
		FOR UPDATE
	`, userID).Scan(&lockedUserID)
}

func loadApplicationForUpdate(ctx context.Context, tx *sql.Tx, applicationID uuid.UUID) (Application, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT a.id, a.driver_user_id, a.display_name,
		       s.code, s.display_name, s.description, s.minimum_model_year, s.implied_service_code,
		       a.vehicle_make, a.vehicle_model, a.vehicle_model_year, a.vehicle_color, a.vehicle_license_plate,
		       a.status, COALESCE(a.rejection_reason, ''), a.submitted_at, a.decided_at,
		       COALESCE(a.decided_by, '')
		FROM driver_onboarding_applications a
		JOIN driver_service_catalog s ON s.code = a.service_code
		WHERE a.id = $1
		FOR UPDATE OF a
	`, applicationID)
	application, err := scanApplication(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	return application, nil
}

type scanFunc func(dest ...any) error

func scanService(scan scanFunc) (ServiceOption, error) {
	var service ServiceOption
	var minimumModelYear sql.NullInt64
	var impliedServiceCode sql.NullString
	if err := scan(
		&service.Code,
		&service.DisplayName,
		&service.Description,
		&minimumModelYear,
		&impliedServiceCode,
	); err != nil {
		return ServiceOption{}, err
	}
	if minimumModelYear.Valid {
		year := int(minimumModelYear.Int64)
		service.MinimumModelYear = &year
	}
	if impliedServiceCode.Valid {
		code := impliedServiceCode.String
		service.ImpliedServiceCode = &code
	}
	return service, nil
}

func scanApplication(scan scanFunc) (Application, error) {
	var application Application
	var minimumModelYear sql.NullInt64
	var impliedServiceCode sql.NullString
	var decidedAt sql.NullTime
	if err := scan(
		&application.ID,
		&application.DriverUserID,
		&application.DisplayName,
		&application.Service.Code,
		&application.Service.DisplayName,
		&application.Service.Description,
		&minimumModelYear,
		&impliedServiceCode,
		&application.Vehicle.Make,
		&application.Vehicle.Model,
		&application.Vehicle.ModelYear,
		&application.Vehicle.Color,
		&application.Vehicle.LicensePlate,
		&application.Status,
		&application.RejectionReason,
		&application.SubmittedAt,
		&decidedAt,
		&application.DecidedBy,
	); err != nil {
		return Application{}, err
	}
	if minimumModelYear.Valid {
		year := int(minimumModelYear.Int64)
		application.Service.MinimumModelYear = &year
	}
	if impliedServiceCode.Valid {
		code := impliedServiceCode.String
		application.Service.ImpliedServiceCode = &code
	}
	if decidedAt.Valid {
		value := decidedAt.Time
		application.DecidedAt = &value
	}
	return application, nil
}
