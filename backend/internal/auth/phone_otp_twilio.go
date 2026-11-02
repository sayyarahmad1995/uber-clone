package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ErrPhoneOTPSendUncertain means Twilio might have accepted the message.
// It must not be retried automatically: Messages.create has no verified
// challenge-UUID idempotency contract.
var ErrPhoneOTPSendUncertain = errors.New("SMS provider submission outcome uncertain")

// TwilioOTPSender sends HiGO-generated codes using Programmable Messaging.
// It does not generate, store or verify OTPs and is not wired into startup.
type TwilioOTPSender struct {
	client *http.Client
	endpoint string
	accountSID string
	apiKeySID string
	apiKeySecret string
	sender string
	useMessagingService bool
}

// NewTwilioOTPSender expects one sender identity: a phone number OR a
// Messaging Service SID. Credentials are server-only and never logged.
func NewTwilioOTPSender(client *http.Client, accountSID, apiKeySID, apiKeySecret, sender string, messagingService bool) (TwilioOTPSender, error) {
	if client == nil || !strings.HasPrefix(accountSID, "AC") || len(accountSID) != 34 ||
		apiKeySID == "" || apiKeySecret == "" || sender == "" ||
		(messagingService && (!strings.HasPrefix(sender, "MG") || len(sender) != 34)) ||
		(!messagingService && !strings.HasPrefix(sender, "+")) {
		return TwilioOTPSender{}, ErrPhoneOTPConfiguration
	}
	return TwilioOTPSender{
		client: client, endpoint: "https://api.twilio.com/2010-04-01/Accounts/" + accountSID + "/Messages.json",
		accountSID: accountSID, apiKeySID: apiKeySID, apiKeySecret: apiKeySecret,
		sender: sender, useMessagingService: messagingService,
	}, nil
}

// SendPhoneOTP submits at most one HTTP request. The challenge ID stays in
// HiGO's Redis queue; Twilio does not promise to deduplicate by that ID.
func (s TwilioOTPSender) SendPhoneOTP(ctx context.Context, id uuid.UUID, phone, code string) error {
	if id == uuid.Nil || len(code) != 6 || s.client == nil {
		return ErrInvalidPhoneChallengeRequest
	}
	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return ErrPhoneOTPCode
		}
	}
	if _, err := (PhoneChallengeRequest{PhoneNumber: phone, Purpose: PhoneChallengeSignInOrRegister}).Validate(); err != nil {
		return err
	}
	form := url.Values{"To": {phone}, "Body": {fmt.Sprintf("Your HiGO verification code is %s. It expires in 5 minutes.", code)}}
	if s.useMessagingService {
		form.Set("MessagingServiceSid", s.sender)
	} else {
		form.Set("From", s.sender)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.apiKeySID, s.apiKeySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		// The server might have accepted the request before a network error.
		return fmt.Errorf("%w: %v", ErrPhoneOTPSendUncertain, err)
	}
	defer resp.Body.Close()
	// A successful response must contain a recognized Twilio Message SID.
	if resp.StatusCode == http.StatusCreated {
		var result struct { SID string `json:"sid"` }
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result); err != nil ||
			len(result.SID) != 34 || (!strings.HasPrefix(result.SID, "SM") && !strings.HasPrefix(result.SID, "MM")) {
			return ErrPhoneOTPSendUncertain
		}
		return nil
	}
	// Never return response bodies: Twilio may echo phone or message details.
	// Explicit 4xx rejection is not an accepted message. Other responses
	// can represent an ambiguous submission and must not be replayed.
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("Twilio rejected SMS submission (HTTP %d)", resp.StatusCode)
	}
	return ErrPhoneOTPSendUncertain
}
