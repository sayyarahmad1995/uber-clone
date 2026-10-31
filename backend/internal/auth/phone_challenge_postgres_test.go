package auth

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

// Integration tests run when TEST_DATABASE_URL points to an isolated *_test database.
func phoneChallengeTestStore(t *testing.T) PhoneChallengeStore {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must name a *_test database")
	}
	admin, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "authotp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		admin.Close()
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := database.Open(u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(8)
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("apply migrations (including 032): %v", err)
	}
	return NewPhoneChallengeStore(db)
}

func phoneChallengeFixture(t *testing.T, store PhoneChallengeStore, phone string, expiry time.Time) (uuid.UUID, []byte) {
	t.Helper()
	id := uuid.New()
	codec, err := NewPhoneOTPCodec(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	digest, err := codec.Digest(id, "012345")
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.Create(context.Background(), NewPhoneChallenge{
		ID: id, PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister,
		CodeDigest: digest, IdempotencyKey: uuid.NewString(),
		ExpiresAt: expiry, ResendAfter: time.Now().Add(30 * time.Second),
	})
	if err != nil || !created.ShouldSend {
		t.Fatalf("create challenge: %+v %v", created, err)
	}
	return id, digest
}

func TestPhoneChallengePostgresLifecycle(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	id, correct := phoneChallengeFixture(t, store, "+923001234567", time.Now().Add(5*time.Minute))
	wrong := bytes.Repeat([]byte{0x44}, 32)
	for i := 0; i < 4; i++ {
		if _, err := store.Consume(ctx, id, wrong); !errors.Is(err, ErrPhoneChallengeInvalid) {
			t.Fatalf("wrong attempt %d: %v", i, err)
		}
	}
	if phone, err := store.Consume(ctx, id, correct); err != nil || phone != "+923001234567" {
		t.Fatalf("valid final attempt: %q %v", phone, err)
	}
	if _, err := store.Consume(ctx, id, correct); !errors.Is(err, ErrPhoneChallengeInvalid) {
		t.Fatalf("replay must fail: %v", err)
	}
	limitedID, valid := phoneChallengeFixture(t, store, "+923001234568", time.Now().Add(5*time.Minute))
	for i := 0; i < 5; i++ {
		_, _ = store.Consume(ctx, limitedID, wrong)
	}
	if _, err := store.Consume(ctx, limitedID, valid); !errors.Is(err, ErrPhoneChallengeLimited) {
		t.Fatalf("attempt limit ignored: %v", err)
	}
	expiredID, expiredDigest := phoneChallengeFixture(t, store, "+923001234569", time.Now().Add(3*time.Second))
	if _, err := store.db.ExecContext(ctx, "UPDATE auth_phone_challenges SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1", expiredID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume(ctx, expiredID, expiredDigest); !errors.Is(err, ErrPhoneChallengeInvalid) {
		t.Fatalf("expired challenge accepted: %v", err)
	}
}

func TestPhoneChallengePostgresDedupAndCooldown(t *testing.T) {
	store := phoneChallengeTestStore(t)
	ctx := context.Background()
	phone := "+923001234570"
	id, digest := phoneChallengeFixture(t, store, phone, time.Now().Add(5*time.Minute))
	var key string
	if err := store.db.QueryRowContext(ctx, "SELECT idempotency_key FROM auth_phone_challenges WHERE id=$1", id).Scan(&key); err != nil {
		t.Fatal(err)
	}
	replay, err := store.Create(ctx, NewPhoneChallenge{
		ID: uuid.New(), PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister,
		CodeDigest: digest, IdempotencyKey: key,
		ExpiresAt: time.Now().Add(5*time.Minute), ResendAfter: time.Now().Add(30*time.Second),
	})
	if err != nil || replay.ShouldSend || replay.ID != id {
		t.Fatalf("idempotency failed: %+v %v", replay, err)
	}
	_, err = store.Create(ctx, NewPhoneChallenge{
		ID: uuid.New(), PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister,
		CodeDigest: digest, IdempotencyKey: uuid.NewString(),
		ExpiresAt: time.Now().Add(5*time.Minute), ResendAfter: time.Now().Add(30*time.Second),
	})
	if !errors.Is(err, ErrPhoneChallengeCooldown) {
		t.Fatalf("different key bypassed cooldown: %v", err)
	}
}

func TestPhoneChallengePostgresConcurrentSingleUse(t *testing.T) {
	store := phoneChallengeTestStore(t)
	id, digest := phoneChallengeFixture(t, store, "+923001234571", time.Now().Add(5*time.Minute))
	const workers = 8
	results := make(chan error, workers)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := store.Consume(context.Background(), id, digest)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrPhoneChallengeInvalid) {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if success != 1 {
		t.Fatalf("expected exactly one successful consumption, got %d", success)
	}
}
