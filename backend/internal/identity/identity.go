package identity

import "context"

type Principal struct{ Subject string }

// Provider is the application boundary to the selected external OIDC provider.
type Provider interface {
	Authenticate(ctx context.Context, bearerToken string) (Principal, error)
}
