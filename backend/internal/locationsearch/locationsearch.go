package locationsearch

import (
	"context"
	"errors"
	"math"
	"strings"
)

var (
	ErrUnavailable  = errors.New("location search is unavailable")
	ErrInvalidInput = errors.New("location search input is invalid")
	ErrNotFound     = errors.New("location search result was not found")
	ErrProvider     = errors.New("location search provider failed")
)

type Point struct {
	Latitude  float64
	Longitude float64
}

func (p Point) Valid() bool {
	return !math.IsNaN(p.Latitude) &&
		!math.IsNaN(p.Longitude) &&
		!math.IsInf(p.Latitude, 0) &&
		!math.IsInf(p.Longitude, 0) &&
		p.Latitude >= -90 &&
		p.Latitude <= 90 &&
		p.Longitude >= -180 &&
		p.Longitude <= 180
}

type AutocompleteInput struct {
	Query                   string
	SessionToken            string
	Bias                    *Point
	RestrictionRadiusMeters int64
}

type Suggestion struct {
	PlaceID string
	Label   string
}

type Place struct {
	PlaceID  string
	Label    string
	Location Point
}

type Provider interface {
	Autocomplete(context.Context, AutocompleteInput) ([]Suggestion, error)
	Details(context.Context, string, string) (Place, error)
	ReverseGeocode(context.Context, Point) (Place, error)
}

type Searcher interface {
	Autocomplete(context.Context, AutocompleteInput) ([]Suggestion, error)
	Details(context.Context, string, string) (Place, error)
	ReverseGeocode(context.Context, Point) (Place, error)
}

type Service struct {
	provider Provider
	policy   SearchPolicyService
}

func NewService(provider Provider) Service {
	return Service{provider: provider}
}

func NewServiceWithPolicy(provider Provider, policy SearchPolicyService) Service {
	return Service{provider: provider, policy: policy}
}

func (s Service) Autocomplete(ctx context.Context, input AutocompleteInput) ([]Suggestion, error) {
	input.Query = strings.TrimSpace(input.Query)
	input.SessionToken = strings.TrimSpace(input.SessionToken)
	if s.provider == nil {
		return nil, ErrUnavailable
	}
	if len([]rune(input.Query)) < 2 || input.SessionToken == "" {
		return nil, ErrInvalidInput
	}
	if input.Bias != nil && !input.Bias.Valid() {
		return nil, ErrInvalidInput
	}

	input.RestrictionRadiusMeters = DefaultAutocompleteRadiusMeters
	if s.policy != nil {
		policy, err := s.policy.Load(ctx)
		if err != nil {
			return nil, err
		}
		if !ValidSearchPolicy(policy) {
			return nil, ErrProvider
		}
		input.RestrictionRadiusMeters = policy.AutocompleteRadiusMeters
	}
	return s.provider.Autocomplete(ctx, input)
}

func (s Service) Details(ctx context.Context, placeID, sessionToken string) (Place, error) {
	placeID = strings.TrimSpace(placeID)
	sessionToken = strings.TrimSpace(sessionToken)
	if s.provider == nil {
		return Place{}, ErrUnavailable
	}
	if placeID == "" {
		return Place{}, ErrInvalidInput
	}
	return s.provider.Details(ctx, placeID, sessionToken)
}

func (s Service) ReverseGeocode(ctx context.Context, point Point) (Place, error) {
	if s.provider == nil {
		return Place{}, ErrUnavailable
	}
	if !point.Valid() {
		return Place{}, ErrInvalidInput
	}
	return s.provider.ReverseGeocode(ctx, point)
}
