package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/driveronboarding"
)

type driverOnboardingRequest struct {
	DisplayName string `json:"display_name"`
	ServiceCode string `json:"service_code"`
	Vehicle     struct {
		Make         string `json:"make"`
		Model        string `json:"model"`
		ModelYear    int    `json:"model_year"`
		Color        string `json:"color"`
		LicensePlate string `json:"license_plate"`
	} `json:"vehicle"`
}

func (r driverOnboardingRequest) input() driveronboarding.ApplicationInput {
	return driveronboarding.ApplicationInput{
		DisplayName: r.DisplayName,
		ServiceCode: r.ServiceCode,
		Vehicle: driveronboarding.VehicleInput{
			Make:         r.Vehicle.Make,
			Model:        r.Vehicle.Model,
			ModelYear:    r.Vehicle.ModelYear,
			Color:        r.Vehicle.Color,
			LicensePlate: r.Vehicle.LicensePlate,
		},
	}
}

func (api *API) listDriverServices(w http.ResponseWriter, r *http.Request) {
	if _, ok := api.requireDriverCapability(w, r); !ok {
		return
	}
	services, err := api.driverOnboarding.ListServices(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load driver services"})
		return
	}
	items := make([]map[string]any, 0, len(services))
	for _, service := range services {
		items = append(items, driverServiceResponse(service))
	}
	writeJSON(w, http.StatusOK, map[string]any{"services": items})
}

func (api *API) getDriverOnboarding(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	application, err := api.driverOnboarding.Latest(r.Context(), u.ID)
	if errors.Is(err, driveronboarding.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "driver onboarding application not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load driver onboarding application"})
		return
	}
	writeJSON(w, http.StatusOK, driverOnboardingResponse(application))
}

func (api *API) precheckDriverOnboarding(w http.ResponseWriter, r *http.Request) {
	if _, ok := api.requireDriverCapability(w, r); !ok {
		return
	}
	body, ok := decodeDriverOnboardingRequest(w, r)
	if !ok {
		return
	}
	result, err := api.driverOnboarding.Precheck(r.Context(), body.input())
	switch {
	case errors.Is(err, driveronboarding.ErrInvalidApplication):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "complete Driver, service, and vehicle details are required"})
		return
	case errors.Is(err, driveronboarding.ErrServiceNotFound):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "selected service is unavailable"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to check service eligibility"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"eligible": result.Eligible,
		"reasons":  result.Reasons,
		"service":  driverServiceResponse(result.Service),
	})
}

func (api *API) submitDriverOnboarding(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireDriverCapability(w, r)
	if !ok {
		return
	}
	body, ok := decodeDriverOnboardingRequest(w, r)
	if !ok {
		return
	}
	application, err := api.driverOnboarding.Submit(r.Context(), u.ID, body.input())
	switch {
	case errors.Is(err, driveronboarding.ErrInvalidApplication):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "complete Driver, service, and vehicle details are required"})
		return
	case errors.Is(err, driveronboarding.ErrServiceNotFound):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "selected service is unavailable"})
		return
	case errors.Is(err, driveronboarding.ErrIneligible):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "vehicle is not eligible for selected service"})
		return
	case errors.Is(err, driveronboarding.ErrPendingApplication):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a Driver onboarding application is already under review"})
		return
	case errors.Is(err, driveronboarding.ErrAlreadyOnboarded):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Driver onboarding is already complete; use the vehicle/service application flow"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to submit Driver onboarding application"})
		return
	}
	writeJSON(w, http.StatusCreated, driverOnboardingResponse(application))
}

func decodeDriverOnboardingRequest(w http.ResponseWriter, r *http.Request) (driverOnboardingRequest, bool) {
	var body driverOnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return driverOnboardingRequest{}, false
	}
	return body, true
}

func driverServiceResponse(service driveronboarding.ServiceOption) map[string]any {
	return map[string]any{
		"code":                 service.Code,
		"display_name":         service.DisplayName,
		"description":          service.Description,
		"minimum_model_year":   service.MinimumModelYear,
		"implied_service_code": service.ImpliedServiceCode,
	}
}

func driverOnboardingResponse(application driveronboarding.Application) map[string]any {
	return map[string]any{
		"id":           application.ID,
		"display_name": application.DisplayName,
		"status":       application.Status,
		"service":      driverServiceResponse(application.Service),
		"vehicle": map[string]any{
			"make":          application.Vehicle.Make,
			"model":         application.Vehicle.Model,
			"model_year":    application.Vehicle.ModelYear,
			"color":         application.Vehicle.Color,
			"license_plate": application.Vehicle.LicensePlate,
		},
		"rejection_reason": application.RejectionReason,
		"submitted_at":     application.SubmittedAt,
		"decided_at":       application.DecidedAt,
	}
}
