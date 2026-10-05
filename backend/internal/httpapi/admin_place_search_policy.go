package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

type adminPlaceSearchPolicyRequest struct {
	AutocompleteRadiusMeters int64 `json:"autocomplete_radius_meters"`
}

func (api *API) getAdminPlaceSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "place search policy is unavailable",
		})
		return
	}
	policy, err := api.locationSearchPolicy.Load(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "unable to load place search policy",
		})
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) updateAdminPlaceSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "place search policy is unavailable",
		})
		return
	}

	var body adminPlaceSearchPolicyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	policy, err := api.locationSearchPolicy.Update(
		r.Context(),
		locationsearch.SearchPolicy{
			AutocompleteRadiusMeters: body.AutocompleteRadiusMeters,
		},
		adminReviewer(r),
	)
	if err != nil {
		writePlaceSearchPolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) adminUpdatePlaceSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		http.Error(w, "Place search policy is unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	radiusKilometers, err := strconv.ParseInt(
		r.FormValue("autocomplete_radius_kilometers"),
		10,
		64,
	)
	if err != nil {
		http.Error(w, "Invalid place search radius", http.StatusBadRequest)
		return
	}

	_, err = api.locationSearchPolicy.Update(
		r.Context(),
		locationsearch.SearchPolicy{
			AutocompleteRadiusMeters: radiusKilometers * 1000,
		},
		adminReviewer(r),
	)
	if err != nil {
		if errors.Is(err, locationsearch.ErrInvalidSearchPolicy) {
			http.Error(w, "Place search radius must be between 5 and 50 km", http.StatusBadRequest)
			return
		}
		http.Error(w, "Unable to update place search policy", http.StatusInternalServerError)
		return
	}

	http.Redirect(
		w,
		r,
		"/admin/operations?message="+url.QueryEscape("Place search radius updated"),
		http.StatusSeeOther,
	)
}

func writePlaceSearchPolicyError(w http.ResponseWriter, err error) {
	if errors.Is(err, locationsearch.ErrInvalidSearchPolicy) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "place search radius is outside allowed limits",
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "unable to update place search policy",
	})
}
