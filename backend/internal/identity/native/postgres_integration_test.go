package native

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

// Require an explicitly disposable database; every integration test creates
// an isolated schema and never operates on the ordinary application database.
func nativeTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	u, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(u.Path, "_test") {
		t.Fatal("TEST_DATABASE_URL must reference a *_test database")
	}
	admin, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "native_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		_ = admin.Close()
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := database.Open(u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(10)
	if err := migrations.Apply(db); err != nil {
		t.Fatalf("apply isolated migrations: %v", err)
	}
	return db
}

func TestNativePostgresCredentialsAndSessions(t *testing.T) {
	db := nativeTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	if _, err := db.ExecContext(ctx, "INSERT INTO users(id) VALUES($1)", userID); err != nil {
		t.Fatal(err)
	}
	passwords, err := NewPostgresPasswords(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := passwords.ReadHash(ctx, userID); !errors.Is(err, ErrCredentialsNotFound) {
		t.Fatalf("missing credentials: %v", err)
	}
	hash, err := HashPassword("this is a long passphrase")
	if err != nil {
		t.Fatal(err)
	}
	if err := passwords.SetHash(ctx, userID, hash); err != nil {
		t.Fatal(err)
	}
	stored, err := passwords.ReadHash(ctx, userID)
	if err != nil || stored != hash {
		t.Fatalf("persisted hash mismatch: %v", err)
	}
	store, err := NewPostgresSessions(db)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := NewSessionManager(store)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := manager.Issue(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	id, _, err := digestBearer(tokens.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	var rawMatches int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM native_auth_sessions WHERE id=$1 AND access_digest=$2",
		id, []byte(tokens.AccessToken)).Scan(&rawMatches); err != nil {
		t.Fatal(err)
	}
	if rawMatches != 0 {
		t.Fatal("plain bearer token persisted instead of digest")
	}
	rotated, err := manager.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Refresh(ctx, tokens.RefreshToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("replayed refresh accepted: %v", err)
	}
	if _, err := manager.Authenticate(ctx, tokens.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("old access accepted after refresh: %v", err)
	}
	if user, err := manager.Authenticate(ctx, rotated.AccessToken); err != nil || user != userID {
		t.Fatalf("new access rejected: user=%s err=%v", user, err)
	}
	if err := manager.RevokeUser(ctx, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Authenticate(ctx, rotated.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("access survived account revocation: %v", err)
	}
}

func TestNativePostgresRefreshConcurrentSingleUse(t *testing.T) {
	db := nativeTestDB(t)
	ctx := context.Background()
	userID := uuid.New()
	if _, err := db.ExecContext(ctx, "INSERT INTO users(id) VALUES($1)", userID); err != nil {
		t.Fatal(err)
	}
	store, err := NewPostgresSessions(db)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := NewSessionManager(store)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := manager.Issue(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	const parallel = 6
	var wg sync.WaitGroup
	errs := make(chan error, parallel)
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := manager.Refresh(ctx, tokens.RefreshToken)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("unexpected refresh outcome: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected 1 successful refresh, got %d", successes)
	}
}
