package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

// PhoneChallengeIssuer creates a challenge and an encrypted SMS dispatch
// atomically. Implementations must not send an SMS from the HTTP handler.
type PhoneChallengeIssuer interface {
	Issue(context.Context, auth.PhoneChallengeRequest, string, string) (auth.PhoneChallengeCreation, error)
}

// issuePhoneChallenge is a tightly gated development integration route.
// It must never accept arbitrary destination numbers until the production
// SMS abuse-prevention and trusted-proxy policies have been reviewed.
func (api *API) issuePhoneChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if api.phoneOTPIssuer == nil || len(api.phoneSMSTestAllowedNumbers) == 0 {
		http.NotFound(w, r)
		return
	}

	var input auth.PhoneChallengeRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writePhoneError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writePhoneError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	normalized, err := input.Validate()
	if err != nil {
		writePhoneError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	// Explicit allowlist prevents accidental sends to unapproved numbers
	// during early real-device testing.
	if _, ok := api.phoneSMSTestAllowedNumbers[normalized.PhoneNumber]; !ok {
		writePhoneError(w, http.StatusForbidden, "test_destination_not_allowed")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 128 {
		writePhoneError(w, http.StatusBadRequest, "invalid_idempotency_key")
		return
	}

	// RemoteAddr is set by net/http and is not client-controlled.
	// X-Forwarded-For is intentionally not trusted. When deploying behind a
	// proxy, review the source extraction boundary before broad enablement.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || net.ParseIP(host) == nil {
		writePhoneError(w, http.StatusServiceUnavailable, "issuance_unavailable")
		return
	}

	created, err := api.phoneOTPIssuer.Issue(
		r.Context(), normalized, host, idempotencyKey,
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrPhoneChallengeCooldown):
			w.Header().Set("Retry-After", strconv.Itoa(remainingSeconds(created.ResendAfter)))
			writePhoneError(w, http.StatusTooManyRequests, "resend_cooldown")
		case errors.Is(err, auth.ErrPhoneOTPRateLimited):
			// The underlying limiter uses one-hour windows. This is a
			// conservative upper-bound retry hint, not an exact TTL.
			w.Header().Set("Retry-After", "3600")
			writePhoneError(w, http.StatusTooManyRequests, "rate_limited")
		case errors.Is(err, auth.ErrInvalidPhoneChallengeRequest):
			writePhoneError(w, http.StatusBadRequest, "invalid_request")
		default:
			// Never disclose Redis, dispatch, provider or account details.
			writePhoneError(w, http.StatusServiceUnavailable, "issuance_unavailable")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(struct {
		ChallengeID        string `json:"challenge_id"`
		ResendAfterSeconds int    `json:"resend_after_seconds"`
	}{
		ChallengeID:        created.ID.String(),
		ResendAfterSeconds: remainingSeconds(created.ResendAfter),
	})
}

func remainingSeconds(deadline time.Time) int {
	seconds := int(math.Ceil(time.Until(deadline).Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
