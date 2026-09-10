package kratos

import (
	"context"
	"encoding/json"
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

func TestCompleteVerificationRequiresPassedChallenge(t *testing.T) {
	for _, tc := range []struct {
		name    string
		state   string
		wantErr error
	}{
		{"passed", "passed_challenge", nil},
		{"still active", "sent_email", auth.ErrVerificationInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/self-service/verification" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Query().Get("flow") != "challenge" {
					t.Fatalf("unexpected flow %q", r.URL.Query().Get("flow"))
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"state": tc.state})
			}))
			defer server.Close()

			provider := &Provider{baseURL: server.URL, client: server.Client()}
			err := provider.CompleteVerification(context.Background(), "challenge", "123456")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
