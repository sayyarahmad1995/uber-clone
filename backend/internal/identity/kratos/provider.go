package kratos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

type Provider struct {
	baseURL string
	source  string
	client  *http.Client
}

func New(baseURL, source string) (*Provider, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("Kratos URL is required")
	}
	if strings.TrimSpace(source) == "" {
		return nil, errors.New("identity source is required")
	}
	return &Provider{
		baseURL: strings.TrimRight(baseURL, "/"),
		source:  strings.TrimSpace(source),
		client:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *Provider) Authenticate(ctx context.Context, token string) (identity.Principal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/sessions/whoami", nil)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	req.Header.Set("X-Session-Token", token)
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
			ID string `json:"id"`
		} `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil || session.Identity.ID == "" {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	return identity.Principal{Issuer: p.source, Subject: session.Identity.ID}, nil
}
