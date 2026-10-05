package pricing

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"
)

var (
	ErrInvalidInput      = errors.New("pricing input is invalid")
	ErrPolicyUnavailable = errors.New("pricing policy is unavailable")
	ErrPolicyInvalid     = errors.New("pricing policy is invalid")
	ErrOverflow          = errors.New("pricing calculation overflow")
)

const defaultCurrency = "PKR"

type Policy struct {
	ServiceCode        string     `json:"service_code"`
	Currency           string     `json:"currency"`
	Version            int64      `json:"version"`
	BaseFareMinor      int64      `json:"base_fare_minor"`
	RateMinorPerKM     int64      `json:"rate_minor_per_km"`
	RateMinorPerMinute int64      `json:"rate_minor_per_minute"`
	MinimumFareMinor   int64      `json:"minimum_fare_minor"`
	RoundingIncrement  int64      `json:"rounding_increment_minor"`
	EffectiveFrom      time.Time  `json:"effective_from"`
	EffectiveUntil     *time.Time `json:"effective_until,omitempty"`
	Approved           bool       `json:"approved"`
	Active             bool       `json:"active"`
}

func (p Policy) Valid(at time.Time) bool {
	service := strings.TrimSpace(p.ServiceCode)
	currency := strings.ToUpper(strings.TrimSpace(p.Currency))
	if service == "" || len(currency) != 3 || p.Version <= 0 ||
		p.BaseFareMinor < 0 ||
		p.RateMinorPerKM < 0 ||
		p.RateMinorPerMinute < 0 ||
		p.MinimumFareMinor <= 0 ||
		p.RoundingIncrement <= 0 ||
		p.EffectiveFrom.IsZero() ||
		!p.Approved ||
		!p.Active {
		return false
	}
	if at.Before(p.EffectiveFrom) {
		return false
	}
	if p.EffectiveUntil != nil && !at.Before(*p.EffectiveUntil) {
		return false
	}
	return true
}

type Fare struct {
	AmountMinor int64
	Currency    string
}

type Suggestion struct {
	Fare          Fare
	PolicyVersion int64
}

type Repository interface {
	ActivePolicy(context.Context, string, string, time.Time) (Policy, error)
}

type Suggester interface {
	Suggest(context.Context, string, int64, int64) (Suggestion, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) Service {
	return Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s Service) Suggest(
	ctx context.Context,
	serviceCode string,
	distanceMeters int64,
	durationSeconds int64,
) (Suggestion, error) {
	serviceCode = strings.ToLower(strings.TrimSpace(serviceCode))
	if s.repository == nil || serviceCode == "" || distanceMeters <= 0 || durationSeconds <= 0 {
		return Suggestion{}, ErrInvalidInput
	}

	at := s.now().UTC()
	policy, err := s.repository.ActivePolicy(ctx, serviceCode, defaultCurrency, at)
	if err != nil {
		return Suggestion{}, err
	}
	if !policy.Valid(at) ||
		strings.ToLower(strings.TrimSpace(policy.ServiceCode)) != serviceCode ||
		strings.ToUpper(strings.TrimSpace(policy.Currency)) != defaultCurrency {
		return Suggestion{}, ErrPolicyInvalid
	}

	amount, err := calculate(policy, distanceMeters, durationSeconds)
	if err != nil {
		return Suggestion{}, err
	}
	return Suggestion{
		Fare: Fare{
			AmountMinor: amount,
			Currency:    defaultCurrency,
		},
		PolicyVersion: policy.Version,
	}, nil
}

// calculate evaluates the rational formula exactly and performs one final
// half-up rounding step to the policy increment. No binary floating point is
// used for money.
func calculate(policy Policy, distanceMeters, durationSeconds int64) (int64, error) {
	if distanceMeters <= 0 || durationSeconds <= 0 {
		return 0, ErrInvalidInput
	}

	// LCM(1000 meters/km, 60 seconds/minute) = 3000.
	const denominator int64 = 3000
	numerator := big.NewInt(0)
	numerator.Add(numerator, new(big.Int).Mul(
		big.NewInt(policy.BaseFareMinor),
		big.NewInt(denominator),
	))
	numerator.Add(numerator, new(big.Int).Mul(
		new(big.Int).Mul(big.NewInt(policy.RateMinorPerKM), big.NewInt(distanceMeters)),
		big.NewInt(3),
	))
	numerator.Add(numerator, new(big.Int).Mul(
		new(big.Int).Mul(big.NewInt(policy.RateMinorPerMinute), big.NewInt(durationSeconds)),
		big.NewInt(50),
	))

	incrementDenominator := new(big.Int).Mul(
		big.NewInt(denominator),
		big.NewInt(policy.RoundingIncrement),
	)
	units, remainder := new(big.Int), new(big.Int)
	units.QuoRem(numerator, incrementDenominator, remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(incrementDenominator) >= 0 {
		units.Add(units, big.NewInt(1))
	}

	rounded := new(big.Int).Mul(units, big.NewInt(policy.RoundingIncrement))
	minimum := big.NewInt(policy.MinimumFareMinor)
	if rounded.Cmp(minimum) < 0 {
		rounded.Set(minimum)
	}
	if !rounded.IsInt64() {
		return 0, ErrOverflow
	}
	return rounded.Int64(), nil
}
