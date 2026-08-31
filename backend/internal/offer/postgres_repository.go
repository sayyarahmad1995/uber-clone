package offer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type PostgresRepository struct{ db *sql.DB }
func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db:db} }

func (r PostgresRepository) Market(ctx context.Context,rideRequestID uuid.UUID)(Market,error) {
	var market Market; var bookingMode,status string
	err:=r.db.QueryRowContext(ctx,`SELECT id,proposed_fare_minor,currency,booking_mode,status FROM ride_requests WHERE id=$1`,rideRequestID).Scan(&market.RideRequestID,&market.ProposedAmountMinor,&market.Currency,&bookingMode,&status)
	if errors.Is(err,sql.ErrNoRows) { return Market{},ErrRideNotFound }; if err != nil { return Market{},err }
	if bookingMode != "offers" || status != "requested" { return Market{},ErrRideNotOpen }
	return market,nil
}

func (r PostgresRepository) Upsert(ctx context.Context,rideRequestID,driverUserID uuid.UUID,amountMinor,minimumMinor,maximumMinor int64,currency string)(Offer,error) {
	tx,err:=r.db.BeginTx(ctx,nil); if err != nil { return Offer{},err }; defer tx.Rollback()
	var actualCurrency string; var riderUserID uuid.UUID; var bookingMode,status string
	if err:=tx.QueryRowContext(ctx,`SELECT currency,rider_user_id,booking_mode,status FROM ride_requests WHERE id=$1 FOR SHARE`,rideRequestID).Scan(&actualCurrency,&riderUserID,&bookingMode,&status); err != nil { if errors.Is(err,sql.ErrNoRows) { return Offer{},ErrRideNotFound }; return Offer{},err }
	if bookingMode != "offers" || status != "requested" || actualCurrency != currency { return Offer{},ErrRideNotOpen }
	if driverUserID == riderUserID { return Offer{},ErrDriverIneligible }
	if amountMinor < minimumMinor || amountMinor > maximumMinor { return Offer{},ErrAmountOutOfRange }
	var eligible bool
	if err:=tx.QueryRowContext(ctx,`SELECT EXISTS (SELECT 1 FROM driver_profiles p JOIN driver_vehicles v ON v.driver_user_id=p.user_id JOIN user_capabilities c ON c.user_id=p.user_id AND c.capability='driver' WHERE p.user_id=$1 AND p.status='active' AND p.is_online=TRUE AND NOT EXISTS (SELECT 1 FROM ride_driver_candidates active WHERE active.driver_user_id=p.user_id AND active.status IN ('pending','accepted') AND active.released_at IS NULL))`,driverUserID).Scan(&eligible); err != nil { return Offer{},err }
	if !eligible { return Offer{},ErrDriverIneligible }
	var result Offer
	if err:=tx.QueryRowContext(ctx,`INSERT INTO ride_offers (ride_request_id,driver_user_id,amount_minor,currency) VALUES ($1,$2,$3,$4) ON CONFLICT (ride_request_id,driver_user_id) DO UPDATE SET amount_minor=EXCLUDED.amount_minor,updated_at=NOW() RETURNING ride_request_id,driver_user_id,amount_minor,currency,created_at,updated_at`,rideRequestID,driverUserID,amountMinor,currency).Scan(&result.RideRequestID,&result.DriverUserID,&result.AmountMinor,&result.Currency,&result.CreatedAt,&result.UpdatedAt); err != nil { return Offer{},err }
	if err:=tx.Commit(); err != nil { return Offer{},err }; return result,nil
}

func (r PostgresRepository) ListForRider(ctx context.Context,rideRequestID,riderUserID uuid.UUID)([]Offer,error) {
	var ownedID uuid.UUID; var mode,status string
	if err:=r.db.QueryRowContext(ctx,`SELECT id,booking_mode,status FROM ride_requests WHERE id=$1 AND rider_user_id=$2`,rideRequestID,riderUserID).Scan(&ownedID,&mode,&status); err != nil { if errors.Is(err,sql.ErrNoRows) { return nil,ErrRideNotFound }; return nil,err }
	if mode != "offers" || status != "requested" { return nil,ErrRideNotOpen }
	rows,err:=r.db.QueryContext(ctx,`SELECT ride_request_id,driver_user_id,amount_minor,currency,created_at,updated_at FROM ride_offers WHERE ride_request_id=$1 ORDER BY amount_minor ASC,created_at ASC,driver_user_id ASC`,rideRequestID); if err != nil { return nil,err }; defer rows.Close()
	offers:=make([]Offer,0); for rows.Next() { var item Offer; if err:=rows.Scan(&item.RideRequestID,&item.DriverUserID,&item.AmountMinor,&item.Currency,&item.CreatedAt,&item.UpdatedAt); err != nil { return nil,err }; offers=append(offers,item) }
	if err:=rows.Err(); err != nil { return nil,err }; return offers,nil
}
