package kratos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

// VerifyPhoneIdentity performs a fresh privileged Kratos identity lookup.
// It must only be used after operator identification of the exact subject;
// it does not search by phone or establish whether duplicate identities exist.
func (p *Provider) VerifyPhoneIdentity(ctx context.Context, subject, phone string) (bool, error) {
	if subject == "" || phone == "" {
		return false, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.adminURL+"/admin/identities/"+url.PathEscape(subject), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, auth.ErrUnavailable
	}
	var result struct {
		ID     string `json:"id"`
		Traits struct {
			Phone string `json:"phone"`
		} `json:"traits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.ID == subject && result.Traits.Phone == phone, nil
}
