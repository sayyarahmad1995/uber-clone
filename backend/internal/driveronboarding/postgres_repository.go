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
		       a.status, COALESCE(a.rejection_reason, ''), a.submitted_at, a.decided_at
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
	now := time.Now().UTC()
	applicationID := uuid.New()
	_, err := r.db.ExecContext(ctx, `
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
