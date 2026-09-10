package httpapi

import "net/http"

func (api *API) listDriverVehicles(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	vehicles, err := api.drivers.ListVehicles(r.Context(), u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load vehicles"})
		return
	}
	items := make([]map[string]any, 0, len(vehicles))
	for _, vehicle := range vehicles {
		enrollments := make([]map[string]any, 0, len(vehicle.Enrollments))
		for _, enrollment := range vehicle.Enrollments {
			enrollments = append(enrollments, map[string]any{
				"service_code": enrollment.ServiceCode, "display_name": enrollment.DisplayName,
				"approved_at": enrollment.ApprovedAt, "service_active": enrollment.ServiceActive,
			})
		}
		items = append(items, map[string]any{
			"id": vehicle.ID, "make": vehicle.Make, "model": vehicle.Model,
			"model_year": modelYearResponse(vehicle.ModelYear), "color": vehicle.Color,
			"license_plate": vehicle.LicensePlate, "approved_service_enrollments": enrollments,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"vehicles": items})
}
