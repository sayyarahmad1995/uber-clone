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

type ProviderError struct {
	Status int
	Message string
}

func (e *ProviderError) Error() string { return e.Message }
func (e *ProviderError) ClientError() (int, string) { return e.Status, e.Message }

type Provider struct {
	baseURL string
	client  *http.Client
}

type flow struct {
	ID string `json:"id"`
	UI struct {
		Action string `json:"action"`
		Method string `json:"method"`
	} `json:"ui"`
}

func New(baseURL string) (*Provider, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("Kratos URL is required")
	}
	return &Provider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (p *Provider) Register(ctx context.Context, c auth.Credentials) (auth.Verification, error) {
	flow, err := p.createFlow(ctx, "/self-service/registration/api")
	if err != nil {
		return auth.Verification{}, err
	}

	body := map[string]any{
		"method":   "password",
		"password": c.Password,
		"traits": map[string]string{
			"email": c.Identifier,
		},
	}

	var result struct {
		ContinueWith []struct {
			Action string `json:"action"`
			Flow *struct {
				ID string `json:"id"`
			} `json:"flow"`
		} `json:"continue_with"`
	}
	if err := p.submitFlow(ctx, "/self-service/registration", flow.ID, body, &result); err != nil {
		return auth.Verification{}, err
	}
	for _, item := range result.ContinueWith {
		if item.Action == "show_verification_ui" && item.Flow != nil && item.Flow.ID != "" {
			return auth.Verification{FlowID: item.Flow.ID}, nil
		}
	}
	return auth.Verification{}, errors.New("Kratos registration completed without a verification flow")
}

func (p *Provider) Login(ctx context.Context, c auth.Credentials) (auth.Session, error) {
	flow, err := p.createFlow(ctx, "/self-service/login/api")
	if err != nil {
		return auth.Session{}, auth.ErrUnavailable
	}

	body := map[string]any{
		"method":     "password",
		"identifier": c.Identifier,
		"password":   c.Password,
	}

	var result struct {
		SessionToken string `json:"session_token"`
		Session      struct {
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"session"`
	}

	if err := p.submitFlow(ctx, "/self-service/login", flow.ID, body, &result); err != nil {
		return auth.Session{}, err
	}
	if result.SessionToken == "" {
		return auth.Session{}, auth.ErrInvalidCredentials
	}

	return auth.Session{
		AccessToken: result.SessionToken,
		ExpiresIn:  max(1, int64(time.Until(result.Session.ExpiresAt).Seconds())),
	}, nil
}

func (p *Provider) StartVerification(ctx context.Context, email string) (auth.Verification, error) {
	flow, err := p.createFlow(ctx, "/self-service/verification/api")
	if err != nil { return auth.Verification{}, err }
	if err := p.submitFlow(ctx, "/self-service/verification", flow.ID, map[string]any{"method":"code","email":email}, nil); err != nil {
		return auth.Verification{}, err
	}
	return auth.Verification{FlowID: flow.ID}, nil
}

func (p *Provider) CompleteVerification(ctx context.Context, flowID, code string) error {
	return p.submitFlow(ctx, "/self-service/verification", flowID, map[string]any{"method":"code","code":code}, nil)
}

func (p *Provider) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return nil
	}

	payload, err := json.Marshal(map[string]string{"session_token": token})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		p.baseURL+"/self-service/logout/api",
		bytes.NewReader(payload),
	)
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

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("Kratos logout failed: %s", resp.Status)
	}
	return nil
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
		return flow{}, fmt.Errorf("Kratos flow initialization failed: %s", resp.Status)
	}

	var result flow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return flow{}, err
	}
	if result.ID == "" {
		return flow{}, errors.New("Kratos flow has no ID")
	}
	if result.UI.Method != "" && result.UI.Method != http.MethodPost {
		return flow{}, fmt.Errorf("unsupported Kratos flow method: %s", result.UI.Method)
	}

	return result, nil
}

func (p *Provider) submitFlow(ctx context.Context, path, flowID string, body any, out any) error {
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
		var problem struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
			UI struct {
				Messages []struct {
					Text string `json:"text"`
				} `json:"messages"`
				Nodes []struct {
					Messages []struct {
						Text string `json:"text"`
					} `json:"messages"`
				} `json:"nodes"`
			} `json:"ui"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&problem)
		message := problem.Error.Message
		if message == "" && len(problem.UI.Messages) > 0 {
			message = problem.UI.Messages[0].Text
		}
		if message == "" {
			for _, node := range problem.UI.Nodes {
				if len(node.Messages) > 0 && node.Messages[0].Text != "" {
					message = node.Messages[0].Text
					break
				}
			}
		}
		if message == "" {
			message = "authentication request failed"
		}
		return &ProviderError{Status: resp.StatusCode, Message: message}
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
