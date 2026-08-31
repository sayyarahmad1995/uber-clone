package auth

import (
	"testing"
	"net/http"
)

func TestSessionNotExtendableUsesApplicationConflictContract(t *testing.T) {
	status, code, message := httpAuthError(ErrSessionNotExtendable)
	if status != http.StatusConflict {
		t.Fatalf("got status %d, want %d", status, http.StatusConflict)
	}
	if code != "session_not_extendable" {
		t.Fatalf("got code %q, want session_not_extendable", code)
	}
	if message != "Session cannot be extended yet." {
		t.Fatalf("unexpected message %q", message)
	}
}
