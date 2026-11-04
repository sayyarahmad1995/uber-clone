// Package gotrue adapts Supabase Auth's authenticated /user endpoint to
// HiGO's provider-neutral identity contract.
//
// This adapter is intentionally not wired into the running API yet. Until
// sign-up, login, recovery, verified phone binding and session restoration
// are validated together, existing Kratos authentication remains active.
package gotrue

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
)

var ErrConfig = errors.New("invalid Supabase Auth configuration")

type Provider struct {
	baseURL string
	source  string
	client  *http.Client
}

// New accepts only a trusted, origin-only URL. Do not configure an arbitrary
// user-controlled endpoint: the adapter forwards bearer tokens to this host.
// Plain HTTP is permitted only for private development/service networks;
// production must use TLS or a trusted isolated service-network transport.
func New(baseURL, identitySource string) (*Provider, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Path != "" && u.Path != "/") || strings.TrimSpace(identitySource) == "" {
		return nil, ErrConfig
	}
	return &Provider{
		baseURL: strings.TrimRight(u.String(), "/"),
		source:  strings.TrimSpace(identitySource),
		client:  &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// AuthenticateVerified asks GoTrue to validate the *presented* session and
// resolves its authoritative subject. Do not trust a locally decoded JWT or
// client-supplied phone/verification flag. A verified phone is mandatory for
// HiGO's phone-first product, including when an email is also present.
func (p *Provider) AuthenticateVerified(ctx context.Context, bearer string) (identity.Principal, error) {
	if p == nil || p.client == nil || p.baseURL == "" ||
		len(bearer) == 0 || len(bearer) > 8192 || strings.ContainsAny(bearer, " \r\n\t") {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/user", nil)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	var user struct {
		ID               string     `json:"id"`
		Phone            string     `json:"phone"`
		PhoneConfirmedAt *time.Time `json:"phone_confirmed_at"`
		BannedUntil      *time.Time `json:"banned_until"`
		IsAnonymous      bool       `json:"is_anonymous"`
	}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 64<<10))
	if err := decoder.Decode(&user); err != nil {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	id, err := uuid.Parse(user.ID)
	if err != nil || id == uuid.Nil || user.IsAnonymous {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	if user.BannedUntil != nil && user.BannedUntil.After(time.Now()) {
		return identity.Principal{}, identity.ErrInvalidToken
	}
	if user.PhoneConfirmedAt == nil || user.PhoneConfirmedAt.IsZero() ||
		!validE164(user.Phone) {
		return identity.Principal{}, identity.ErrVerificationRequired
	}
	return identity.Principal{Issuer: p.source, Subject: id.String()}, nil
}

func validE164(phone string) bool {
	if len(phone) < 9 || len(phone) > 16 || phone[0] != '+' ||
		phone[1] < '1' || phone[1] > '9' {
		return false
	}
	for i := 2; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}
	return true
}

var _ identity.Provider = (*Provider)(nil)
