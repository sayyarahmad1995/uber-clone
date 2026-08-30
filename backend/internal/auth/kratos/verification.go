package kratos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func (p *Provider) EnsureVerifiedSession(ctx context.Context, token string) error {
	token = bearer(token)
	if token == "" {
		return auth.ErrInvalidCredentials
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/sessions/whoami", nil)
	if err != nil {
		return auth.ErrUnavailable
	}
	req.Header.Set("X-Session-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return auth.ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return auth.ErrInvalidCredentials
	}
	if resp.StatusCode >= 300 {
		return auth.ErrUnavailable
	}

	var result struct {
		Identity struct {
			VerifiableAddresses []struct {
				Verified bool `json:"verified"`
			} `json:"verifiable_addresses"`
		} `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return auth.ErrUnavailable
	}
	for _, address := range result.Identity.VerifiableAddresses {
		if address.Verified {
			return nil
		}
	}
	return auth.ErrVerificationRequired
}
