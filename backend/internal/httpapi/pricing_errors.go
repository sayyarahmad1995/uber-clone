package httpapi

import (
	"errors"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"net/http"
)

func writePricingError(w http.ResponseWriter, err error) {
	status, code, message := 500, "pricing_failed", "Unable to load pricing."
	switch {
	case errors.Is(err, pricing.ErrPolicyChanged):
		status, code, message = 409, "pricing_policy_changed", "Rates changed. Refresh the suggestion and submit again."
	case errors.Is(err, pricing.ErrUnavailable):
		status, code, message = 503, "pricing_policy_unavailable", "Suggested fare is unavailable for this service."
	case errors.Is(err, pricing.ErrInvalidService):
		status, code, message = 400, "invalid_pricing_service", "Selected service is unavailable."
	case errors.Is(err, pricing.ErrInvalidPolicy):
		status, code, message = 400, "invalid_pricing_policy", "A valid PKR pricing policy version is required."
	case errors.Is(err, pricing.ErrInvalidRoute), errors.Is(err, pricing.ErrFareOutOfRange):
		status, code, message = 503, "pricing_unavailable", "Unable to calculate a suggested fare for this route."
	}
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
