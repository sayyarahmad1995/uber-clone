package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
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
		Pickup: ride.Location{Latitude: *body.Pickup.Latitude, Longitude: *body.Pickup.Longitude},
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

func (app application) createRideRequest(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok { return }
	var body createRideRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid request"}); return
	}
	input, ok := body.input()
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"pickup, destination, and complete fare fields are required when provided"}); return
	}
	request, err := app.rides.Create(r.Context(), u.ID, input)
	switch {
	case errors.Is(err, ride.ErrInvalidLocation): writeJSON(w,http.StatusBadRequest,map[string]string{"error":"pickup and destination coordinates are invalid"}); return
	case errors.Is(err, ride.ErrInvalidBookingMode): writeJSON(w,http.StatusBadRequest,map[string]string{"error":"booking_mode must be automatic or offers"}); return
	case errors.Is(err, ride.ErrInvalidFare): writeJSON(w,http.StatusBadRequest,map[string]string{"error":"offers booking requires a positive proposed_fare with a three-letter currency"}); return
	case err != nil: writeJSON(w,http.StatusInternalServerError,map[string]string{"error":"unable to create ride request"}); return
	}
	writeRideRequest(w,http.StatusCreated,request)
}

func (app application) requireRiderCapability(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	u, ok := app.currentUser(w,r)
	if !ok { return user.User{}, false }
	for _, capability := range u.Capabilities { if capability == user.CapabilityRider { return u,true } }
	writeJSON(w,http.StatusForbidden,map[string]string{"error":"rider capability required"})
	return user.User{},false
}

func writeRideRequest(w http.ResponseWriter, status int, request ride.Request) {
	response := map[string]any{
		"id":request.ID,"rider_user_id":request.RiderUserID,
		"pickup":map[string]any{"latitude":request.Pickup.Latitude,"longitude":request.Pickup.Longitude},
		"destination":map[string]any{"latitude":request.Destination.Latitude,"longitude":request.Destination.Longitude},
		"booking_mode":request.BookingMode,"status":request.Status,"created_at":request.CreatedAt,
	}
	if request.ProposedFare != nil { response["proposed_fare"] = map[string]any{"amount_minor":request.ProposedFare.AmountMinor,"currency":request.ProposedFare.Currency} }
	writeJSON(w,status,response)
}
