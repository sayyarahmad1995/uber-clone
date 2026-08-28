package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Capability string

const (
	CapabilityRider Capability = "rider"
)

type User struct {
	ID              uuid.UUID
	ExternalSubject string
	Capabilities    []Capability
	CreatedAt       time.Time
}

type Repository interface {
	CreateWithDefaultRider(ctx context.Context, subject string) (User, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) GetOrCreate(ctx context.Context, externalSubject string) (User, error) {
	return s.repository.CreateWithDefaultRider(ctx, externalSubject)
}
