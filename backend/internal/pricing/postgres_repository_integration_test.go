package pricing

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func pricingDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.HasSuffix(strings.TrimPrefix(u.Path, "/"), "_test") {
		t.Fatal("test database required")
	}
	db, err := database.Open(raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	code := "pricing_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := db.Exec("INSERT INTO driver_service_catalog(code,display_name,is_active,rider_visible,sort_order) VALUES($1,'Pricing fixture',true,true,999)", code); err != nil {
		t.Fatal(err)
	}
	return db, code
}
func draftFor(code string) Draft {
	return Draft{ServiceCode: code, Currency: "PKR", BaseFareMinor: 10000, RateMinorPerKm: 1000, RateMinorPerMinute: 100, MinimumFareMinor: 10000, RoundingIncrementMinor: 100}
}
func TestPricingPublishHistory(t *testing.T) {
	db, code := pricingDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()
	a, err := repo.Publish(ctx, draftFor(code), "reviewer-A")
	if err != nil {
		t.Fatal(err)
	}
	d := draftFor(code)
	d.BaseFareMinor = 20000
	b, err := repo.Publish(ctx, d, "reviewer-B")
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != 1 || b.Version != 2 || a.ID == b.ID || a.PublishedBy != "reviewer-A" || a.EffectiveFrom.IsZero() {
		t.Fatalf("versions %+v %+v", a, b)
	}
	current, err := repo.Current(ctx, code, "PKR")
	if err != nil || current.ID != b.ID {
		t.Fatalf("current %+v %v", current, err)
	}
	var base int64
	if err := db.QueryRow("SELECT base_fare_minor FROM ride_pricing_policy_versions WHERE id=$1", a.ID).Scan(&base); err != nil || base != 10000 {
		t.Fatalf("history %d %v", base, err)
	}
	if _, err := db.Exec("UPDATE ride_pricing_policy_versions SET base_fare_minor=1 WHERE id=$1", a.ID); err == nil {
		t.Fatal("history mutated")
	}
	if _, err := db.Exec("DELETE FROM ride_pricing_policy_versions WHERE id=$1", a.ID); err == nil {
		t.Fatal("history deleted")
	}
	if _, err := repo.Publish(ctx, draftFor("missing_service"), "reviewer"); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("unknown service %v", err)
	}
	db.Exec("UPDATE driver_service_catalog SET rider_visible=false WHERE code=$1", code)
	if _, err := repo.Publish(ctx, d, "reviewer"); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("hidden service %v", err)
	}
}
func TestPricingDisable(t *testing.T) {
	db, code := pricingDB(t)
	repo := NewPostgresRepository(db)
	ctx := context.Background()
	a, err := repo.Publish(ctx, draftFor(code), "A")
	if err != nil {
		t.Fatal(err)
	}
	b, err := repo.Publish(ctx, draftFor(code), "B")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Disable(ctx, code, "PKR", a.ID, "admin"); !errors.Is(err, ErrPolicyChanged) {
		t.Fatalf("stale disable %v", err)
	}
	if err := repo.Disable(ctx, code, "PKR", b.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Current(ctx, code, "PKR"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("disabled %v", err)
	}
	c, err := repo.Publish(ctx, draftFor(code), "C")
	if err != nil || c.Version != 3 {
		t.Fatalf("version after disable %+v %v", c, err)
	}
}
func TestPricingConcurrentPublish(t *testing.T) {
	db, code := pricingDB(t)
	repo := NewPostgresRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan Policy, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			p, e := repo.Publish(ctx, draftFor(code), "concurrent")
			results <- p
			errs <- e
		}()
	}
	close(start)
	wg.Wait()
	for i := 0; i < 2; i++ {
		if e := <-errs; e != nil {
			t.Fatal(e)
		}
	}
	a, b := <-results, <-results
	if a.Version+b.Version != 3 || a.Version == b.Version {
		t.Fatalf("versions %d %d", a.Version, b.Version)
	}
	var n int
	if err := db.QueryRow("SELECT count(*) FROM ride_pricing_policy_current WHERE service_code=$1", code).Scan(&n); err != nil || n != 1 {
		t.Fatalf("current pointers %d %v", n, err)
	}
}
func TestPricingMissingPointerRace(t *testing.T) {
	db, code := pricingDB(t)
	repo := NewPostgresRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT code FROM driver_service_catalog WHERE code=$1 FOR UPDATE", code); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 1)
	go func() { _, e := repo.Publish(ctx, draftFor(code), "publisher"); results <- e }()
	// Publication cannot bypass the shared service lock when no pointer exists.
	waitPricingBlock(t, ctx, db)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-results; err != nil {
		t.Fatal(err)
	}
	p, err := repo.Current(ctx, code, "PKR")
	if err != nil || p.Version != 1 {
		t.Fatalf("first publication %+v %v", p, err)
	}
}
func waitPricingBlock(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE '%driver_service_catalog%')").Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("expected blocked service lock")
		case <-ticker.C:
		}
	}
}
