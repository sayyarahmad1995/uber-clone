package ride

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func requestPricingDB(t *testing.T) (*sql.DB, string, uuid.UUID, *pricing.PostgresRepository, pricing.Policy) {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	u, e := url.Parse(raw)
	if e != nil || !strings.HasSuffix(strings.TrimPrefix(u.Path, "/"), "_test") {
		t.Fatal("test database required")
	}
	db, e := database.Open(raw)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e := migrations.Apply(db); e != nil {
		t.Fatal(e)
	}
	code := "request_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	user := uuid.New()
	if _, e := db.Exec("INSERT INTO users(id) VALUES($1)", user); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec("INSERT INTO driver_service_catalog(code,display_name,is_active,rider_visible) VALUES($1,'Request pricing test',true,true)", code); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { _, _ = db.Exec("UPDATE driver_service_catalog SET rider_visible=false WHERE code=$1", code) })
	policies := pricing.NewPostgresRepository(db)
	p, e := policies.Publish(context.Background(), requestDraft(code), "owner-A")
	if e != nil {
		t.Fatal(e)
	}
	return db, code, user, policies, p
}
func requestDraft(code string) pricing.Draft {
	return pricing.Draft{ServiceCode: code, Currency: "PKR", BaseFareMinor: 10000, RateMinorPerKm: 1000, RateMinorPerMinute: 100, MinimumFareMinor: 10000, RoundingIncrementMinor: 100}
}
func requestInput(code string, p pricing.Policy) CreateInput {
	return CreateInput{ServiceCode: code, Pickup: Location{24.86, 67.01}, Destination: Location{24.90, 67.05}, ProposedFare: &Money{110000, "PKR"}, PricingPolicyVersion: p.ID.String()}
}
func TestRequestPricingSnapshot(t *testing.T) {
	db, code, user, policies, a := requestPricingDB(t)
	repo := NewPostgresRepositoryWithPricing(db, policies, true)
	r, e := repo.Create(context.Background(), user, requestInput(code, a))
	if e != nil {
		t.Fatal(e)
	}
	if r.ProposedFare.AmountMinor != 110000 || r.PricingSnapshot == nil || r.PricingSnapshot.PolicyID != a.ID || r.PricingSnapshot.BaseFareMinor != 10000 || r.PricingSnapshot.RateMinorPerKm != 1000 || r.PricingSnapshot.RateMinorPerMinute != 100 || r.PricingSnapshot.MinimumFareMinor != 10000 || r.PricingSnapshot.RoundingIncrementMinor != 100 || r.PricingSnapshot.Version != 1 || r.PricingSnapshot.Currency != "PKR" || r.PricingSnapshot.SnapshotAt.IsZero() {
		t.Fatalf("snapshot %+v", r)
	}
	d := requestDraft(code)
	d.BaseFareMinor = 20000
	if _, e := policies.Publish(context.Background(), d, "owner-B"); e != nil {
		t.Fatal(e)
	}
	var base, proposal int64
	if e := db.QueryRow("SELECT s.base_fare_minor,r.proposed_fare_minor FROM ride_request_pricing_snapshots s JOIN ride_requests r ON r.id=s.ride_request_id WHERE r.id=$1", r.ID).Scan(&base, &proposal); e != nil || base != 10000 || proposal != 110000 {
		t.Fatalf("repriced %d %d %v", base, proposal, e)
	}
	if _, e := db.Exec("UPDATE ride_request_pricing_snapshots SET base_fare_minor=1 WHERE ride_request_id=$1", r.ID); e == nil {
		t.Fatal("snapshot mutated")
	}
	if _, e := db.Exec("DELETE FROM ride_request_pricing_snapshots WHERE ride_request_id=$1", r.ID); e == nil {
		t.Fatal("snapshot deleted")
	}
	legacyInput := requestInput(code, a)
	legacyInput.PricingPolicyVersion = ""
	legacy, e := NewPostgresRepository(db).Create(context.Background(), user, legacyInput)
	if e != nil || legacy.PricingSnapshot != nil {
		t.Fatalf("legacy %+v %v", legacy, e)
	}
}
func TestRequestPricingPolicyChanged(t *testing.T) {
	db, code, user, policies, a := requestPricingDB(t)
	repo := NewPostgresRepositoryWithPricing(db, policies, true)
	b, e := policies.Publish(context.Background(), requestDraft(code), "B")
	if e != nil {
		t.Fatal(e)
	}
	if _, e := repo.Create(context.Background(), user, requestInput(code, a)); !errors.Is(e, pricing.ErrPolicyChanged) {
		t.Fatalf("stale version %v", e)
	}
	unknown := a
	unknown.ID = uuid.New()
	if _, e := repo.Create(context.Background(), user, requestInput(code, unknown)); !errors.Is(e, pricing.ErrPolicyChanged) {
		t.Fatalf("unknown version %v", e)
	}
	in := requestInput(code, b)
	in.PricingPolicyVersion = ""
	if _, e := repo.Create(context.Background(), user, in); !errors.Is(e, pricing.ErrInvalidPolicy) {
		t.Fatalf("missing version %v", e)
	}
	in = requestInput(code, b)
	in.ProposedFare.Currency = "USD"
	if _, e := repo.Create(context.Background(), user, in); !errors.Is(e, pricing.ErrInvalidPolicy) {
		t.Fatalf("currency %v", e)
	}
	assertRequestCount(t, db, user, 0)
}
func TestRequestPricingRollback(t *testing.T) {
	db, code, user, policies, a := requestPricingDB(t)
	name := "fail_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, e := db.Exec("CREATE FUNCTION " + name + "() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'fixture snapshot failure'; END; $$"); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec("CREATE TRIGGER " + name + " BEFORE INSERT ON ride_request_pricing_snapshots FOR EACH ROW WHEN (NEW.service_code='" + code + "') EXECUTE FUNCTION " + name + "()"); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		db.Exec("DROP TRIGGER " + name + " ON ride_request_pricing_snapshots")
		db.Exec("DROP FUNCTION " + name + "()")
	})
	if _, e := NewPostgresRepositoryWithPricing(db, policies, true).Create(context.Background(), user, requestInput(code, a)); e == nil {
		t.Fatal("snapshot failure accepted")
	}
	assertRequestCount(t, db, user, 0)
}
func TestRequestPricingPublicationRace(t *testing.T) {
	t.Run("creator locks first", func(t *testing.T) {
		db, code, user, policies, a := requestPricingDB(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		lock, e := db.BeginTx(ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		defer lock.Rollback()
		// Hold a compatible service lock; a pending publisher must wait.
		if _, e := pricing.LockCurrent(ctx, lock, code, "PKR"); e != nil {
			t.Fatal(e)
		}
		pub := make(chan error, 1)
		go func() { _, e := policies.Publish(ctx, requestDraft(code), "B"); pub <- e }()
		waitRequestBlock(t, ctx, db)
		// Avoid PostgreSQL's queued-lock fairness: create under the already-held transaction.
		repo := NewPostgresRepositoryWithPricing(db, policies, true)
		r, e := repo.createPricedInTx(ctx, lock, user, requestInput(code, a))
		if e != nil {
			t.Fatal(e)
		}
		if e := lock.Commit(); e != nil {
			t.Fatal(e)
		}
		if e := <-pub; e != nil {
			t.Fatal(e)
		}
		var id uuid.UUID
		if e := db.QueryRow("SELECT policy_id FROM ride_request_pricing_snapshots WHERE ride_request_id=$1", r.ID).Scan(&id); e != nil || id != a.ID {
			t.Fatalf("snapshot %s %v", id, e)
		}
	})
	t.Run("publisher locks first", func(t *testing.T) {
		db, code, user, policies, a := requestPricingDB(t)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		b, e := policies.Publish(ctx, requestDraft(code), "prepared-B")
		if e != nil {
			t.Fatal(e)
		}
		if _, e := db.Exec("UPDATE ride_pricing_policy_current SET current_policy_id=$1 WHERE service_code=$2", a.ID, code); e != nil {
			t.Fatal(e)
		}
		tx, e := db.BeginTx(ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		defer tx.Rollback()
		if _, e := tx.ExecContext(ctx, "SELECT code FROM driver_service_catalog WHERE code=$1 FOR UPDATE", code); e != nil {
			t.Fatal(e)
		}
		done := make(chan error, 1)
		go func() {
			_, e := NewPostgresRepositoryWithPricing(db, policies, true).Create(ctx, user, requestInput(code, a))
			done <- e
		}()
		waitRequestBlock(t, ctx, db)
		// Commit the new pointer while holding the same exclusive lock as publication.
		if _, e := tx.ExecContext(ctx, "UPDATE ride_pricing_policy_current SET current_policy_id=$1 WHERE service_code=$2", b.ID, code); e != nil {
			t.Fatal(e)
		}
		if e := tx.Commit(); e != nil {
			t.Fatal(e)
		}
		if e := <-done; !errors.Is(e, pricing.ErrPolicyChanged) {
			t.Fatalf("published race %v", e)
		}
		assertRequestCount(t, db, user, 0)
	})
}
func assertRequestCount(t *testing.T, db *sql.DB, user uuid.UUID, want int) {
	t.Helper()
	var n int
	if e := db.QueryRow("SELECT count(*) FROM ride_requests WHERE rider_user_id=$1", user).Scan(&n); e != nil || n != want {
		t.Fatalf("requests %d want %d: %v", n, want, e)
	}
}
func waitRequestBlock(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	timer := time.NewTicker(5 * time.Millisecond)
	defer timer.Stop()
	for {
		var b bool
		if e := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE '%driver_service_catalog%')").Scan(&b); e != nil {
			t.Fatal(e)
		}
		if b {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("expected blocked lock")
		case <-timer.C:
		}
	}
}
