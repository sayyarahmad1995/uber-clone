package kratos

import (
	"errors"
	"net/http"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestClassifyRegistrationErrorPasswordField(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		fields:   []string{"password"},
		messages: []string{"The password was found in data breaches."},
	}
	if got := classifyRegistrationError(err); !errors.Is(got, auth.ErrPasswordRejected) {
		t.Fatalf("got %v, want password rejection", got)
	}
}

func TestClassifyRegistrationErrorIdentifierConflict(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		messages: []string{"An account with the same identifier exists already."},
	}
	if got := classifyRegistrationError(err); !errors.Is(got, auth.ErrIdentifierConflict) {
		t.Fatalf("got %v, want identifier conflict", got)
	}
}

func TestClassifyRegistrationErrorUnknownClientFailure(t *testing.T) {
	err := providerResponseError{
		status:   http.StatusBadRequest,
		messages: []string{"registration rejected"},
	}
	if got := classifyRegistrationError(err); !errors.Is(got, auth.ErrRegistrationInvalid) {
		t.Fatalf("got %v, want invalid registration", got)
	}
}
