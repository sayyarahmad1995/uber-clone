package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

type Provider struct {
	baseURL  string
	adminURL string
	client   *http.Client
}

type flow struct {
	ID string `json:"id"`
	UI struct {
		Method string `json:"method"`
	} `json:"ui"`
}

func New(baseURL, adminURL string) (*Provider, error) {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(adminURL) == "" {
		return nil, errors.New("Kratos public and admin URLs are required")
	}
	return &Provider{
		baseURL:  strings.TrimRight(baseURL, "/"),
		adminURL: strings.TrimRight(adminURL, "/"),
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *Provider) Register(ctx context.Context, c auth.Credentials) (auth.VerificationChallenge, error) {
	f, err := p.createFlow(ctx, "/self-service/registration/api")
	if err != nil {
		return auth.VerificationChallenge{}, auth.ErrUnavailable
	}
	body := map[string]any{
		"method":   "password",
		"password": c.Password,
		"traits":   map[string]string{"email": c.Identifier},
	}
	var result struct {
		ContinueWith []struct {
			Action string `json:"action"`
			Flow   *struct {
				ID string `json:"id"`
			} `json:"flow"`
		} `json:"continue_with"`
	}
	if err := p.submitFlow(ctx, "/self-service/registration", f.ID, body, &result); err != nil {
		return auth.VerificationChallenge{}, classifyRegistrationError(err)
	}
	for _, item := range result.ContinueWith {
		if item.Action == "show_verification_ui" && item.Flow != nil && item.Flow.ID != "" {
			return auth.VerificationChallenge{ChallengeID: item.Flow.ID}, nil
		}
	}
	return auth.VerificationChallenge{}, auth.ErrUnavailable
}

func (p *Provider) Login(ctx context.Context, c auth.Credentials) (auth.Session, error) {
	f, err := p.createFlow(ctx, "/self-service/login/api")
	if err != nil {
		return auth.Session{}, auth.ErrUnavailable
	}
	body := map[string]any{"method": "password", "identifier": c.Identifier, "password": c.Password}
	var result struct {
		SessionToken string `json:"session_token"`
		Session      struct {
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"session"`
	}
	if err := p.submitFlow(ctx, "/self-service/login", f.ID, body, &result); err != nil {
		if isClientFailure(err) {
			return auth.Session{}, auth.ErrInvalidCredentials
		}
		return auth.Session{}, auth.ErrUnavailable
	}
	if result.SessionToken == "" || result.Session.ExpiresAt.IsZero() {
		return auth.Session{}, auth.ErrInvalidCredentials
	}
	return auth.Session{AccessToken: result.SessionToken, ExpiresIn: max(1, int64(time.Until(result.Session.ExpiresAt).Seconds()))}, nil
}

func (p *Provider) StartVerification(ctx context.Context, email string) (auth.VerificationChallenge, error) {
	f, err := p.createFlow(ctx, "/self-service/verification/api")
	if err != nil {
		return auth.VerificationChallenge{}, auth.ErrUnavailable
	}
	if err := p.submitFlow(ctx, "/self-service/verification", f.ID, map[string]any{"method": "code", "email": email}, nil); err != nil {
		if isClientFailure(err) {
			return auth.VerificationChallenge{}, auth.ErrVerificationInvalid
		}
		return auth.VerificationChallenge{}, auth.ErrUnavailable
	}
	return auth.VerificationChallenge{ChallengeID: f.ID}, nil
}

func (p *Provider) CompleteVerification(ctx context.Context, challengeID, code string) error {
	err := p.submitFlow(ctx, "/self-service/verification", challengeID, map[string]any{"method": "code", "code": code}, nil)
	if err == nil {
		return nil
	}
	if isClientFailure(err) {
		return auth.ErrVerificationInvalid
	}
	return auth.ErrUnavailable
}

func (p *Provider) ExtendSession(ctx context.Context, token string) (auth.Session, error) {
	token = bearer(token)
	if token == "" {
		return auth.Session{}, auth.ErrInvalidCredentials
	}
	current, err := p.session(ctx, token)
	if err != nil {
		return auth.Session{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, p.adminURL+"/sessions/"+url.PathEscape(current.ID)+"/extend", nil)
	if err != nil {
		return auth.Session{}, auth.ErrUnavailable
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return auth.Session{}, auth.ErrUnavailable
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		return auth.Session{}, auth.ErrUnavailable
	}
	updated := current
	if resp.StatusCode < 300 {
		updated, err = p.session(ctx, token)
		if err != nil {
			return auth.Session{}, err
		}
	}
	return auth.Session{AccessToken: token, ExpiresIn: max(1, int64(time.Until(updated.ExpiresAt).Seconds()))}, nil
}

func (p *Provider) Logout(ctx context.Context, token string) error {
	token = bearer(token)
	if token == "" {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{"session_token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p.baseURL+"/self-service/logout/api", bytes.NewReader(payload))
	if err != nil {
		return auth.ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return auth.ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return auth.ErrInvalidCredentials
	}
	return auth.ErrUnavailable
}

type kratosSession struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (p *Provider) session(ctx context.Context, token string) (kratosSession, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/sessions/whoami", nil)
	if err != nil {
		return kratosSession{}, auth.ErrUnavailable
	}
	req.Header.Set("X-Session-Token", token)
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return kratosSession{}, auth.ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return kratosSession{}, auth.ErrInvalidCredentials
	}
	if resp.StatusCode >= 300 {
		return kratosSession{}, auth.ErrUnavailable
	}
	var session kratosSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil || session.ID == "" || session.ExpiresAt.IsZero() {
		return kratosSession{}, auth.ErrUnavailable
	}
	return session, nil
}

func (p *Provider) createFlow(ctx context.Context, path string) (flow, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
	if err != nil {
		return flow{}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return flow{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return flow{}, fmt.Errorf("flow initialization failed: %s", resp.Status)
	}
	var result flow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return flow{}, err
	}
	if result.ID == "" || (result.UI.Method != "" && result.UI.Method != http.MethodPost) {
		return flow{}, errors.New("provider returned invalid flow")
	}
	return result, nil
}

type providerResponseError struct {
	status   int
	fields   []string
	messages []string
}

func (e providerResponseError) Error() string {
	return fmt.Sprintf("provider request failed: %d", e.status)
}

func isClientFailure(err error) bool {
	var providerErr providerResponseError
	return errors.As(err, &providerErr) && providerErr.status >= 400 && providerErr.status < 500
}

func classifyRegistrationError(err error) error {
	var providerErr providerResponseError
	if !errors.As(err, &providerErr) {
		return auth.ErrUnavailable
	}
	if providerErr.status < 400 || providerErr.status >= 500 {
		return auth.ErrUnavailable
	}

	for _, field := range providerErr.fields {
		if strings.Contains(strings.ToLower(field), "password") {
			return auth.ErrPasswordRejected
		}
	}

	text := strings.ToLower(strings.Join(providerErr.messages, " "))
	if strings.Contains(text, "password") && (
		strings.Contains(text, "breach") ||
		strings.Contains(text, "pwn") ||
		strings.Contains(text, "weak") ||
		strings.Contains(text, "short") ||
		strings.Contains(text, "minimum") ||
		strings.Contains(text, "requirement")) {
		return auth.ErrPasswordRejected
	}
	if strings.Contains(text, "already") && (
		strings.Contains(text, "exist") ||
		strings.Contains(text, "used") ||
		strings.Contains(text, "taken") ||
		strings.Contains(text, "registered")) {
		return auth.ErrIdentifierConflict
	}

	return auth.ErrRegistrationInvalid
}

func (p *Provider) submitFlow(ctx context.Context, path, flowID string, body, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	endpoint, err := url.Parse(p.baseURL + path)
	if err != nil {
		return err
	}
	query := endpoint.Query()
	query.Set("flow", flowID)
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		providerErr := providerResponseError{status: resp.StatusCode}
		var problem struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
			UI struct {
				Messages []struct {
					Text string `json:"text"`
				} `json:"messages"`
				Nodes []struct {
					Attributes struct {
						Name string `json:"name"`
					} `json:"attributes"`
					Messages []struct {
						Text string `json:"text"`
					} `json:"messages"`
				} `json:"nodes"`
			} `json:"ui"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&problem); err == nil {
			if strings.TrimSpace(problem.Error.Message) != "" {
				providerErr.messages = append(providerErr.messages, problem.Error.Message)
			}
			for _, message := range problem.UI.Messages {
				if strings.TrimSpace(message.Text) != "" {
					providerErr.messages = append(providerErr.messages, message.Text)
				}
			}
			for _, node := range problem.UI.Nodes {
				if strings.TrimSpace(node.Attributes.Name) != "" && len(node.Messages) > 0 {
					providerErr.fields = append(providerErr.fields, node.Attributes.Name)
				}
				for _, message := range node.Messages {
					if strings.TrimSpace(message.Text) != "" {
						providerErr.messages = append(providerErr.messages, message.Text)
					}
				}
			}
		}
		return providerErr
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func bearer(token string) string {
	return strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
