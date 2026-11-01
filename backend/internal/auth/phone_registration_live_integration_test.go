package auth_test

import (
	"context"
 "errors"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
 identitykratos "github.com/sayyarahmad1995/uber-clone/backend/internal/identity/kratos"
	authkratos "github.com/sayyarahmad1995/uber-clone/backend/internal/auth/kratos"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
)

// Explicit opt-in: live Kratos creation, Redis proof consumption, PostgreSQL
// account creation. Only use disposable Kratos and isolated *_test database.
func TestLivePhoneRegistrationCoordinator(t *testing.T) {
	dsn, redisURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_REDIS_URL")
	publicURL, adminURL := os.Getenv("TEST_KRATOS_PUBLIC_URL"), os.Getenv("TEST_KRATOS_ADMIN_URL")
	if dsn == "" || redisURL == "" || publicURL == "" || adminURL == "" {
		t.Skip("requires TEST_DATABASE_URL, TEST_REDIS_URL and both TEST_KRATOS URLs")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || !strings.HasSuffix(parsed.Path, "_test") {
		t.Fatal("requires isolated *_test PostgreSQL database")
	}
	adminEndpoint, err := url.Parse(adminURL)
	if err != nil || adminEndpoint.Hostname() != "kratos-test" {
		t.Fatal("requires disposable kratos-test admin endpoint")
	}
	if u, e := url.Parse(publicURL); e != nil || u.Hostname() != "kratos-test" {
		t.Fatal("requires disposable kratos-test public endpoint")
	}
	root, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	schema := "phone_e2e_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = root.ExecContext(context.Background(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := root.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	}()
	q := parsed.Query()
	q.Set("search_path", schema)
	parsed.RawQuery = q.Encode()
	db, err := database.Open(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	repo, err := auth.NewPostgresPhoneAccountResolver(db, "primary-identity-v1")
	if err != nil {
		t.Fatal(err)
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	redisClient := redis.NewClient(opts)
	defer redisClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	proofs, err := auth.NewRedisPhoneProofs(redisClient)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := authkratos.New(publicURL, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 8)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	// Test-only E.164-shaped identifier; no SMS sent.
	phone := fmt.Sprintf("+9%013d", (uint64(random[0])<<40|uint64(random[1])<<32|uint64(random[2])<<24|uint64(random[3])<<16|uint64(random[4])<<8|uint64(random[5]))%10000000000000)
	tokenBytes := make([]byte, 32)
	if _, err = rand.Read(tokenBytes); err != nil {
		t.Fatal(err)
	}
	token := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(token))
	key := "higo:otp:proof:v1:" + hex.EncodeToString(hash[:])
	if err = redisClient.Set(ctx, key, phone+"|sign_in_or_register", 2*time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	defer redisClient.Del(context.Background(), key)
	service := auth.PhoneRegistrationCoordinator{Proofs: proofs, Identities: provider, Operations: repo}
	id, err := service.Register(ctx, token, "HiGO-Isolated-Registration-Test-123!")
	// Always cleanup after session tests, including when registration fails.
	defer func() {
		op, lookupErr := repo.GetRegistration(context.Background(), phone)
		if lookupErr == nil && op.Subject != "" {
		cleanupReq, e := http.NewRequest(http.MethodDelete, strings.TrimRight(adminURL, "/")+"/admin/identities/"+url.PathEscape(op.Subject), nil)
		if e == nil {
			resp, e := http.DefaultClient.Do(cleanupReq)
			if e != nil {
				t.Errorf("Kratos cleanup: %v", e)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
					t.Errorf("Kratos cleanup status: %s", resp.Status)
				}
			}
		}
		}
	}()
	op, lookupErr := repo.GetRegistration(ctx, phone)
	if lookupErr != nil { t.Fatal(lookupErr) }
	if err != nil || id == uuid.Nil {
		t.Fatalf("live coordinator failed: user=%v error=%v operation=%+v", id, err, op)
	}
	binding, err := repo.ResolveVerifiedPhone(ctx, phone)
	if err != nil || !binding.Exists || binding.IdentitySubject != op.Subject {
		t.Fatalf("binding mismatch: %+v %v", binding, err)
	}
	if op.State != "completed" || op.UserID != id {
		t.Fatalf("operation incomplete: %+v", op)
	}
	if _, err = service.Register(ctx, token, "HiGO-Isolated-Registration-Test-123!"); err == nil {
		t.Fatal("replayed proof unexpectedly accepted")
	}

	// Exercise the actual Kratos password login and both HiGO verification guards.
	const password = "HiGO-Isolated-Registration-Test-123!"
	provider.WithVerifiedPhoneBindings("primary-identity-v1", repo)
	routeIdentity, err := identitykratos.New(publicURL, "primary-identity-v1")
	if err != nil { t.Fatal(err) }
	routeIdentity.WithVerifiedPhoneBindings(repo)
	if _, err = provider.Login(ctx, auth.Credentials{Identifier:phone,Password:"incorrect-password"}); !errors.Is(err,auth.ErrInvalidCredentials) {
		t.Fatalf("incorrect password accepted: %v",err)
	}
	session, err := provider.Login(ctx, auth.Credentials{Identifier:phone,Password:password})
	if err != nil || session.AccessToken=="" { t.Fatalf("phone login failed: %v",err) }
	if err=provider.EnsureVerifiedSession(ctx,session.AccessToken); err!=nil {t.Fatalf("login authorization: %v",err)}
	principal, err:=routeIdentity.AuthenticateVerified(ctx,session.AccessToken)
	if err!=nil || principal.Subject!=op.Subject || principal.Issuer!="primary-identity-v1" {
		t.Fatalf("protected identity guard: %+v %v",principal,err)
	}
	if _,err=routeIdentity.AuthenticateVerified(ctx,"invalid-session"); !errors.Is(err,identity.ErrInvalidToken){
		t.Fatalf("invalid token accepted: %v",err)
	}
	// Revoking phone ownership must immediately prevent both HiGO authorization paths.
	if _,err=db.ExecContext(ctx,"DELETE FROM user_verified_phones WHERE phone_number=$1",phone);err!=nil {t.Fatal(err)}
	if err=provider.EnsureVerifiedSession(ctx,session.AccessToken); !errors.Is(err,auth.ErrVerificationRequired){
		t.Fatalf("unbound phone passed login guard: %v",err)
	}
	if _,err=routeIdentity.AuthenticateVerified(ctx,session.AccessToken); !errors.Is(err,identity.ErrVerificationRequired){
		t.Fatalf("unbound phone passed protected-route guard: %v",err)
	}
	if err=provider.Logout(ctx,session.AccessToken);err!=nil {t.Fatalf("logout: %v",err)}
	if _,err=routeIdentity.AuthenticateVerified(ctx,session.AccessToken); !errors.Is(err,identity.ErrInvalidToken){
		t.Fatalf("logged-out session accepted: %v",err)
	}
}
