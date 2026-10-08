package pricing

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"time"
)

const MaxAmountMinor int64 = 1_000_000_000_000
const CalculationRule = "half_up_once_v1"

var (
	ErrInvalidPolicy  = errors.New("invalid pricing policy")
	ErrInvalidRoute   = errors.New("invalid route metrics")
	ErrFareOutOfRange = errors.New("suggested fare is out of range")
	ErrUnavailable    = errors.New("pricing policy unavailable")
	ErrPolicyChanged  = errors.New("pricing policy changed")
	ErrInvalidService = errors.New("invalid pricing service")
)

type Draft struct {
	ServiceCode            string
	Currency               string
	BaseFareMinor          int64
	RateMinorPerKm         int64
	RateMinorPerMinute     int64
	MinimumFareMinor       int64
	RoundingIncrementMinor int64
}
type Policy struct {
	ID                     uuid.UUID
	ServiceCode            string
	Currency               string
	Version                int64
	BaseFareMinor          int64
	RateMinorPerKm         int64
	RateMinorPerMinute     int64
	MinimumFareMinor       int64
	RoundingIncrementMinor int64
	CalculationRule        string
	EffectiveFrom          time.Time
	PublishedAt            time.Time
	PublishedBy            string
}

func (p Policy) Draft() Draft {
	return Draft{p.ServiceCode, p.Currency, p.BaseFareMinor, p.RateMinorPerKm, p.RateMinorPerMinute, p.MinimumFareMinor, p.RoundingIncrementMinor}
}
func Validate(p Draft) error {
	if p.ServiceCode == "" || p.Currency != "PKR" || p.MinimumFareMinor <= 0 || p.RoundingIncrementMinor <= 0 {
		return ErrInvalidPolicy
	}
	for _, v := range []int64{p.BaseFareMinor, p.RateMinorPerKm, p.RateMinorPerMinute, p.MinimumFareMinor, p.RoundingIncrementMinor} {
		if v < 0 || v > MaxAmountMinor {
			return ErrInvalidPolicy
		}
	}
	return nil
}

type Repository interface {
	Current(context.Context, string, string) (Policy, error)
	Publish(context.Context, Draft, string) (Policy, error)
	Disable(context.Context, string, string, uuid.UUID, string) error
	List(context.Context) ([]Policy, error)
}
