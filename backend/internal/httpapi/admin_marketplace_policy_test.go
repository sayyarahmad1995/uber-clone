package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

type fakeMarketplacePolicyService struct {
	loaded      marketplace.TimingPolicy
	loadErr     error
	updated     marketplace.TimingPolicy
	updateErr   error
	updateCalls int
	updateActor string
	updateInput marketplace.TimingPolicy
}

func (f *fakeMarketplacePolicyService) Load(context.Context) (marketplace.TimingPolicy, error) {
	return f.loaded, f.loadErr
}

func (f *fakeMarketplacePolicyService) Update(_ context.Context, policy marketplace.TimingPolicy, actor string) (marketplace.TimingPolicy, error) {
	f.updateCalls++
	f.updateInput = policy
	f.updateActor = actor
	return f.updated, f.updateErr
}

func TestGetAdminMarketplaceTimingPolicyUsesService(t *testing.T) {
	want := marketplace.TimingPolicy{
		RideRequestTTLSeconds:       180,
		DriverOpportunityTTLSeconds: 30,
		OfferDecisionTTLSeconds:     10,
		UpdatedAt:                   time.Date(2026, 9, 15, 9, 0, 0, 0, time.UTC),
		UpdatedBy:                   "reviewer",
	}
	service := &fakeMarketplacePolicyService{loaded: want}
	api := &API{marketplacePolicy: service}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/marketplace-timing-policy", nil)

	api.getAdminMarketplaceTimingPolicy(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var got marketplace.TimingPolicy
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got != want {
		t.Fatalf("response = %#v, want %#v", got, want)
	}
}

func TestUpdateAdminMarketplaceTimingPolicyUsesService(t *testing.T) {
	want := marketplace.TimingPolicy{
		RideRequestTTLSeconds:       240,
		DriverOpportunityTTLSeconds: 40,
		OfferDecisionTTLSeconds:     15,
		UpdatedAt:                   time.Date(2026, 9, 15, 9, 1, 0, 0, time.UTC),
		UpdatedBy:                   "reviewer",
	}
	service := &fakeMarketplacePolicyService{updated: want}
	api := &API{marketplacePolicy: service}
	body := []byte(`{"ride_request_ttl_seconds":240,"driver_opportunity_ttl_seconds":40,"offer_decision_ttl_seconds":15}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/v1/admin/marketplace-timing-policy", bytes.NewReader(body))
	request.SetBasicAuth("reviewer", "unused")

	api.updateAdminMarketplaceTimingPolicy(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.updateCalls != 1 {
		t.Fatalf("service Update() calls = %d, want 1", service.updateCalls)
	}
	if service.updateActor != "reviewer" {
		t.Fatalf("service actor = %q, want reviewer", service.updateActor)
	}
	wantInput := marketplace.TimingPolicy{
		RideRequestTTLSeconds:       240,
		DriverOpportunityTTLSeconds: 40,
		OfferDecisionTTLSeconds:     15,
	}
	if service.updateInput != wantInput {
		t.Fatalf("service policy = %#v, want %#v", service.updateInput, wantInput)
	}
}
