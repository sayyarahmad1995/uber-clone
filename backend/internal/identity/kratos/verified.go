package kratos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

func (p *Provider) AuthenticateVerified(ctx context.Context, token string) (identity.Principal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/sessions/whoami", nil)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	req.Header.Set("X-Session-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return identity.Principal{}, identity.ErrInvalidToken
	}

	var session struct {
		Identity struct {
			ID                  string `json:"id"`
			VerifiableAddresses []struct {
				Verified bool `json:"verified"`
			} `json:"verifiable_addresses"`
		} `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil || session.Identity.ID == "" {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	for _, address := range session.Identity.VerifiableAddresses {
		if address.Verified {
			return identity.Principal{Issuer: p.source, Subject: session.Identity.ID}, nil
		}
	}
	return identity.Principal{}, identity.ErrVerificationRequired
}
