package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeProvider struct {
	registerErr error
	loginErr    error
}

func (p fakeProvider) Register(context.Context, Credentials) (VerificationChallenge, error) {
	return VerificationChallenge{ChallengeID: "challenge"}, p.registerErr
}
func (p fakeProvider) Login(context.Context, Credentials) (Session, error) {
	return Session{AccessToken: "token", ExpiresIn: 3600}, p.loginErr
}
func (fakeProvider) Logout(context.Context, string) error { return nil }
func (fakeProvider) ExtendSession(context.Context, string) (Session, error) {
	return Session{AccessToken: "token", ExpiresIn: 3600}, nil
}
func (fakeProvider) StartVerification(context.Context, string) (VerificationChallenge, error) {
	return VerificationChallenge{ChallengeID: "challenge"}, nil
}
func (fakeProvider) CompleteVerification(context.Context, string, string) error { return nil }

func TestLoginUsesStableJSONContract(t *testing.T) {
	h := NewHandler(NewService(fakeProvider{}))
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"identifier":"a","password":"b"}`))
	w := httptest.NewRecorder()
	h.Login(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d", w.Code)
	}
	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.AccessToken != "token" || response.ExpiresIn != 3600 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestRegisterUsesApplicationVerificationContract(t *testing.T) {
	h := NewHandler(NewService(fakeProvider{}))
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"identifier":"a","password":"b"}`))
	w := httptest.NewRecorder()
	h.Register(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("got %d", w.Code)
	}
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response["verification_id"] != "challenge" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if _, leaked := response["verification_flow_id"]; leaked {
		t.Fatal("provider flow terminology leaked into public API")
	}
}

func TestProviderErrorsAreMappedByApplication(t *testing.T) {
	h := NewHandler(NewService(fakeProvider{loginErr: ErrInvalidCredentials}))
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"identifier":"a","password":"wrong"}`))
	w := httptest.NewRecorder()
	h.Login(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", w.Code)
	}

	h = NewHandler(NewService(fakeProvider{loginErr: errors.New("provider-specific failure")}))
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"identifier":"a","password":"b"}`))
	w = httptest.NewRecorder()
	h.Login(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "provider-specific") {
		t.Fatal("provider error leaked into public response")
	}
}
