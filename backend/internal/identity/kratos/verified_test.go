package kratos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func TestAuthenticateVerifiedRejectsUnverifiedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"id":"identity-1","verifiable_addresses":[{"verified":false}]}}`))
	}))
	defer server.Close()

	provider := &Provider{baseURL: server.URL, source: "primary-identity-v1", client: server.Client()}
	_, err := provider.AuthenticateVerified(context.Background(), "token")
	if !errors.Is(err, identity.ErrVerificationRequired) {
		t.Fatalf("got %v, want verification required", err)
	}
}

func TestAuthenticateVerifiedReturnsVerifiedPrincipal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"id":"identity-1","verifiable_addresses":[{"verified":true}]}}`))
	}))
	defer server.Close()

	provider := &Provider{baseURL: server.URL, source: "primary-identity-v1", client: server.Client()}
	principal, err := provider.AuthenticateVerified(context.Background(), "token")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if principal.Issuer != "primary-identity-v1" || principal.Subject != "identity-1" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}
