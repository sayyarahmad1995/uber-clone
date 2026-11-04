package gotrue

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func TestAuthenticateVerifiedPhone(t *testing.T) {
	const subject = "94a05608-a1bc-4ae1-8ad3-70e2d866b511"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/user" || r.Method != http.MethodGet ||
			r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Fatalf("unexpected upstream request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + subject + `","phone":"+923001234567","phone_confirmed_at":"2026-11-01T12:00:00Z"}`))
	}))
	defer server.Close()
	provider, err := New(server.URL, "supabase-auth-v1")
	if err != nil {
		t.Fatal(err)
	}
	principal, err := provider.AuthenticateVerified(context.Background(), "valid-token")
	if err != nil {
		t.Fatal(err)
	}
	if principal.Issuer != "supabase-auth-v1" || principal.Subject != subject || requests != 1 {
		t.Fatalf("unexpected verified principal: %+v requests=%d", principal, requests)
	}
}

func TestAuthenticateRejectsInvalidOrUnverifiedPhone(t *testing.T) {
	cases := []struct {
		name     string
		code     int
		payload  string
		expected error
	}{
		{"unauthorized", http.StatusUnauthorized, `{"message":"bad token"}`, identity.ErrInvalidToken},
		{"empty identity", http.StatusOK, `{"id":""}`, identity.ErrInvalidToken},
		{"missing phone", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","phone_confirmed_at":"2026-11-01T12:00:00Z"}`, identity.ErrVerificationRequired},
		{"unverified", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","phone":"+923001234567","phone_confirmed_at":null}`, identity.ErrVerificationRequired},
		{"unverified email only", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","email_confirmed_at":"2026-11-01T12:00:00Z","phone":"+923001234567"}`, identity.ErrVerificationRequired},
		{"malformed phone", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","phone":"03001234567","phone_confirmed_at":"2026-11-01T12:00:00Z"}`, identity.ErrVerificationRequired},
		{"anonymous", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","phone":"+923001234567","phone_confirmed_at":"2026-11-01T12:00:00Z","is_anonymous":true}`, identity.ErrInvalidToken},
		{"future ban", http.StatusOK, `{"id":"94a05608-a1bc-4ae1-8ad3-70e2d866b511","phone":"+923001234567","phone_confirmed_at":"2026-11-01T12:00:00Z","banned_until":"2099-01-01T00:00:00Z"}`, identity.ErrInvalidToken},
		{"bad JSON", http.StatusOK, "{", identity.ErrInvalidToken},
		{"oversized body", http.StatusOK, strings.Repeat("x", 70000), identity.ErrInvalidToken},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.payload))
			}))
			defer server.Close()
			provider, err := New(server.URL, "supabase-auth-v1")
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.AuthenticateVerified(context.Background(), "valid-token")
			if !errors.Is(err, tc.expected) {
				t.Fatalf("got %v, want %v", err, tc.expected)
			}
		})
	}
}

func TestRejectsMalformedConfigAndTokens(t *testing.T) {
	for _, origin := range []string{"", "localhost", "file:///tmp/x", "http://user:pass@localhost:9999", "https://localhost/path", "http://localhost?x=1"} {
		if _, err := New(origin, "supabase-auth-v1"); !errors.Is(err, ErrConfig) {
			t.Errorf("accepted origin %q: %v", origin, err)
		}
	}
	if _, err := New("http://localhost:9999", ""); !errors.Is(err, ErrConfig) {
		t.Fatal("accepted missing identity source")
	}
	provider, err := New("http://127.0.0.1:1", "supabase-auth-v1")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	for _, token := range []string{"", "Bearer foo", "line\nbreak", strings.Repeat("x", 8193)} {
		if _, err := provider.AuthenticateVerified(ctx, token); !errors.Is(err, identity.ErrInvalidToken) {
			t.Errorf("accepted invalid token %q: %v", token[:min(len(token), 24)], err)
		}
	}
}
