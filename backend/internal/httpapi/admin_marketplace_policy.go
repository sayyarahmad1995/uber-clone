package httpapi

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
)

type adminMarketplacePolicyRequest struct {
	RideRequestTTLSeconds       int64 `json:"ride_request_ttl_seconds"`
	DriverOpportunityTTLSeconds int64 `json:"driver_opportunity_ttl_seconds"`
	OfferDecisionTTLSeconds     int64 `json:"offer_decision_ttl_seconds"`
}

type adminOperationsView struct {
	Policy              marketplace.TimingPolicy
	PlaceSearchPolicy   locationsearch.SearchPolicy
	PlaceSearchRadiusKm int64
	PricingPolicies     []adminPricingPolicyView
	Message             string
}

type adminPricingPolicyView struct {
	ServiceCode       string
	Version           int64
	BaseFare          string
	RatePerKM         string
	RatePerMinute     string
	MinimumFare       string
	RoundingIncrement string
}

var adminOperationsTemplate = template.Must(template.New("admin-operations").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>HiGO operations</title>
<style>
body{font-family:system-ui,sans-serif;max-width:760px;margin:40px auto;padding:0 20px;color:#1f2937}.card{border:1px solid #e5e7eb;border-radius:10px;padding:18px;margin:18px 0}.grid{display:grid;grid-template-columns:1fr 150px;gap:14px 18px;align-items:center}label{font-weight:600}input{width:100%;box-sizing:border-box;padding:9px}button{padding:10px 16px;border:0;border-radius:7px;background:#1d4ed8;color:white;cursor:pointer}a{color:#1d4ed8}.muted{color:#6b7280}.message{padding:10px;background:#eff6ff;border-radius:7px}.hint{font-size:.9em;color:#6b7280;margin-top:3px}
</style>
</head>
<body>
<h1>HiGO operations</h1>
<p><a href="/admin/driver-onboarding">Driver onboarding review →</a></p>
{{if .Message}}<p class="message">{{.Message}}</p>{{end}}
<div class="card">
<h2>Marketplace timing</h2>
<p class="muted">Changes apply only to newly created ride requests, Driver opportunities, and offers. Existing deadlines are not changed.</p>
<form method="post" action="/admin/operations/marketplace-timing-policy">
<div class="grid">
<div><label for="ride_request_ttl_seconds">Ride request lifetime</label><div class="hint">60–600 seconds</div></div>
<input id="ride_request_ttl_seconds" name="ride_request_ttl_seconds" type="number" min="60" max="600" required value="{{.Policy.RideRequestTTLSeconds}}">
<div><label for="driver_opportunity_ttl_seconds">Driver response window</label><div class="hint">10–60 seconds</div></div>
<input id="driver_opportunity_ttl_seconds" name="driver_opportunity_ttl_seconds" type="number" min="10" max="60" required value="{{.Policy.DriverOpportunityTTLSeconds}}">
<div><label for="offer_decision_ttl_seconds">Rider offer decision window</label><div class="hint">5–60 seconds</div></div>
<input id="offer_decision_ttl_seconds" name="offer_decision_ttl_seconds" type="number" min="5" max="60" required value="{{.Policy.OfferDecisionTTLSeconds}}">
</div>
<p><button type="submit">Save marketplace timing</button></p>
</form>
<p class="muted">Last updated {{.Policy.UpdatedAt}} by {{.Policy.UpdatedBy}}</p>
</div>
<div class="card">
<h2>Rider place search</h2>
<p class="muted">Controls how far Google autocomplete searches around the Rider's current device location. This does not snap or move Rider-selected pins.</p>
<form method="post" action="/admin/operations/place-search-policy">
<div class="grid">
<div><label for="autocomplete_radius_kilometers">Nearby search radius</label><div class="hint">5–50 km; default 50 km</div></div>
<input id="autocomplete_radius_kilometers" name="autocomplete_radius_kilometers" type="number" min="5" max="50" step="1" required value="{{.PlaceSearchRadiusKm}}">
</div>
<p><button type="submit">Save place search radius</button></p>
</form>
<p class="muted">Last updated {{.PlaceSearchPolicy.UpdatedAt}} by {{.PlaceSearchPolicy.UpdatedBy}}</p>
</div>
<div class="card">
<h2>Suggested fare pricing</h2>
<p class="muted">Publishing creates a new approved version for one Rider-visible service. Existing Ride Requests and Trips are never repriced.</p>
{{if .PricingPolicies}}
<ul>
{{range .PricingPolicies}}
<li><strong>{{.ServiceCode}}</strong> · v{{.Version}} · base PKR {{.BaseFare}} · PKR {{.RatePerKM}}/km · PKR {{.RatePerMinute}}/min · minimum PKR {{.MinimumFare}} · round PKR {{.RoundingIncrement}}</li>
{{end}}
</ul>
{{else}}
<p class="muted">No approved active pricing policies. Ride previews cannot suggest fares until a policy is published.</p>
{{end}}
<form method="post" action="/admin/operations/pricing-policy">
<div class="grid">
<div><label for="pricing_service_code">Service code</label><div class="hint">Must be active and Rider-visible</div></div>
<input id="pricing_service_code" name="service_code" type="text" required>
<div><label for="base_fare">Base fare</label><div class="hint">PKR</div></div>
<input id="base_fare" name="base_fare" type="number" min="0" step="0.01" required>
<div><label for="rate_per_km">Distance rate</label><div class="hint">PKR per km</div></div>
<input id="rate_per_km" name="rate_per_km" type="number" min="0" step="0.01" required>
<div><label for="rate_per_minute">Duration rate</label><div class="hint">PKR per route minute</div></div>
<input id="rate_per_minute" name="rate_per_minute" type="number" min="0" step="0.01" required>
<div><label for="minimum_fare">Minimum fare</label><div class="hint">PKR</div></div>
<input id="minimum_fare" name="minimum_fare" type="number" min="0.01" step="0.01" required>
<div><label for="rounding_increment">Final rounding increment</label><div class="hint">PKR; one final half-up rounding step</div></div>
<input id="rounding_increment" name="rounding_increment" type="number" min="0.01" step="0.01" required>
</div>
<p><button type="submit">Publish pricing version</button></p>
</form>
</div>
</body>
</html>`))

func (api *API) getAdminMarketplaceTimingPolicy(w http.ResponseWriter, r *http.Request) {
	policy, err := api.marketplacePolicy.Load(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load marketplace timing policy"})
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) updateAdminMarketplaceTimingPolicy(w http.ResponseWriter, r *http.Request) {
	var body adminMarketplacePolicyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	policy, err := api.marketplacePolicy.Update(r.Context(), timingPolicyFromAdminRequest(body), adminReviewer(r))
	if err != nil {
		writeMarketplacePolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) adminOperationsIndex(w http.ResponseWriter, r *http.Request) {
	policy, err := api.marketplacePolicy.Load(r.Context())
	if err != nil {
		http.Error(w, "Unable to load marketplace timing policy", http.StatusInternalServerError)
		return
	}

	searchPolicy := locationsearch.SearchPolicy{
		AutocompleteRadiusMeters: locationsearch.DefaultAutocompleteRadiusMeters,
		UpdatedBy:                "default",
	}
	if api.locationSearchPolicy != nil {
		searchPolicy, err = api.locationSearchPolicy.Load(r.Context())
		if err != nil {
			http.Error(w, "Unable to load place search policy", http.StatusInternalServerError)
			return
		}
	}

	pricingPolicies := make([]adminPricingPolicyView, 0)
	if api.pricingPolicies != nil {
		active, err := api.pricingPolicies.ListActive(r.Context())
		if err != nil {
			http.Error(w, "Unable to load pricing policies", http.StatusInternalServerError)
			return
		}
		for _, policy := range active {
			pricingPolicies = append(pricingPolicies, adminPricingPolicyView{
				ServiceCode:       policy.ServiceCode,
				Version:           policy.Version,
				BaseFare:          pricing.FormatMajorAmount(policy.BaseFareMinor),
				RatePerKM:         pricing.FormatMajorAmount(policy.RateMinorPerKM),
				RatePerMinute:     pricing.FormatMajorAmount(policy.RateMinorPerMinute),
				MinimumFare:       pricing.FormatMajorAmount(policy.MinimumFareMinor),
				RoundingIncrement: pricing.FormatMajorAmount(policy.RoundingIncrement),
			})
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminOperationsTemplate.Execute(w, adminOperationsView{
		Policy:              policy,
		PlaceSearchPolicy:   searchPolicy,
		PlaceSearchRadiusKm: searchPolicy.AutocompleteRadiusMeters / 1000,
		PricingPolicies:     pricingPolicies,
		Message:             r.URL.Query().Get("message"),
	})
}

func (api *API) adminUpdateMarketplaceTimingPolicy(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	rideTTL, err := strconv.ParseInt(r.FormValue("ride_request_ttl_seconds"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid marketplace timing policy", http.StatusBadRequest)
		return
	}
	opportunityTTL, err := strconv.ParseInt(r.FormValue("driver_opportunity_ttl_seconds"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid marketplace timing policy", http.StatusBadRequest)
		return
	}
	offerTTL, err := strconv.ParseInt(r.FormValue("offer_decision_ttl_seconds"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid marketplace timing policy", http.StatusBadRequest)
		return
	}
	_, err = api.marketplacePolicy.Update(r.Context(), marketplace.TimingPolicy{
		RideRequestTTLSeconds:       rideTTL,
		DriverOpportunityTTLSeconds: opportunityTTL,
		OfferDecisionTTLSeconds:     offerTTL,
	}, adminReviewer(r))
	if err != nil {
		if errors.Is(err, marketplace.ErrInvalidTimingPolicy) {
			http.Error(w, "Invalid marketplace timing policy", http.StatusBadRequest)
			return
		}
		http.Error(w, "Unable to update marketplace timing policy", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/operations?message="+url.QueryEscape("Marketplace timing policy updated"), http.StatusSeeOther)
}

func timingPolicyFromAdminRequest(body adminMarketplacePolicyRequest) marketplace.TimingPolicy {
	return marketplace.TimingPolicy{
		RideRequestTTLSeconds:       body.RideRequestTTLSeconds,
		DriverOpportunityTTLSeconds: body.DriverOpportunityTTLSeconds,
		OfferDecisionTTLSeconds:     body.OfferDecisionTTLSeconds,
	}
}

func writeMarketplacePolicyError(w http.ResponseWriter, err error) {
	if errors.Is(err, marketplace.ErrInvalidTimingPolicy) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "marketplace timing policy is outside allowed limits"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update marketplace timing policy"})
}
