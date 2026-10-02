package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

type adminLocationSearchPolicyRequest struct {
	NamedPlaceSnapRadiusMeters int64 `json:"named_place_snap_radius_meters"`
}

func (api *API) getAdminLocationSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load location search policy"})
		return
	}
	policy, err := api.locationSearchPolicy.Load(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load location search policy"})
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) updateAdminLocationSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update location search policy"})
		return
	}
	var body adminLocationSearchPolicyRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	policy, err := api.locationSearchPolicy.Update(
		r.Context(),
		locationsearch.Policy{NamedPlaceSnapRadiusMeters: body.NamedPlaceSnapRadiusMeters},
		adminReviewer(r),
	)
	if err != nil {
		writeLocationSearchPolicyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (api *API) adminUpdateLocationSearchPolicy(w http.ResponseWriter, r *http.Request) {
	if api.locationSearchPolicy == nil {
		http.Error(w, "Unable to update location search policy", http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	radius, err := strconv.ParseInt(r.FormValue("named_place_snap_radius_meters"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid location search policy", http.StatusBadRequest)
		return
	}
	_, err = api.locationSearchPolicy.Update(
		r.Context(),
		locationsearch.Policy{NamedPlaceSnapRadiusMeters: radius},
		adminReviewer(r),
	)
	if err != nil {
		if errors.Is(err, locationsearch.ErrInvalidPolicy) {
			http.Error(w, "Invalid location search policy", http.StatusBadRequest)
			return
		}
		http.Error(w, "Unable to update location search policy", http.StatusInternalServerError)
		return
	}
	http.Redirect(
		w,
		r,
		"/admin/operations?message="+url.QueryEscape("Location search policy updated"),
		http.StatusSeeOther,
	)
}

func writeLocationSearchPolicyError(w http.ResponseWriter, err error) {
	if errors.Is(err, locationsearch.ErrInvalidPolicy) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "location search policy is outside allowed limits"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to update location search policy"})
}
