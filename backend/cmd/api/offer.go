package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
)

type submitRideOfferBody struct {
	AmountMinor *int64 `json:"amount_minor"`
}

func (app application) submitRideOffer(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}

	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}

	var body submitRideOfferBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AmountMinor == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "amount_minor is required"})
		return
	}

	result, err := app.offers.Submit(r.Context(), rideRequestID, u.ID, *body.AmountMinor)
	if writeOfferError(w, err) {
		return
	}

	response := rideOfferResponse(result.Offer)
	if result.Trip != nil {
		response["trip"] = tripResponse(*result.Trip)
	}
	writeJSON(w, http.StatusOK, response)
}

func (app application) listRideOffers(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok {
		return
	}

	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return
	}

	results, err := app.offers.ListForRider(r.Context(), rideRequestID, u.ID)
	if writeOfferError(w, err) {
		return
	}

	items := make([]map[string]any, 0, len(results))
	for _, result := range results {
		items = append(items, rideOfferResponse(result))
	}
	writeJSON(w, http.StatusOK, map[string]any{"offers": items})
}

func (app application) acceptRideOffer(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, driverUserID, ok := parseOfferPath(w, r)
	if !ok {
		return
	}
	result, err := app.offers.Accept(r.Context(), rideRequestID, u.ID, driverUserID)
	if writeOfferError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"trip": tripResponse(result)})
}

func (app application) rejectRideOffer(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireRiderCapability(w, r)
	if !ok {
		return
	}
	rideRequestID, driverUserID, ok := parseOfferPath(w, r)
	if !ok {
		return
	}
	result, err := app.offers.Reject(r.Context(), rideRequestID, u.ID, driverUserID)
	if writeOfferError(w, err) {
		return
	}
	writeRideOffer(w, http.StatusOK, result)
}

func parseOfferPath(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	rideRequestID, err := uuid.Parse(r.PathValue("ride_request_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid ride_request_id"})
		return uuid.Nil, uuid.Nil, false
	}
	driverUserID, err := uuid.Parse(r.PathValue("driver_user_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid driver_user_id"})
		return uuid.Nil, uuid.Nil, false
	}
	return rideRequestID, driverUserID, true
}

func writeOfferError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, offer.ErrRideNotFound), errors.Is(err, offer.ErrOfferNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request or offer not found"})
	case errors.Is(err, offer.ErrRideNotOpen):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "ride request is not open for offers"})
	case errors.Is(err, offer.ErrOfferNotActionable):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "ride offer is not actionable"})
	case errors.Is(err, offer.ErrDriverIneligible):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "driver is not eligible to offer or assign"})
	case errors.Is(err, offer.ErrAmountOutOfRange):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "offer must be between 90% and 130% of the rider proposed fare"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to process ride offer"})
	}
	return true
}

func writeRideOffer(w http.ResponseWriter, status int, result offer.Offer) {
	writeJSON(w, status, rideOfferResponse(result))
}

func rideOfferResponse(result offer.Offer) map[string]any {
	return map[string]any{
		"ride_request_id": result.RideRequestID,
		"driver_user_id":  result.DriverUserID,
		"fare": map[string]any{
			"amount_minor": result.AmountMinor,
			"currency":     result.Currency,
		},
		"status":     result.Status,
		"created_at": result.CreatedAt,
		"updated_at": result.UpdatedAt,
		"decided_at": result.DecidedAt,
	}
}
