package auth

import (
	"errors"
	"strings"
)

var ErrInvalidPhoneChallengeRequest = errors.New("invalid phone challenge request")

// PhoneChallengePurpose is limited to explicitly supported phone auth flows.
// Account recovery and verified-email changes need separate, purpose-bound
// challenge policies before they can be issued.
type PhoneChallengePurpose string

const PhoneChallengeSignInOrRegister PhoneChallengePurpose = "sign_in_or_register"

// PhoneChallengeRequest is the public request shape for issuing a phone OTP.
// No handler is registered until durable challenge storage, an SMS provider,
// rate limits and idempotent issuing are implemented.
type PhoneChallengeRequest struct {
	PhoneNumber string                `json:"phone_number"`
	Purpose     PhoneChallengePurpose `json:"purpose"`
}

// Validate returns a normalized request or an error. This checks E.164 syntax,
// not whether a national numbering plan assigns or routes the number.
// Full phone-number validation and proof of possession remain mandatory.
func (r PhoneChallengeRequest) Validate() (PhoneChallengeRequest, error) {
	phone := strings.TrimSpace(r.PhoneNumber)
	if r.Purpose != PhoneChallengeSignInOrRegister || len(phone) < 9 || len(phone) > 16 || phone[0] != '+' {
		return PhoneChallengeRequest{}, ErrInvalidPhoneChallengeRequest
	}
	if phone[1] < '1' || phone[1] > '9' {
		return PhoneChallengeRequest{}, ErrInvalidPhoneChallengeRequest
	}
	for i := 2; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return PhoneChallengeRequest{}, ErrInvalidPhoneChallengeRequest
		}
	}
	return PhoneChallengeRequest{PhoneNumber: phone, Purpose: r.Purpose}, nil
}
