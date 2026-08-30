package kratos

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestClassifyRegistrationErrorPasswordFieldPreservesMessage(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		fields:   []string{"password"},
		messages: []string{"The password was found in data breaches and must not be used."},
	}
	got := classifyRegistrationError(err)
	if !errors.Is(got, auth.ErrPasswordRejected) {
		t.Fatalf("got %v, want password rejection", got)
	}
	var publicErr *auth.PublicError
	if !errors.As(got, &publicErr) {
		t.Fatal("expected public error")
	}
	if publicErr.Code != "password_rejected" || !strings.Contains(strings.ToLower(publicErr.Message), "breach") {
		t.Fatalf("unexpected public error: %+v", publicErr)
	}
}

func TestClassifyRegistrationErrorIdentifierConflict(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		messages: []string{"An account with the same identifier exists already."},
	}
	got := classifyRegistrationError(err)
	if !errors.Is(got, auth.ErrIdentifierConflict) {
		t.Fatalf("got %v, want identifier conflict", got)
	}
	var publicErr *auth.PublicError
	if !errors.As(got, &publicErr) || publicErr.Code != "identifier_already_exists" {
		t.Fatalf("unexpected public error: %+v", publicErr)
	}
}

func TestClassifyRegistrationErrorUnknownClientFailure(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		messages: []string{"Registration rejected for the supplied values."},
	}
	got := classifyRegistrationError(err)
	if !errors.Is(got, auth.ErrRegistrationInvalid) {
		t.Fatalf("got %v, want invalid registration", got)
	}
	var publicErr *auth.PublicError
	if !errors.As(got, &publicErr) || publicErr.Code != "registration_invalid" {
		t.Fatalf("unexpected public error: %+v", publicErr)
	}
}
