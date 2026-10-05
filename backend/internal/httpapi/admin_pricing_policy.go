package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
)

type adminPricingPolicyRequest struct {
	ServiceCode            string `json:"service_code"`
	Currency               string `json:"currency"`
	BaseFareMinor          int64  `json:"base_fare_minor"`
	RateMinorPerKM         int64  `json:"rate_minor_per_km"`
	RateMinorPerMinute     int64  `json:"rate_minor_per_minute"`
	MinimumFareMinor       int64  `json:"minimum_fare_minor"`
	RoundingIncrementMinor int64  `json:"rounding_increment_minor"`
}

func (api *API) listAdminPricingPolicies(w http.ResponseWriter, r *http.Request) {
	if api.pricingPolicies == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "pricing policy management is unavailable",
		})
		return
	}
	policies, err := api.pricingPolicies.ListActive(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "unable to load pricing policies",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"pricing_policies": policies})
}

func (api *API) publishAdminPricingPolicy(w http.ResponseWriter, r *http.Request) {
	if api.pricingPolicies == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "pricing policy management is unavailable",
		})
		return
	}

	var body adminPricingPolicyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	policy, err := api.pricingPolicies.Publish(
		r.Context(),
		pricing.Draft{
			ServiceCode:        body.ServiceCode,
			Currency:           body.Currency,
			BaseFareMinor:      body.BaseFareMinor,
			RateMinorPerKM:     body.RateMinorPerKM,
			RateMinorPerMinute: body.RateMinorPerMinute,
			MinimumFareMinor:   body.MinimumFareMinor,
			RoundingIncrement:  body.RoundingIncrementMinor,
		},
		adminReviewer(r),
	)
	if err != nil {
		writePricingPolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, policy)
}

func (api *API) adminPublishPricingPolicy(w http.ResponseWriter, r *http.Request) {
	if api.pricingPolicies == nil {
		http.Error(w, "Pricing policy management is unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	parse := func(name string) (int64, error) {
		return pricing.ParseMajorAmount(r.FormValue(name))
	}
	baseFare, err := parse("base_fare")
	if err != nil {
		http.Error(w, "Invalid base fare", http.StatusBadRequest)
		return
	}
	ratePerKM, err := parse("rate_per_km")
	if err != nil {
		http.Error(w, "Invalid per-kilometer rate", http.StatusBadRequest)
		return
	}
	ratePerMinute, err := parse("rate_per_minute")
	if err != nil {
		http.Error(w, "Invalid per-minute rate", http.StatusBadRequest)
		return
	}
	minimumFare, err := parse("minimum_fare")
	if err != nil {
		http.Error(w, "Invalid minimum fare", http.StatusBadRequest)
		return
	}
	roundingIncrement, err := parse("rounding_increment")
	if err != nil || roundingIncrement <= 0 {
		http.Error(w, "Invalid rounding increment", http.StatusBadRequest)
		return
	}

	policy, err := api.pricingPolicies.Publish(
		r.Context(),
		pricing.Draft{
			ServiceCode:        strings.TrimSpace(r.FormValue("service_code")),
			Currency:           "PKR",
			BaseFareMinor:      baseFare,
			RateMinorPerKM:     ratePerKM,
			RateMinorPerMinute: ratePerMinute,
			MinimumFareMinor:   minimumFare,
			RoundingIncrement:  roundingIncrement,
		},
		adminReviewer(r),
	)
	if err != nil {
		if errors.Is(err, pricing.ErrPolicyInvalid) {
			http.Error(w, "Invalid pricing policy or unavailable Rider service", http.StatusBadRequest)
			return
		}
		http.Error(w, "Unable to publish pricing policy", http.StatusInternalServerError)
		return
	}

	message := "Published " + policy.ServiceCode + " pricing policy v" +
		strconv.FormatInt(policy.Version, 10)
	http.Redirect(
		w,
		r,
		"/admin/operations?message="+url.QueryEscape(message),
		http.StatusSeeOther,
	)
}

func writePricingPolicyError(w http.ResponseWriter, err error) {
	if errors.Is(err, pricing.ErrPolicyInvalid) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid pricing policy or unavailable Rider service",
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "unable to publish pricing policy",
	})
}
