package ride

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
)

// Removing LockCurrent's catalog lock must break both orderings: a missing
// pointer has no row to lock, and a waiting creator must read the first pointer
// in a new statement after publication commits.
func TestRequestPricingFirstPublicationRace(t *testing.T) {
	t.Run("creator first rejects without writes and blocks publication", func(t *testing.T) {
		db, code, user, policies := requestPricingEmptyDB(t)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		repo := NewPostgresRepositoryWithPricing(db, policies, true)
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		var creatorPID int
		if err := tx.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&creatorPID); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.createPricedInTx(ctx, tx, user, requestInput(code, pricing.Policy{ID: uuid.New()})); !errors.Is(err, pricing.ErrUnavailable) {
			t.Fatalf("unpriced request: %v; want pricing unavailable", err)
		}
		assertFirstPricingNoWrites(t, db, user)
		published := make(chan firstPricingPublication, 1)
		go func() {
			p, err := policies.Publish(ctx, requestDraft(code), "first-policy")
			published <- firstPricingPublication{p, err}
		}()
		waitFirstPricingBlockedBy(t, ctx, db, creatorPID)
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		result := <-published
		if result.err != nil {
			t.Fatal(result.err)
		}
		assertFirstPricingNoWrites(t, db, user)
		assertFirstPricingRequest(t, ctx, db, repo, code, user, result.policy)
	})

	t.Run("publisher first creator waits and sees committed pointer", func(t *testing.T) {
		db, code, user, policies := requestPricingEmptyDB(t)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Pause the real Publish after it owns the catalog lock, before its
		// first policy insert completes. The pause is confined to this service.
		gate, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer gate.Rollback()
		var gatePID int
		if err := gate.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&gatePID); err != nil {
			t.Fatal(err)
		}
		key := time.Now().UnixNano()
		if _, err := gate.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", key); err != nil {
			t.Fatal(err)
		}
		name := "first_" + code
		functionSQL := fmt.Sprintf("CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_advisory_xact_lock(%d); RETURN NEW; END; $$", name, key)
		if _, err := db.ExecContext(ctx, functionSQL); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.Exec("DROP FUNCTION " + name + "()") })
		triggerSQL := fmt.Sprintf("CREATE TRIGGER %s BEFORE INSERT ON ride_pricing_policy_versions FOR EACH ROW WHEN (NEW.service_code='%s') EXECUTE FUNCTION %s()", name, code, name)
		if _, err := db.ExecContext(ctx, triggerSQL); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = db.Exec("DROP TRIGGER " + name + " ON ride_pricing_policy_versions") })
		published := make(chan firstPricingPublication, 1)
		go func() {
			p, err := policies.Publish(ctx, requestDraft(code), "first-policy")
			published <- firstPricingPublication{p, err}
		}()
		publisherPID := waitFirstPricingBlockedBy(t, ctx, db, gatePID)
		repo := NewPostgresRepositoryWithPricing(db, policies, true)
		created := make(chan error, 1)
		go func() {
			_, err := repo.Create(ctx, user, requestInput(code, pricing.Policy{ID: uuid.New()}))
			created <- err
		}()
		waitFirstPricingBlockedBy(t, ctx, db, publisherPID)
		assertFirstPricingNoWrites(t, db, user)
		if err := gate.Commit(); err != nil {
			t.Fatal(err)
		}
		result := <-published
		if result.err != nil {
			t.Fatal(result.err)
		}
		// PolicyChanged rather than Unavailable proves the waiter saw the
		// newly committed first pointer, then rejected its unknown UUID.
		if err := <-created; !errors.Is(err, pricing.ErrPolicyChanged) {
			t.Fatalf("waiting request: %v; want policy changed", err)
		}
		assertFirstPricingNoWrites(t, db, user)
		assertFirstPricingRequest(t, ctx, db, repo, code, user, result.policy)
	})
}

type firstPricingPublication struct {
	policy pricing.Policy
	err    error
}

func waitFirstPricingBlockedBy(t *testing.T, ctx context.Context, db *sql.DB, blockerPID int) int {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var pid int
		err := db.QueryRowContext(ctx, "SELECT pid FROM pg_stat_activity WHERE datname=current_database() AND $1=ANY(pg_blocking_pids(pid)) LIMIT 1", blockerPID).Scan(&pid)
		if err == nil {
			return pid
		}
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("waiting for blocker %d: %v", blockerPID, err)
		}
		select {
		case <-ctx.Done():
			t.Fatalf("expected transaction blocked by PID %d: %v", blockerPID, ctx.Err())
		case <-ticker.C:
		}
	}
}

func assertFirstPricingNoWrites(t *testing.T, db *sql.DB, user uuid.UUID) {
	t.Helper()
	assertRequestCount(t, db, user, 0)
	var snapshots int
	if err := db.QueryRow("SELECT count(*) FROM ride_request_pricing_snapshots s JOIN ride_requests r ON r.id=s.ride_request_id WHERE r.rider_user_id=$1", user).Scan(&snapshots); err != nil || snapshots != 0 {
		t.Fatalf("snapshots %d; want 0: %v", snapshots, err)
	}
}

func assertFirstPricingRequest(t *testing.T, ctx context.Context, db *sql.DB, repo PostgresRepository, code string, user uuid.UUID, p pricing.Policy) {
	t.Helper()
	if p.Version != 1 {
		t.Fatalf("first publication version %d; want 1", p.Version)
	}
	r, err := repo.Create(ctx, user, requestInput(code, p))
	if err != nil {
		t.Fatal(err)
	}
	var id uuid.UUID
	var version, base, amount int64
	if err := db.QueryRowContext(ctx, "SELECT s.policy_id,s.version,s.base_fare_minor,r.proposed_fare_minor FROM ride_request_pricing_snapshots s JOIN ride_requests r ON r.id=s.ride_request_id WHERE r.id=$1", r.ID).Scan(&id, &version, &base, &amount); err != nil {
		t.Fatal(err)
	}
	if id != p.ID || version != 1 || base != 10000 || amount != 110000 {
		t.Fatalf("first-policy snapshot id=%s version=%d base=%d amount=%d", id, version, base, amount)
	}
	assertRequestCount(t, db, user, 1)
}
