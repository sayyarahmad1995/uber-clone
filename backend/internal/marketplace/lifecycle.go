package marketplace

import (
	"context"
	"database/sql"
	"time"
)

const (
	DefaultRideRequestTTL      = 3 * time.Minute
	DefaultDriverOpportunityTTL = 30 * time.Second
	DefaultOfferDecisionTTL     = 10 * time.Second
)

type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Expire materializes deadline-driven marketplace state using PostgreSQL time.
// PostgreSQL remains authoritative even when a future cache/notification layer exists.
func Expire(ctx context.Context, db Execer) error {
	statements := []string{
		`UPDATE ride_offers
		 SET status = 'expired', decided_at = statement_timestamp(), updated_at = statement_timestamp()
		 WHERE status = 'pending' AND expires_at <= statement_timestamp()`,
		`UPDATE driver_ride_request_opportunities o
		 SET status = 'offer_expired', responded_at = COALESCE(o.responded_at, statement_timestamp())
		 WHERE o.status = 'offered'
		   AND EXISTS (
		       SELECT 1 FROM ride_offers ro
		       WHERE ro.ride_request_id = o.ride_request_id
		         AND ro.driver_user_id = o.driver_user_id
		         AND ro.status = 'expired'
		   )`,
		`UPDATE driver_ride_request_opportunities
		 SET status = 'window_expired', responded_at = statement_timestamp()
		 WHERE status = 'open' AND visible_until <= statement_timestamp()`,
		`UPDATE ride_requests rr
		 SET status = 'expired'
		 WHERE rr.status = 'requested'
		   AND rr.expires_at <= statement_timestamp()
		   AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.ride_request_id = rr.id)`,
		`UPDATE ride_offers ro
		 SET status = 'closed', decided_at = statement_timestamp(), updated_at = statement_timestamp()
		 WHERE ro.status = 'pending'
		   AND EXISTS (
		       SELECT 1 FROM ride_requests rr
		       WHERE rr.id = ro.ride_request_id AND rr.status IN ('accepted', 'cancelled', 'expired')
		   )`,
		`UPDATE driver_ride_request_opportunities o
		 SET status = 'closed', responded_at = COALESCE(o.responded_at, statement_timestamp())
		 WHERE o.status IN ('open', 'offered')
		   AND EXISTS (
		       SELECT 1 FROM ride_requests rr
		       WHERE rr.id = o.ride_request_id AND rr.status IN ('accepted', 'cancelled', 'expired')
		   )`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
