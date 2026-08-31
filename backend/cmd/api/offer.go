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
	writeRideOffer(w, http.StatusOK, result)
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

func writeOfferError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, offer.ErrRideNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ride request not found"})
	case errors.Is(err, offer.ErrRideNotOpen):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "ride request is not open for offers"})
	case errors.Is(err, offer.ErrDriverIneligible):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "driver is not eligible to offer"})
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
		"created_at": result.CreatedAt,
		"updated_at": result.UpdatedAt,
	}
}
