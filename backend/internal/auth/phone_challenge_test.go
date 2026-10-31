package auth

import (
	"errors"
	"testing"
)

func TestPhoneChallengeRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		purpose PhoneChallengePurpose
		valid   bool
	}{
		{"pakistan e164", "+923001234567", PhoneChallengeSignInOrRegister, true},
		{"trims whitespace", "  +923001234567  ", PhoneChallengeSignInOrRegister, true},
		{"minimum digits", "+12345678", PhoneChallengeSignInOrRegister, true},
		{"maximum digits", "+123456789012345", PhoneChallengeSignInOrRegister, true},
		{"missing plus", "923001234567", PhoneChallengeSignInOrRegister, false},
		{"local phone", "03001234567", PhoneChallengeSignInOrRegister, false},
		{"leading country zero", "+023001234567", PhoneChallengeSignInOrRegister, false},
		{"too short", "+1234567", PhoneChallengeSignInOrRegister, false},
		{"too long", "+1234567890123456", PhoneChallengeSignInOrRegister, false},
		{"separators", "+92 3001234567", PhoneChallengeSignInOrRegister, false},
		{"letters", "+92300abc4567", PhoneChallengeSignInOrRegister, false},
		{"unicode digits", "+923۰۰۱۲۳۴۵۶۷", PhoneChallengeSignInOrRegister, false},
		{"unapproved purpose", "+923001234567", "replace_verified_email", false},
		{"no purpose", "+923001234567", "", false},
		{"no phone", "", PhoneChallengeSignInOrRegister, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (PhoneChallengeRequest{PhoneNumber: tt.phone, Purpose: tt.purpose}).Validate()
			if tt.valid {
				if err != nil {
					t.Fatalf("unexpected validation error: %v", err)
				}
				if got.PhoneNumber == "" || got.Purpose != PhoneChallengeSignInOrRegister {
					t.Fatalf("unexpected validated value: %+v", got)
				}
			} else if !errors.Is(err, ErrInvalidPhoneChallengeRequest) {
				t.Fatalf("got error %v, want ErrInvalidPhoneChallengeRequest", err)
			}
		})
	}
}
