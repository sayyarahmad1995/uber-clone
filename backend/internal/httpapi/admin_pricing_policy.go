package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"html/template"
	"net/http"
	"net/url"
)

type adminPricingBody struct {
	ServiceCode            string `json:"service_code"`
	Currency               string `json:"currency"`
	BaseFareMinor          int64  `json:"base_fare_minor"`
	RateMinorPerKm         int64  `json:"rate_minor_per_km"`
	RateMinorPerMinute     int64  `json:"rate_minor_per_minute"`
	MinimumFareMinor       int64  `json:"minimum_fare_minor"`
	RoundingIncrementMinor int64  `json:"rounding_increment_minor"`
}

func (b adminPricingBody) draft() pricing.Draft {
	return pricing.Draft{ServiceCode: b.ServiceCode, Currency: b.Currency, BaseFareMinor: b.BaseFareMinor, RateMinorPerKm: b.RateMinorPerKm, RateMinorPerMinute: b.RateMinorPerMinute, MinimumFareMinor: b.MinimumFareMinor, RoundingIncrementMinor: b.RoundingIncrementMinor}
}
func policyJSON(p pricing.Policy) map[string]any {
	return map[string]any{"id": p.ID, "service_code": p.ServiceCode, "currency": p.Currency, "version": p.Version, "base_fare_minor": p.BaseFareMinor, "rate_minor_per_km": p.RateMinorPerKm, "rate_minor_per_minute": p.RateMinorPerMinute, "minimum_fare_minor": p.MinimumFareMinor, "rounding_increment_minor": p.RoundingIncrementMinor, "calculation_rule": p.CalculationRule, "effective_from": p.EffectiveFrom, "published_at": p.PublishedAt, "published_by": p.PublishedBy}
}
func (api *API) pricingReady(w http.ResponseWriter) bool {
	if api.pricing == nil {
		writePricingError(w, pricing.ErrUnavailable)
		return false
	}
	return true
}
func (api *API) getAdminPricingPolicies(w http.ResponseWriter, r *http.Request) {
	if !api.pricingReady(w) {
		return
	}
	policies, err := api.pricing.List(r.Context())
	if err != nil {
		writePricingError(w, err)
		return
	}
	result := make([]map[string]any, 0, len(policies))
	for _, p := range policies {
		result = append(result, policyJSON(p))
	}
	writeJSON(w, 200, map[string]any{"policies": result})
}
func (api *API) publishAdminPricingPolicy(w http.ResponseWriter, r *http.Request) {
	if !api.pricingReady(w) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	var body adminPricingBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writePricingError(w, pricing.ErrInvalidPolicy)
		return
	}
	p, err := api.pricing.Publish(r.Context(), body.draft(), adminReviewer(r))
	if err != nil {
		writePricingError(w, err)
		return
	}
	writeJSON(w, 201, policyJSON(p))
}
func (api *API) disablePricingPolicy(w http.ResponseWriter, r *http.Request, idText string) bool {
	if !api.pricingReady(w) {
		return false
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		writePricingError(w, pricing.ErrInvalidPolicy)
		return false
	}
	policies, err := api.pricing.List(r.Context())
	if err != nil {
		writePricingError(w, err)
		return false
	}
	for _, p := range policies {
		if p.ID == id {
			if err := api.pricing.Disable(r.Context(), p.ServiceCode, p.Currency, id, adminReviewer(r)); err != nil {
				writePricingError(w, err)
				return false
			}
			return true
		}
	}
	writePricingError(w, pricing.ErrInvalidPolicy)
	return false
}
func (api *API) disableAdminPricingPolicy(w http.ResponseWriter, r *http.Request) {
	if api.disablePricingPolicy(w, r, r.PathValue("policy_id")) {
		writeJSON(w, 200, map[string]bool{"disabled": true})
	}
}

type adminPricingItem struct {
	Policy  pricing.Policy
	Current bool
}

