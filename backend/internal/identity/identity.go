package identity

import (
	"context"
	"errors"
)

var (
	ErrUnauthenticated = errors.New("identity is unauthenticated")
	ErrInvalidToken    = errors.New("identity token is invalid")
)

type Principal struct {
	Issuer  string
	Subject string
}

// Provider is the provider-neutral application boundary for external OIDC identity.
type Provider interface {
	Authenticate(ctx context.Context, bearerToken string) (Principal, error)
}
