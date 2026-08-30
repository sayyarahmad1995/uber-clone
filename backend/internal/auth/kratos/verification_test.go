package kratos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestEnsureVerifiedSessionRejectsUnverifiedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Session-Token") != "token" {
			t.Fatalf("unexpected session token header: %q", r.Header.Get("X-Session-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"verifiable_addresses":[{"verified":false}]}}`))
	}))
	defer server.Close()

	provider := &Provider{baseURL: server.URL, client: server.Client()}
	err := provider.EnsureVerifiedSession(context.Background(), "token")
	if !errors.Is(err, auth.ErrVerificationRequired) {
		t.Fatalf("got %v, want verification required", err)
	}
}

func TestEnsureVerifiedSessionAcceptsVerifiedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"identity":{"verifiable_addresses":[{"verified":true}]}}`))
	}))
	defer server.Close()

	provider := &Provider{baseURL: server.URL, client: server.Client()}
	if err := provider.EnsureVerifiedSession(context.Background(), "Bearer token"); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}
