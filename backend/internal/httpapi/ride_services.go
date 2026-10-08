package httpapi

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/rideservice"
)

func (api *API) listRideServices(w http.ResponseWriter, r *http.Request) {
	if _, ok := api.requireRiderCapability(w, r); !ok {
		return
	}
	services, err := api.rideServices.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "unable to load ride services",
		})
		return
	}
	writeJSON(w, http.StatusOK, rideServicesResponse(services))
}

func rideServicesResponse(services []rideservice.Option) map[string]any {
	items := make([]map[string]any, 0, len(services))
	for _, service := range services {
		items = append(items, map[string]any{
			"pricing_required":   service.PricingRequired,
			"code":               service.Code,
			"display_name":       service.DisplayName,
			"description":        service.Description,
			"display_order":      service.DisplayOrder,
			"presentation_token": service.PresentationToken,
		})
	}
	return map[string]any{"services": items}
}
