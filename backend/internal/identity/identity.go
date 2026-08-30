package identity

import (
	"context"
	"errors"
)

var (
	ErrUnauthenticated      = errors.New("identity is unauthenticated")
	ErrInvalidToken         = errors.New("identity token is invalid")
	ErrVerificationRequired = errors.New("identity verification is required")
)

type Principal struct {
	Issuer  string
	Subject string
}

// Provider is the provider-neutral application boundary for external identity.
// It returns a principal only when the external credential is valid and the
// identity satisfies the application's verification policy.
type Provider interface {
	AuthenticateVerified(ctx context.Context, bearerToken string) (Principal, error)
}
