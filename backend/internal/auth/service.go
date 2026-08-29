package auth

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnavailable = errors.New("authentication service unavailable")
)

type Credentials struct {
	Identifier string
	Password string
}

type Verification struct { FlowID string }

type Session struct {
	AccessToken string
	RefreshToken string
	ExpiresIn int64
}

type IdentityProvider interface {
	Register(context.Context, Credentials) (Verification, error)
	Login(context.Context, Credentials) (Session, error)
	Refresh(context.Context, string) (Session, error)
	Logout(context.Context, string) error
	StartVerification(context.Context, string) (Verification, error)
	CompleteVerification(context.Context, string, string) error
}

type Service struct { provider IdentityProvider }

func NewService(provider IdentityProvider) Service { return Service{provider: provider} }
func (s Service) Register(ctx context.Context, c Credentials) (Verification,error) { return s.provider.Register(ctx,c) }
func (s Service) Login(ctx context.Context, c Credentials) (Session,error) { return s.provider.Login(ctx,c) }
func (s Service) Refresh(ctx context.Context, token string) (Session,error) { return s.provider.Refresh(ctx,token) }
func (s Service) Logout(ctx context.Context, token string) error { return s.provider.Logout(ctx,token) }
func (s Service) StartVerification(ctx context.Context, email string) (Verification,error) { return s.provider.StartVerification(ctx,email) }
func (s Service) CompleteVerification(ctx context.Context, flowID, code string) error { return s.provider.CompleteVerification(ctx,flowID,code) }