var adminPricingTemplate = template.Must(template.New("pricing").Funcs(template.FuncMap{"rupees": func(v int64) string { return formatPKR(v) }}).Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>HiGO pricing</title>
<style>body{font-family:system-ui;max-width:1000px;margin:30px auto;padding:0 20px}label{display:block;margin-top:12px}input{padding:8px}table{border-collapse:collapse;width:100%}th,td{text-align:left;padding:8px;border-bottom:1px solid #ddd}button{padding:10px;margin-top:12px}.history{overflow:auto}</style></head>
<body><p><a href="/admin/operations">Operations</a></p><h1>Pricing policies</h1><p>All amounts are PKR rupees. Publishing approves these rates for new previews only. Open ride requests and agreed Trip fares remain unchanged.</p>
<form method="post" action="/admin/operations/pricing/publish">
<label>Service code <input name="service_code" required maxlength="100"></label><p>Use an active Rider-visible catalog code, including a new compatible service that has no pricing yet.</p>
<label>Base fare <input type="number" name="base_fare" min="0" max="10000000000" step="0.01" required></label>
<label>Rate per kilometer <input type="number" name="rate_per_km" min="0" max="10000000000" step="0.01" required></label>
<label>Rate per minute <input type="number" name="rate_per_minute" min="0" max="10000000000" step="0.01" required></label>
<label>Minimum fare <input type="number" name="minimum_fare" min="0.01" max="10000000000" step="0.01" required></label>
<label>Rounding increment <input type="number" name="rounding_increment" min="0.01" max="10000000000" step="0.01" required></label>
<button type="submit">Publish approved tariff</button></form>
<h2>Published history</h2><div class="history"><table><thead><tr><th>Service / version</th><th>Base</th><th>Per km</th><th>Per minute</th><th>Minimum</th><th>Rounding</th><th>Approved by / effective</th><th>State</th></tr></thead><tbody>
{{range .}}<tr><td>{{.Policy.ServiceCode}} / {{.Policy.Version}}</td><td>{{rupees .Policy.BaseFareMinor}}</td><td>{{rupees .Policy.RateMinorPerKm}}</td><td>{{rupees .Policy.RateMinorPerMinute}}</td><td>{{rupees .Policy.MinimumFareMinor}}</td><td>{{rupees .Policy.RoundingIncrementMinor}}</td><td>{{.Policy.PublishedBy}} / {{.Policy.EffectiveFrom}}</td><td>{{if .Current}}Current<form method="post" action="/admin/operations/pricing/disable"><input type="hidden" name="policy_id" value="{{.Policy.ID}}"><button type="submit">Disable new priced requests</button></form>{{else}}Historical{{end}}</td></tr>{{else}}<tr><td colspan="8">No approved tariffs have been published.</td></tr>{{end}}
</tbody></table></div></body></html>`))

func formatPKR(v int64) string { return fmt.Sprintf("%d.%02d", v/100, v%100) }
func (api *API) adminPricingIndex(w http.ResponseWriter, r *http.Request) {
	if !api.pricingReady(w) {
		return
	}
	policies, err := api.pricing.List(r.Context())
	if err != nil {
		writePricingError(w, err)
		return
	}
	items := make([]adminPricingItem, 0, len(policies))
	for _, p := range policies {
		current, err := api.pricing.Current(r.Context(), p.ServiceCode, p.Currency)
		if err != nil && !errors.Is(err, pricing.ErrUnavailable) && !errors.Is(err, pricing.ErrInvalidService) {
			writePricingError(w, err)
			return
		}
		items = append(items, adminPricingItem{p, err == nil && current.ID == p.ID})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminPricingTemplate.Execute(w, items)
}
func (api *API) adminPublishPricingPolicy(w http.ResponseWriter, r *http.Request) {
	if !api.pricingReady(w) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := r.ParseForm(); err != nil {
		writePricingError(w, pricing.ErrInvalidPolicy)
		return
	}
	fields := []string{"base_fare", "rate_per_km", "rate_per_minute", "minimum_fare", "rounding_increment"}
	amounts := make([]int64, len(fields))
	for i, key := range fields {
		n, err := pricing.ParsePKR(r.FormValue(key))
		if err != nil {
			writePricingError(w, err)
			return
		}
		amounts[i] = n
	}
	d := pricing.Draft{ServiceCode: r.FormValue("service_code"), Currency: "PKR", BaseFareMinor: amounts[0], RateMinorPerKm: amounts[1], RateMinorPerMinute: amounts[2], MinimumFareMinor: amounts[3], RoundingIncrementMinor: amounts[4]}
	if _, err := api.pricing.Publish(r.Context(), d, adminReviewer(r)); err != nil {
		writePricingError(w, err)
		return
	}
	http.Redirect(w, r, "/admin/operations/pricing?message="+url.QueryEscape("Approved tariff published"), 303)
}
func (api *API) adminDisablePricingPolicy(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := r.ParseForm(); err != nil {
		writePricingError(w, pricing.ErrInvalidPolicy)
		return
	}
	if api.disablePricingPolicy(w, r, r.FormValue("policy_id")) {
		http.Redirect(w, r, "/admin/operations/pricing", 303)
	}
}
