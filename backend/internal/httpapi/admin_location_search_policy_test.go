package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

type fakeLocationSearchPolicyService struct {
	policy locationsearch.Policy
}

func (f *fakeLocationSearchPolicyService) Load(context.Context) (locationsearch.Policy, error) {
	return f.policy, nil
}

func (f *fakeLocationSearchPolicyService) Update(
	_ context.Context,
	policy locationsearch.Policy,
	actor string,
) (locationsearch.Policy, error) {
	if actor == "" || !locationsearch.ValidPolicy(policy) {
		return locationsearch.Policy{}, locationsearch.ErrInvalidPolicy
	}
	policy.UpdatedAt = time.Now().UTC()
	policy.UpdatedBy = actor
	f.policy = policy
	return policy, nil
}

type fakeMarketplacePolicyService struct {
	policy marketplace.TimingPolicy
}

func (f *fakeMarketplacePolicyService) Load(context.Context) (marketplace.TimingPolicy, error) {
	return f.policy, nil
}

func (f *fakeMarketplacePolicyService) Update(
	_ context.Context,
	policy marketplace.TimingPolicy,
	actor string,
) (marketplace.TimingPolicy, error) {
	policy.UpdatedBy = actor
	f.policy = policy
	return policy, nil
}

func adminOperationsTestAPI() (*API, *fakeLocationSearchPolicyService) {
	locationPolicy := &fakeLocationSearchPolicyService{
		policy: locationsearch.Policy{
			NamedPlaceSnapRadiusMeters: 5,
			UpdatedAt:                  time.Now().UTC(),
			UpdatedBy:                  "migration",
		},
	}
	return &API{
		marketplacePolicy: &fakeMarketplacePolicyService{
			policy: marketplace.TimingPolicy{
				RideRequestTTLSeconds:       180,
				DriverOpportunityTTLSeconds: 30,
				OfferDecisionTTLSeconds:     10,
				UpdatedAt:                   time.Now().UTC(),
				UpdatedBy:                   "migration",
			},
		},
		locationSearchPolicy: locationPolicy,
		adminReviewUsername:  "reviewer",
		adminReviewPassword:  "secret",
	}, locationPolicy
}

func TestAdminOperationsRendersLocationSearchPolicy(t *testing.T) {
	api, _ := adminOperationsTestAPI()
	request := httptest.NewRequest(http.MethodGet, "/admin/operations", nil)
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "Named-place snap radius") ||
		!strings.Contains(body, "value=\"5\"") {
		t.Fatalf("location policy was not rendered: %s", body)
	}
}

func TestAdminCanUpdateLocationSearchPolicyWithoutRestart(t *testing.T) {
	api, policy := adminOperationsTestAPI()
	request := httptest.NewRequest(
		http.MethodPost,
		"http://application.test/admin/operations/location-search-policy",
		strings.NewReader("named_place_snap_radius_meters=7"),
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "http://application.test")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d: %s", response.Code, response.Body.String())
	}
	if policy.policy.NamedPlaceSnapRadiusMeters != 7 ||
		policy.policy.UpdatedBy != "reviewer" {
		t.Fatalf("admin update was not applied/audited: %#v", policy.policy)
	}
}

func TestAdminRejectsLocationSearchPolicyOutsideBounds(t *testing.T) {
	api, policy := adminOperationsTestAPI()
	request := httptest.NewRequest(
		http.MethodPut,
		"http://application.test/v1/admin/location-search-policy",
		strings.NewReader(`{"named_place_snap_radius_meters":51}`),
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "http://application.test")
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
	if policy.policy.NamedPlaceSnapRadiusMeters != 5 {
		t.Fatalf("invalid update mutated policy: %#v", policy.policy)
	}
}
