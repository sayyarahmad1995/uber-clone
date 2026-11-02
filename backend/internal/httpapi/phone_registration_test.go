package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestPhoneRegistrationRouteDisabledByDefault(t *testing.T) {
	api := New(Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/phone/register", strings.NewReader(`{"proof_token":"x","password":"y"}`))
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		t.Fatalf("disabled registration status=%d", response.Code)
	}
}

func TestPhoneRegistrationRouteDoesNotMountWithoutCoordinator(t *testing.T) {
	api := New(Dependencies{PhoneRegistrationEnabled: true})
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/auth/phone/register", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing coordinator must disable registration, got %d", response.Code)
	}
}

func TestPhoneRegistrationRejectsMalformedRequestBeforeAccessingServices(t *testing.T) {
	api := New(Dependencies{PhoneRegistrationEnabled: true, PhoneRegistration: &auth.PhoneRegistrationCoordinator{}})
	for _, raw := range []string{`{}`, `{"proof_token":"x","password":""}`, `{"proof_token":"x","password":"y","unexpected":true}`, `{"proof_token":"x","password":"y"}{}`} {
		response := httptest.NewRecorder()
		api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/auth/phone/register", strings.NewReader(raw)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("malformed %q returned %d", raw, response.Code)
		}
	}
}
