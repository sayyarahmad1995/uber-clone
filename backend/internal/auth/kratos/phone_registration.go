package kratos

import (
	"context"
	"strings"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

// PhoneRegistration is the result of Kratos password registration, not proof
// of phone possession. Do not return or expose a session from this method.
// The caller must have independently verified phone ownership before using it.
type PhoneRegistration = auth.PhoneRegistrationIdentity

// RegisterPhoneIdentity creates a phone/password Kratos identity using the
// existing self-service API. Missing identity IDs fail closed: no email
// placeholder is fabricated, and a registration response does not imply that
// account binding was committed to PostgreSQL.
func (p *Provider) RegisterPhoneIdentity(ctx context.Context, phone, password string) (PhoneRegistration, error) {
	normalized, err := (auth.PhoneChallengeRequest{
		PhoneNumber: phone, Purpose: auth.PhoneChallengeSignInOrRegister,
	}).Validate()
	if err != nil || strings.TrimSpace(password) == "" {
		return PhoneRegistration{}, auth.ErrRegistrationInvalid
	}
	f, err := p.createFlow(ctx, "/self-service/registration/api")
	if err != nil {
		return PhoneRegistration{}, auth.ErrUnavailable
	}
	body := map[string]any{
		"method":   "password",
		"password": password,
		"traits":   map[string]string{"phone": normalized.PhoneNumber},
	}
	var result struct {
		Identity struct {
			ID string `json:"id"`
		} `json:"identity"`
	}
	if err := p.submitFlow(ctx, "/self-service/registration", f.ID, body, &result); err != nil {
		return PhoneRegistration{}, classifyRegistrationError(err)
	}
	if strings.TrimSpace(result.Identity.ID) == "" {
		return PhoneRegistration{}, auth.ErrUnavailable
	}
	return PhoneRegistration{IdentitySubject: result.Identity.ID}, nil
}
