package kratos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestExtendSessionReturnsNotExtendableWhenKratosReturnsNotFound(t *testing.T) {
	expiresAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339Nano)
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" {
			t.Fatalf("unexpected public path %s", r.URL.Path)
		}
		if r.Header.Get("X-Session-Token") != "token" {
			t.Fatalf("unexpected session token %q", r.Header.Get("X-Session-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "session-id", "expires_at": expiresAt})
	}))
	defer public.Close()

	admin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/sessions/session-id/extend" {
			t.Fatalf("unexpected admin request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer admin.Close()

	provider := &Provider{baseURL: public.URL, adminURL: admin.URL, client: http.DefaultClient}
	_, err := provider.ExtendSession(context.Background(), "Bearer token")
	if !errors.Is(err, auth.ErrSessionNotExtendable) {
		t.Fatalf("got %v, want session not extendable", err)
	}
}

func TestExtendSessionReturnsNotExtendableWhenExpiryDoesNotAdvance(t *testing.T) {
	expiresAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339Nano)
	var whoamiCalls atomic.Int32
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		whoamiCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "session-id", "expires_at": expiresAt})
	}))
	defer public.Close()

	admin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/sessions/session-id/extend" {
			t.Fatalf("unexpected admin request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer admin.Close()

	provider := &Provider{baseURL: public.URL, adminURL: admin.URL, client: http.DefaultClient}
	_, err := provider.ExtendSession(context.Background(), "token")
	if !errors.Is(err, auth.ErrSessionNotExtendable) {
		t.Fatalf("got %v, want session not extendable", err)
	}
	if whoamiCalls.Load() != 2 {
		t.Fatalf("got %d whoami calls, want 2", whoamiCalls.Load())
	}
}

func TestExtendSessionReturnsRefreshedExpiryAfterSuccessfulExtension(t *testing.T) {
	var whoamiCalls atomic.Int32
	public := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := whoamiCalls.Add(1)
		expiresAt := time.Now().Add(10 * time.Minute)
		if call > 1 {
			expiresAt = time.Now().Add(2 * time.Hour)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"id":         "session-id",
			"expires_at": expiresAt.UTC().Format(time.RFC3339Nano),
		})
	}))
	defer public.Close()

	admin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/sessions/session-id/extend" {
			t.Fatalf("unexpected admin request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer admin.Close()

	provider := &Provider{baseURL: public.URL, adminURL: admin.URL, client: http.DefaultClient}
	session, err := provider.ExtendSession(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if session.AccessToken != "token" {
		t.Fatalf("unexpected access token %q", session.AccessToken)
	}
	if session.ExpiresIn < 7000 {
		t.Fatalf("expected refreshed expiry near two hours, got %d", session.ExpiresIn)
	}
	if whoamiCalls.Load() != 2 {
		t.Fatalf("got %d whoami calls, want 2", whoamiCalls.Load())
	}
}
