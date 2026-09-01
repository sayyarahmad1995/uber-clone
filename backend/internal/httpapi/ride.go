package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ridestatus"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
)

type rideLocationRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type rideFareRequest struct {
	AmountMinor *int64  `json:"amount_minor"`
	Currency    *string `json:"currency"`
}

type createRideRequestBody struct {
	Pickup       *rideLocationRequest `json:"pickup"`
	Destination  *rideLocationRequest `json:"destination"`
	BookingMode  ride.BookingMode     `json:"booking_mode"`
	ProposedFare *rideFareRequest     `json:"proposed_fare"`
}

func (body createRideRequestBody) input() (ride.CreateInput, bool) {
	if body.Pickup == nil || body.Destination == nil || body.Pickup.Latitude == nil || body.Pickup.Longitude == nil || body.Destination.Latitude == nil || body.Destination.Longitude == nil {
		return ride.CreateInput{}, false
	}
	input := ride.CreateInput{
		Pickup:      ride.Location{Latitude: *body.Pickup.Latitude, Longitude: *body.Pickup.Longitude},
		Destination: ride.Location{Latitude: *body.Destination.Latitude, Longitude: *body.Destination.Longitude},
		BookingMode: body.BookingMode,
	}
	if body.ProposedFare != nil {
		if body.ProposedFare.AmountMinor == nil || body.ProposedFare.Currency == nil {
			return ride.CreateInput{}, false
		}
		input.ProposedFare = &ride.Money{AmountMinor: *body.ProposedFare.AmountMinor, Currency: *body.ProposedFare.Currency}
	}
	return input, true
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup, destination, and complete fare fields are required when provided"})
		return
	}
	request, err := api.rides.Create(r.Context(), u.ID, input)
	switch {
	case errors.Is(err, ride.ErrInvalidLocation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pickup and destination coordinates are invalid"})
		return
	case errors.Is(err, ride.ErrInvalidBookingMode):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "booking_mode must be automatic or offers"})
		return
	case errors.Is(err, ride.ErrInvalidFare):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "offers booking requires a positive proposed_fare with a three-letter currency"})
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
		"id":            request.ID,
		"rider_user_id": request.RiderUserID,
		"pickup":        map[string]any{"latitude": request.Pickup.Latitude, "longitude": request.Pickup.Longitude},
		"destination":   map[string]any{"latitude": request.Destination.Latitude, "longitude": request.Destination.Longitude},
		"booking_mode":  request.BookingMode,
		"status":        request.Status,
		"created_at":    request.CreatedAt,
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
		"id":           request.ID,
		"pickup":       map[string]any{"latitude": request.Pickup.Latitude, "longitude": request.Pickup.Longitude},
		"destination":  map[string]any{"latitude": request.Destination.Latitude, "longitude": request.Destination.Longitude},
		"booking_mode": request.BookingMode,
		"status":       request.Status,
		"created_at":   request.CreatedAt,
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
			"driver_user_id": assignedTrip.DriverUserID,
			"status":         assignedTrip.Status,
			"assigned_at":    assignedTrip.AssignedAt,
			"started_at":     assignedTrip.StartedAt,
			"completed_at":   assignedTrip.CompletedAt,
			"cancelled_at":   assignedTrip.CancelledAt,
		}
	}
	return response
}
