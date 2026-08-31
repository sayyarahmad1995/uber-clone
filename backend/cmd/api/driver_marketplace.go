package main

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/offer"
)

func (app application) discoverDriverMarketplace(w http.ResponseWriter, r *http.Request) {
	u, ok := app.requireDriverCapability(w, r)
	if !ok {
		return
	}

	items, err := app.offers.Discover(r.Context(), u.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to discover marketplace ride requests"})
		return
	}

	response := make([]map[string]any, 0, len(items))
	for _, item := range items {
		response = append(response, driverMarketplaceItemResponse(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"ride_requests": response})
}

func driverMarketplaceItemResponse(item offer.DiscoveryItem) map[string]any {
	response := map[string]any{
		"id": item.RideRequestID,
		"pickup": map[string]float64{
			"latitude":  item.Pickup.Latitude,
			"longitude": item.Pickup.Longitude,
		},
		"destination": map[string]float64{
			"latitude":  item.Destination.Latitude,
			"longitude": item.Destination.Longitude,
		},
		"proposed_fare": map[string]any{
			"amount_minor": item.ProposedFare.ProposedAmountMinor,
			"currency":     item.ProposedFare.Currency,
		},
		"created_at": item.CreatedAt,
		"own_offer":  nil,
	}
	if item.OwnOffer != nil {
		response["own_offer"] = rideOfferResponse(*item.OwnOffer)
	}
	return response
}
