package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
)

type fakePricingPolicyManager struct {
	draft  pricing.Draft
	actor  string
	policy pricing.Policy
	err    error
}

func (f *fakePricingPolicyManager) Publish(
	_ context.Context,
	draft pricing.Draft,
	actor string,
) (pricing.Policy, error) {
	f.draft = draft
	f.actor = actor
	if f.err != nil {
		return pricing.Policy{}, f.err
	}
	if f.policy.Version == 0 {
		f.policy = pricing.Policy{
			ServiceCode:        draft.ServiceCode,
			Currency:           draft.Currency,
			Version:            1,
			BaseFareMinor:      draft.BaseFareMinor,
			RateMinorPerKM:     draft.RateMinorPerKM,
			RateMinorPerMinute: draft.RateMinorPerMinute,
			MinimumFareMinor:   draft.MinimumFareMinor,
			RoundingIncrement:  draft.RoundingIncrement,
			Approved:           true,
			Active:             true,
		}
	}
	return f.policy, nil
}

func (f *fakePricingPolicyManager) ListActive(context.Context) ([]pricing.Policy, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.policy.Version == 0 {
		return nil, nil
	}
	return []pricing.Policy{f.policy}, nil
}

func TestAdminPublishPricingPolicyParsesPKRAndCreatesVersion(t *testing.T) {
	manager := &fakePricingPolicyManager{}
	api := &API{pricingPolicies: manager}
	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/operations/pricing-policy",
		strings.NewReader(
			"service_code=economy&"+
				"base_fare=100.50&"+
				"rate_per_km=50.25&"+
				"rate_per_minute=10&"+
				"minimum_fare=150&"+
				"rounding_increment=5",
		),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("owner", "secret")
	response := httptest.NewRecorder()

	api.adminPublishPricingPolicy(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d: %s", response.Code, response.Body.String())
	}
	if manager.actor != "owner" {
		t.Fatalf("unexpected actor: %q", manager.actor)
	}
	if manager.draft.ServiceCode != "economy" ||
		manager.draft.Currency != "PKR" ||
		manager.draft.BaseFareMinor != 10050 ||
		manager.draft.RateMinorPerKM != 5025 ||
		manager.draft.RateMinorPerMinute != 1000 ||
		manager.draft.MinimumFareMinor != 15000 ||
		manager.draft.RoundingIncrement != 500 {
		t.Fatalf("unexpected draft: %#v", manager.draft)
	}
}

func TestAdminPublishPricingPolicyRejectsInvalidMoney(t *testing.T) {
	manager := &fakePricingPolicyManager{}
	api := &API{pricingPolicies: manager}
	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/operations/pricing-policy",
		strings.NewReader(
			"service_code=economy&"+
				"base_fare=100&"+
				"rate_per_km=50&"+
				"rate_per_minute=10&"+
				"minimum_fare=150&"+
				"rounding_increment=0",
		),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("owner", "secret")
	response := httptest.NewRecorder()

	api.adminPublishPricingPolicy(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}
