package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeProvider struct{}

func (fakeProvider) Register(context.Context, Credentials) (Verification, error) {
	return Verification{FlowID: "flow"}, nil
}
func (fakeProvider) Login(context.Context, Credentials) (Session, error) {
	return Session{AccessToken: "token", ExpiresIn: 3600}, nil
}
func (fakeProvider) Logout(context.Context, string) error { return nil }
func (fakeProvider) StartVerification(context.Context, string) (Verification, error) {
	return Verification{FlowID: "flow"}, nil
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
