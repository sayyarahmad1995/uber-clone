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

type fakePlaceSearchPolicyService struct {
	policy locationsearch.SearchPolicy
}

func (f *fakePlaceSearchPolicyService) Load(context.Context) (locationsearch.SearchPolicy, error) {
	return f.policy, nil
}

func (f *fakePlaceSearchPolicyService) Update(
	_ context.Context,
	policy locationsearch.SearchPolicy,
	actor string,
) (locationsearch.SearchPolicy, error) {
	if !locationsearch.ValidSearchPolicy(policy) {
		return locationsearch.SearchPolicy{}, locationsearch.ErrInvalidSearchPolicy
	}
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

func TestAdminOperationsNoLongerExposesLocationSnapControls(t *testing.T) {
	api := &API{
		marketplacePolicy: &fakeMarketplacePolicyService{
			policy: marketplace.TimingPolicy{
				RideRequestTTLSeconds:       180,
				DriverOpportunityTTLSeconds: 30,
				OfferDecisionTTLSeconds:     10,
				UpdatedAt:                   time.Now().UTC(),
				UpdatedBy:                   "migration",
			},
		},
		locationSearchPolicy: &fakePlaceSearchPolicyService{
			policy: locationsearch.SearchPolicy{
				AutocompleteRadiusMeters: 25000,
				UpdatedAt:                time.Now().UTC(),
				UpdatedBy:                "migration",
			},
		},
		adminReviewUsername: "reviewer",
		adminReviewPassword: "secret",
	}
	request := httptest.NewRequest(http.MethodGet, "/admin/operations", nil)
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "Marketplace timing") {
		t.Fatalf("marketplace controls missing: %s", body)
	}
	if !strings.Contains(body, "Rider place search") ||
		!strings.Contains(body, "Nearby search radius") ||
		!strings.Contains(body, "value=\"25\"") {
		t.Fatalf("place search controls missing: %s", body)
	}
	if strings.Contains(body, "Named-place snap radius") ||
		strings.Contains(body, "location-search-policy") {
		t.Fatalf("removed snap controls still rendered: %s", body)
	}
}

func TestAdminOperationsUpdatesPlaceSearchRadius(t *testing.T) {
	searchPolicy := &fakePlaceSearchPolicyService{
		policy: locationsearch.SearchPolicy{
			AutocompleteRadiusMeters: 50000,
			UpdatedAt:                time.Now().UTC(),
			UpdatedBy:                "migration",
		},
	}
	api := &API{
		locationSearchPolicy: searchPolicy,
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/admin/operations/place-search-policy",
		strings.NewReader("autocomplete_radius_kilometers=20"),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.adminUpdatePlaceSearchPolicy(response, request)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d: %s", response.Code, response.Body.String())
	}
	if searchPolicy.policy.AutocompleteRadiusMeters != 20000 {
		t.Fatalf("unexpected saved radius: %#v", searchPolicy.policy)
	}
	if searchPolicy.policy.UpdatedBy != "reviewer" {
		t.Fatalf("unexpected policy actor: %#v", searchPolicy.policy)
	}
}
