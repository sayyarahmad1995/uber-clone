package pricing

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

type fakeRepository struct {
	policy Policy
	err    error
}

func (f fakeRepository) ActivePolicy(
	context.Context,
	string,
	string,
	time.Time,
) (Policy, error) {
	return f.policy, f.err
}

func approvedPolicy() Policy {
	return Policy{
		ServiceCode:        "test",
		Currency:           "PKR",
		Version:            3,
		BaseFareMinor:      10000,
		RateMinorPerKM:     5000,
		RateMinorPerMinute: 1000,
		MinimumFareMinor:   15000,
		RoundingIncrement:  500,
		EffectiveFrom:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Approved:           true,
		Active:             true,
	}
}

func TestSuggestUsesOneFinalHalfUpRoundingStep(t *testing.T) {
	service := NewService(fakeRepository{policy: approvedPolicy()})
	service.now = func() time.Time {
		return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	}

	suggestion, err := service.Suggest(context.Background(), "test", 1500, 150)
	if err != nil {
		t.Fatal(err)
	}
	// 100.00 + 75.00 + 25.00 = 200.00 PKR, rounded to 5.00 PKR.
	if suggestion.Fare.AmountMinor != 20000 {
		t.Fatalf("expected 20000, got %d", suggestion.Fare.AmountMinor)
	}
	if suggestion.Fare.Currency != "PKR" || suggestion.PolicyVersion != 3 {
		t.Fatalf("unexpected suggestion: %#v", suggestion)
	}
}

func TestSuggestAppliesMinimumAfterRounding(t *testing.T) {
	policy := approvedPolicy()
	policy.MinimumFareMinor = 25000
	service := NewService(fakeRepository{policy: policy})
	service.now = func() time.Time {
		return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	}

	suggestion, err := service.Suggest(context.Background(), "test", 100, 60)
	if err != nil {
		t.Fatal(err)
	}
	if suggestion.Fare.AmountMinor != 25000 {
		t.Fatalf("expected minimum fare, got %d", suggestion.Fare.AmountMinor)
	}
}

func TestCalculateRoundsHalfUpOnlyAtEnd(t *testing.T) {
	policy := approvedPolicy()
	policy.BaseFareMinor = 0
	policy.RateMinorPerKM = 1
	policy.RateMinorPerMinute = 0
	policy.MinimumFareMinor = 1
	policy.RoundingIncrement = 1

	amount, err := calculate(policy, 500, 60)
	if err != nil {
		t.Fatal(err)
	}
	if amount != 1 {
		t.Fatalf("expected half-up result 1, got %d", amount)
	}
}

func TestSuggestRejectsUnavailableOrInvalidPolicy(t *testing.T) {
	service := NewService(fakeRepository{err: ErrPolicyUnavailable})
	_, err := service.Suggest(context.Background(), "test", 1000, 60)
	if !errors.Is(err, ErrPolicyUnavailable) {
		t.Fatalf("expected unavailable policy, got %v", err)
	}

	policy := approvedPolicy()
	policy.Approved = false
	service = NewService(fakeRepository{policy: policy})
	service.now = func() time.Time {
		return time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	}
	_, err = service.Suggest(context.Background(), "test", 1000, 60)
	if !errors.Is(err, ErrPolicyInvalid) {
		t.Fatalf("expected invalid policy, got %v", err)
	}
}

func TestCalculateDetectsOverflow(t *testing.T) {
	policy := approvedPolicy()
	policy.BaseFareMinor = math.MaxInt64
	policy.RateMinorPerKM = math.MaxInt64
	policy.RateMinorPerMinute = math.MaxInt64
	policy.MinimumFareMinor = 1
	policy.RoundingIncrement = 1

	_, err := calculate(policy, math.MaxInt64, math.MaxInt64)
	if !errors.Is(err, ErrOverflow) {
		t.Fatalf("expected overflow, got %v", err)
	}
}
