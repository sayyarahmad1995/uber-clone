package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type rideLocationRequest struct {
	PlaceID   string   `json:"place_id,omitempty"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type rideFareRequest struct {
	AmountMinor *int64  `json:"amount_minor"`
	Currency    *string `json:"currency"`
}

type createRideRequestBody struct {
	PricingPolicyVersion string               `json:"pricing_policy_version"`
	ServiceCode          string               `json:"service_code"`
	Pickup               *rideLocationRequest `json:"pickup"`
	Destination          *rideLocationRequest `json:"destination"`
	ProposedFare         *rideFareRequest     `json:"proposed_fare"`
}

func (body createRideRequestBody) input() (ride.CreateInput, bool) {
	if body.Pickup == nil || body.Destination == nil || body.Pickup.Latitude == nil || body.Pickup.Longitude == nil || body.Destination.Latitude == nil || body.Destination.Longitude == nil || body.ProposedFare == nil || body.ProposedFare.AmountMinor == nil || body.ProposedFare.Currency == nil {
		return ride.CreateInput{}, false
	}
	return ride.CreateInput{
		PricingPolicyVersion: body.PricingPolicyVersion,
		ServiceCode:          body.ServiceCode,
		Pickup:               ride.Location{Latitude: *body.Pickup.Latitude, Longitude: *body.Pickup.Longitude},
		Destination:          ride.Location{Latitude: *body.Destination.Latitude, Longitude: *body.Destination.Longitude},
		ProposedFare: &ride.Money{
			AmountMinor: *body.ProposedFare.AmountMinor,
			Currency:    *body.ProposedFare.Currency,
		},
	}, true
}

func (api *API) createRideRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	var body createRideRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	input, ok := body.input()
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup, destination, and proposed_fare are required"})
		return
	}
	if api.suggestedFaresEnabled && !api.validateSuggestedFare(w, r, body, input, u.ID.String()) {
		return
	}
	request, err := api.rides.Create(r.Context(), u.ID, input)
	switch {
	case errors.Is(err, pricing.ErrInvalidPolicy), errors.Is(err, pricing.ErrInvalidService), errors.Is(err, pricing.ErrPolicyChanged), errors.Is(err, pricing.ErrUnavailable):
		writePricingError(w, err)
		return
	case errors.Is(err, ride.ErrInvalidService):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "selected service is unavailable"})
		return
	case errors.Is(err, ride.ErrInvalidLocation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup and destination coordinates are invalid"})
		return
	case errors.Is(err, ride.ErrInvalidFare):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "proposed_fare must be positive and use a three-letter currency"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to create ride request"})
		return
	}
	writeRideRequest(w, http.StatusCreated, request)
}

func (api *API) listRideRequests(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	views, err := api.rideStatuses.ListOwned(r.Context(), u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to list ride requests"})
		return
	}
	writeJSON(w, http.StatusOK, rideRequestListResponse(views))
}

func (api *API) getRideRequestStatus(w http.ResponseWriter, r *http.Request) {
	u, ok := api.requireRiderCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}
	view, err := api.rideStatuses.GetOwned(r.Context(), rideRequestID, u.ID)
	switch {
	case errors.Is(err, ridestatus.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request not found"})
		return
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to get ride request status"})
		return
	}
	writeJSON(w, http.StatusOK, rideRequestStatusResponse(view.RideRequest, view.Trip))
}

func writeRideRequest(w http.ResponseWriter, status int, request ride.Request) {
	response := map[string]any{
		"service_code":  request.ServiceCode,
		"id":            request.ID,
		"rider_user_id": request.RiderUserID,
		"pickup":        map[string]any{"latitude": request.Pickup.Latitude, "longitude": request.Pickup.Longitude},
		"destination":   map[string]any{"latitude": request.Destination.Latitude, "longitude": request.Destination.Longitude},
		"status":        request.Status,
		"created_at":    request.CreatedAt,
		"expires_at":    request.ExpiresAt,
	}
	if request.ProposedFare != nil {
		response["proposed_fare"] = map[string]any{"amount_minor": request.ProposedFare.AmountMinor, "currency": request.ProposedFare.Currency}
	}
	writeJSON(w, status, response)
}

func rideRequestListResponse(views []ridestatus.View) map[string]any {
	requests := make([]map[string]any, 0, len(views))
	for _, view := range views {
		requests = append(requests, rideRequestStatusResponse(view.RideRequest, view.Trip))
	}
	return map[string]any{"ride_requests": requests}
}

func rideRequestStatusResponse(request ride.Request, assignedTrip *trip.Trip) map[string]any {
	response := map[string]any{
		"service_code": request.ServiceCode,
		"id":           request.ID,
		"pickup":       map[string]any{"latitude": request.Pickup.Latitude, "longitude": request.Pickup.Longitude},
		"destination":  map[string]any{"latitude": request.Destination.Latitude, "longitude": request.Destination.Longitude},
		"status":       request.Status,
		"created_at":   request.CreatedAt,
		"expires_at":   request.ExpiresAt,
		"trip":         nil,
	}
	if request.ProposedFare != nil {
		response["proposed_fare"] = map[string]any{"amount_minor": request.ProposedFare.AmountMinor, "currency": request.ProposedFare.Currency}
	}
	if request.CancelledAt != nil {
		response["cancelled_at"] = request.CancelledAt
		response["cancelled_by"] = request.CancelledBy
	}
	if assignedTrip != nil {
		response["trip"] = map[string]any{
			"operation_context": assignedTrip.OperationContext,
			"driver_user_id":    assignedTrip.DriverUserID,
			"status":            assignedTrip.Status,
			"assigned_at":       assignedTrip.AssignedAt,
			"started_at":        assignedTrip.StartedAt,
			"completed_at":      assignedTrip.CompletedAt,
			"cancelled_at":      assignedTrip.CancelledAt,
			"settlement":        tripSettlementResponse(assignedTrip.Settlement),
		}
	}
	return response
}

// Revalidate the displayed amount without trusting client route metrics or rates.
// The repository still atomically checks the policy version after routing returns.
func (api *API) validateSuggestedFare(w http.ResponseWriter, r *http.Request, body createRideRequestBody, input ride.CreateInput, userID string) bool {
	expected, err := uuid.Parse(input.PricingPolicyVersion)
	if err != nil || input.ProposedFare.Currency != "PKR" || input.ProposedFare.AmountMinor <= 0 || input.ProposedFare.AmountMinor > ride.MaxFareMinor {
		writePricingError(w, pricing.ErrInvalidPolicy)
		return false
	}
	if api.pricing == nil {
		writePricingError(w, pricing.ErrUnavailable)
		return false
	}
	policy, err := api.pricing.Current(r.Context(), strings.ToLower(strings.TrimSpace(input.ServiceCode)), "PKR")
	if err != nil {
		writePricingError(w, err)
		return false
	}
	if expected != policy.ID {
		writePricingError(w, pricing.ErrPolicyChanged)
		return false
	}
	if strings.TrimSpace(body.Pickup.PlaceID) != "" || strings.TrimSpace(body.Destination.PlaceID) != "" {
		if !api.allowLocationSearch(userID, w) {
			return false
		}
		if !api.validatePricedPlace(w, r, body.Pickup.PlaceID, input.Pickup) || !api.validatePricedPlace(w, r, body.Destination.PlaceID, input.Destination) {
			return false
		}
	}
	if api.routing == nil {
		api.writeRoutingError(w, r, routing.ErrUnavailable)
		return false
	}
	route, err := api.routing.Preview(r.Context(),
		routing.Endpoint{Point: routing.Point{Latitude: input.Pickup.Latitude, Longitude: input.Pickup.Longitude}, PlaceID: body.Pickup.PlaceID},
		routing.Endpoint{Point: routing.Point{Latitude: input.Destination.Latitude, Longitude: input.Destination.Longitude}, PlaceID: body.Destination.PlaceID})
	if err != nil {
		api.writeRoutingError(w, r, err)
		return false
	}
	amount, err := pricing.Calculate(policy, route.DistanceMeters, route.DurationSeconds)
	if err != nil {
		writePricingError(w, err)
		return false
	}
	if amount != input.ProposedFare.AmountMinor {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "suggested_fare_changed", "message": "Suggested fare changed. Review the refreshed fare and submit again."})
		return false
	}
	return true
}


// A Google Place ID takes precedence over coordinates during routing. Resolve
// the selected ID so the priced endpoint cannot differ from the saved request.
func (api *API) validatePricedPlace(w http.ResponseWriter, r *http.Request, placeID string, location ride.Location) bool {
	placeID = strings.TrimSpace(placeID)
	if placeID == "" {
		return true
	}
	place, err := api.locationSearch.Details(r.Context(), placeID, "")
	if err != nil {
		api.writeLocationSearchError(w, r, err)
		return false
	}
	if !place.Location.Valid() || place.Location.Latitude != location.Latitude || place.Location.Longitude != location.Longitude {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "location_selection_changed", "message": "Selected place changed. Select pickup and destination again."})
		return false
	}
	return true
}
