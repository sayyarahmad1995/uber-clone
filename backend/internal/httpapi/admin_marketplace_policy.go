package httpapi

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/marketplace"
)

type adminMarketplacePolicyRequest struct {
	RideRequestTTLSeconds       int64 `json:"ride_request_ttl_seconds"`
	DriverOpportunityTTLSeconds int64 `json:"driver_opportunity_ttl_seconds"`
	OfferDecisionTTLSeconds     int64 `json:"offer_decision_ttl_seconds"`
}

type adminOperationsView struct {
	Policy  marketplace.TimingPolicy
	Message string
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
</body>
</html>`))

func (api *API) getAdminMarketplaceTimingPolicy(w http.ResponseWriter, r *http.Request) {
	policy, err := marketplace.LoadTimingPolicy(r.Context(), api.db)
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
	policy, err := marketplace.UpdateTimingPolicy(r.Context(), api.db, timingPolicyFromAdminRequest(body), adminReviewer(r))
	if err != nil {
		writeMarketplacePolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) adminOperationsIndex(w http.ResponseWriter, r *http.Request) {
	policy, err := marketplace.LoadTimingPolicy(r.Context(), api.db)
	if err != nil {
		http.Error(w, "Unable to load marketplace timing policy", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminOperationsTemplate.Execute(w, adminOperationsView{Policy: policy, Message: r.URL.Query().Get("message")})
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
	_, err = marketplace.UpdateTimingPolicy(r.Context(), api.db, marketplace.TimingPolicy{
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
