package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

// registerPhone is registered only when explicitly configured. OTP issuance is
// intentionally not exposed until delivery and abuse controls are operational.
func (api *API) registerPhone(w http.ResponseWriter, r *http.Request) {
	if api.phoneRegistration == nil {
		http.NotFound(w, r)
		return
	}
	var body struct {
		ProofToken string `json:"proof_token"`
		Password   string `json:"password"`
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 16384))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil || body.ProofToken == "" || body.Password == "" {
		writePhoneError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writePhoneError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	id, err := api.phoneRegistration.Register(r.Context(), body.ProofToken, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrPhoneVerificationProof):
			writePhoneError(w, http.StatusBadRequest, "invalid_or_expired_proof")
		case errors.Is(err, auth.ErrRegistrationInvalid):
			writePhoneError(w, http.StatusBadRequest, "invalid_request")
		case errors.Is(err, auth.ErrPhoneRegistrationPending):
			writePhoneError(w, http.StatusConflict, "registration_pending")
		default:
			writePhoneError(w, http.StatusServiceUnavailable, "registration_unavailable")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"user_id": id.String()})
}

func writePhoneError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
