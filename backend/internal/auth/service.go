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

type Session struct {
	AccessToken string
	RefreshToken string
	ExpiresIn int64
}

type IdentityProvider interface {
	Register(context.Context, Credentials) error
	Login(context.Context, Credentials) (Session, error)
	Refresh(context.Context, string) (Session, error)
	Logout(context.Context, string) error
	Verify(context.Context, string) error
}

type Service struct { provider IdentityProvider }

func NewService(provider IdentityProvider) Service { return Service{provider: provider} }
func (s Service) Register(ctx context.Context, c Credentials) error { return s.provider.Register(ctx,c) }
func (s Service) Login(ctx context.Context, c Credentials) (Session,error) { return s.provider.Login(ctx,c) }
func (s Service) Refresh(ctx context.Context, token string) (Session,error) { return s.provider.Refresh(ctx,token) }
func (s Service) Logout(ctx context.Context, token string) error { return s.provider.Logout(ctx,token) }
func (s Service) Verify(ctx context.Context, email string) error { return s.provider.Verify(ctx,email) }
