package kratos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

type phoneSessionBindingStub struct {
	allowed                bool
	err                    error
	issuer, subject, phone string
	calls                  int
}

func (s *phoneSessionBindingStub) MatchesVerifiedPhone(_ context.Context, issuer, subject, phone string) (bool, error) {
	s.calls++
	s.issuer, s.subject, s.phone = issuer, subject, phone
	return s.allowed, s.err
}

func TestEnsureVerifiedSessionPhoneBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" || r.Header.Get("X-Session-Token") != "token" {
			t.Errorf("unexpected request: %v %q", r.URL.Path, r.Header.Get("X-Session-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"id":"kratos-user","traits":{"phone":"+923001234567"},"verifiable_addresses":[]}}`))
	}))
	defer server.Close()
	provider := &Provider{baseURL: server.URL, client: server.Client()}
	if err := provider.EnsureVerifiedSession(context.Background(), "token"); !errors.Is(err, auth.ErrVerificationRequired) {
		t.Fatalf("unconfigured: %v", err)
	}
	checker := &phoneSessionBindingStub{}
	provider.WithVerifiedPhoneBindings("primary-identity-v1", checker)
	if err := provider.EnsureVerifiedSession(context.Background(), "token"); !errors.Is(err, auth.ErrVerificationRequired) {
		t.Fatalf("unbound: %v", err)
	}
	checker.allowed = true
	if err := provider.EnsureVerifiedSession(context.Background(), "token"); err != nil {
		t.Fatalf("bound: %v", err)
	}
	if checker.issuer != "primary-identity-v1" || checker.subject != "kratos-user" || checker.phone != "+923001234567" {
		t.Fatalf("lookup values: %+v", checker)
	}
	checker.err = errors.New("database failure")
	if err := provider.EnsureVerifiedSession(context.Background(), "token"); !errors.Is(err, auth.ErrUnavailable) {
		t.Fatalf("database fail-open: %v", err)
	}
}

func TestEnsureVerifiedSessionVerifiedEmailWithoutPhoneLookup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"id":"existing-user","traits":{"email":"test@example.com","phone":"+923001234567"},"verifiable_addresses":[{"verified":true}]}}`))
	}))
	defer server.Close()
	checker := &phoneSessionBindingStub{err: errors.New("database unavailable")}
	provider := &Provider{baseURL: server.URL, client: server.Client()}
	provider.WithVerifiedPhoneBindings("primary-identity-v1", checker)
	if err := provider.EnsureVerifiedSession(context.Background(), "token"); err != nil {
		t.Fatalf("email flow broken: %v", err)
	}
	if checker.calls != 0 {
		t.Fatal("verified email unnecessarily queried phone database")
	}
}
