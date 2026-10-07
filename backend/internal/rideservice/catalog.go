package rideservice

import "context"

// Option is the public presentation of one currently bookable ride service.
// Driver onboarding eligibility and tariff policy are intentionally separate.
type Option struct {
 PricingRequired bool
	Code              string
	DisplayName       string
	Description       string
	DisplayOrder      int
	PresentationToken string
}

type Repository interface {
	List(context.Context) ([]Option, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return Service{repository: repository}
}

func (s Service) List(ctx context.Context) ([]Option, error) {
	return s.repository.List(ctx)
}
