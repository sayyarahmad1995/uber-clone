package kratos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

type stubVerifiedPhone struct {
	allow                  bool
	err                    error
	called                 int
	issuer, subject, phone string
}

func (s *stubVerifiedPhone) MatchesVerifiedPhone(_ context.Context, issuer, subject, phone string) (bool, error) {
	s.called++
	s.issuer = issuer
	s.subject = subject
	s.phone = phone
	return s.allow, s.err
}

func TestAuthenticateVerifiedPhoneBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"id":"kratos-user","traits":{"phone":"+923001234567"},"verifiable_addresses":[]}}`))
	}))
	defer server.Close()
	provider := &Provider{baseURL: server.URL, source: "primary-identity-v1", client: server.Client()}
	if _, err := provider.AuthenticateVerified(context.Background(), "token"); !errors.Is(err, identity.ErrVerificationRequired) {
		t.Fatalf("unconfigured phone allowed: %v", err)
	}
	checker := &stubVerifiedPhone{}
	provider.WithVerifiedPhoneBindings(checker)
	if _, err := provider.AuthenticateVerified(context.Background(), "token"); !errors.Is(err, identity.ErrVerificationRequired) {
		t.Fatalf("unbound phone allowed: %v", err)
	}
	checker.allow = true
	got, err := provider.AuthenticateVerified(context.Background(), "token")
	if err != nil || got.Subject != "kratos-user" || checker.issuer != "primary-identity-v1" || checker.subject != "kratos-user" || checker.phone != "+923001234567" {
		t.Fatalf("binding: %+v %v checker=%+v", got, err, checker)
	}
	checker.err = errors.New("db unavailable")
	if _, err := provider.AuthenticateVerified(context.Background(), "token"); !errors.Is(err, checker.err) {
		t.Fatalf("db lookup fail-open: %v", err)
	}
}
