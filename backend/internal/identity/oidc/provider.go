package oidc

import (
	"context"
	"errors"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

type Provider struct {
	issuer   string
	verifier *gooidc.IDTokenVerifier
}

func New(ctx context.Context, issuer, clientID string) (*Provider, error) {
	if strings.TrimSpace(issuer) == "" {
		return nil, errors.New("OIDC issuer is required")
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New("OIDC client ID is required")
	}

	discovery, err := gooidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	return &Provider{
		issuer: issuer,
		verifier: discovery.Verifier(&gooidc.Config{ClientID: clientID}),
	}, nil
}

func (p *Provider) Authenticate(ctx context.Context, bearerToken string) (identity.Principal, error) {
	token, err := p.verifier.Verify(ctx, bearerToken)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	if token.Subject == "" || token.Issuer == "" {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	return identity.Principal{Issuer: token.Issuer, Subject: token.Subject}, nil
}
