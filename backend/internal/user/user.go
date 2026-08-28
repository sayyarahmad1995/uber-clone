package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Capability string

const CapabilityRider Capability = "rider"

type ExternalIdentity struct {
	Issuer  string
	Subject string
}

type User struct {
	ID           uuid.UUID
	Capabilities []Capability
	CreatedAt    time.Time
}

type Repository interface {
	CreateWithDefaultRider(ctx context.Context, identity ExternalIdentity) (User, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service { return Service{repository: repository} }

func (s Service) GetOrCreate(ctx context.Context, identity ExternalIdentity) (User, error) {
	return s.repository.CreateWithDefaultRider(ctx, identity)
}
