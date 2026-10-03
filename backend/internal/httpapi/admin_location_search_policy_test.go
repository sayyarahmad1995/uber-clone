package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

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
	if strings.Contains(body, "Named-place snap radius") ||
		strings.Contains(body, "location-search-policy") {
		t.Fatalf("removed snap controls still rendered: %s", body)
	}
}
