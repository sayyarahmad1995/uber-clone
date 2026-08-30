package oidc

import (
	"context"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func (p *Provider) AuthenticateVerified(ctx context.Context, bearerToken string) (identity.Principal, error) {
	token, err := p.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	if token.Subject == "" || token.Issuer == "" {
		return identity.Principal{}, identity.ErrInvalidToken
	}

	var claims struct {
		EmailVerified bool `json:"email_verified"`
	}
	if err := token.Claims(&claims); err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	if !claims.EmailVerified {
		return identity.Principal{}, identity.ErrVerificationRequired
	}

	return identity.Principal{Issuer: token.Issuer, Subject: token.Subject}, nil
}
