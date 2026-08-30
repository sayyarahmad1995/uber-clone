package auth

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrIdentifierConflict  = errors.New("identifier already exists")
	ErrPasswordRejected    = errors.New("password does not meet requirements")
	ErrRegistrationInvalid = errors.New("registration request is invalid")
	ErrVerificationInvalid = errors.New("verification is invalid or expired")
	ErrUnavailable         = errors.New("authentication service unavailable")
)

type Credentials struct {
	Identifier string
	Password   string
}

// VerificationChallenge is an application-owned verification concept.
// ChallengeID is opaque to callers; provider adapters may map it to their own
// transaction, flow, or challenge identifier internally.
type VerificationChallenge struct {
	ChallengeID string
}

type Session struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Provider is the application-owned authentication port. Implementations must
// translate provider-specific concepts and errors into these types and errors.
type Provider interface {
	Register(context.Context, Credentials) (VerificationChallenge, error)
	Login(context.Context, Credentials) (Session, error)
	Logout(context.Context, string) error
	ExtendSession(context.Context, string) (Session, error)
	StartVerification(context.Context, string) (VerificationChallenge, error)
	CompleteVerification(context.Context, string, string) error
}

type Service struct{ provider Provider }

func NewService(provider Provider) Service { return Service{provider: provider} }
func (s Service) Register(ctx context.Context, c Credentials) (VerificationChallenge, error) {
	return s.provider.Register(ctx, c)
}
func (s Service) Login(ctx context.Context, c Credentials) (Session, error) {
	return s.provider.Login(ctx, c)
}
func (s Service) Logout(ctx context.Context, token string) error {
	return s.provider.Logout(ctx, token)
}
func (s Service) ExtendSession(ctx context.Context, token string) (Session, error) {
	return s.provider.ExtendSession(ctx, token)
}
func (s Service) StartVerification(ctx context.Context, email string) (VerificationChallenge, error) {
	return s.provider.StartVerification(ctx, email)
}
func (s Service) CompleteVerification(ctx context.Context, challengeID, code string) error {
	return s.provider.CompleteVerification(ctx, challengeID, code)
}
