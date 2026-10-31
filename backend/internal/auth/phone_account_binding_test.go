package auth

import (
	"context"
	"errors"
	"testing"
)

type stubPhoneResolver struct {
	got     string
	binding PhoneAccountBinding
	err     error
}

func (s *stubPhoneResolver) ResolveVerifiedPhone(_ context.Context, phone string) (PhoneAccountBinding, error) {
	s.got = phone
	return s.binding, s.err
}

func TestPhoneAccountCheckerRoutesOnlyAuthoritativeBindings(t *testing.T) {
	for _, tt := range []struct {
		name    string
		binding PhoneAccountBinding
		want    PhoneAuthNextStep
	}{
		{"new account", PhoneAccountBinding{}, PhoneAuthRegister},
		{"existing account", PhoneAccountBinding{Exists: true, IdentitySubject: "kratos-identity-id", VerifiedPhone: "+923001234567"}, PhoneAuthLogin},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := &stubPhoneResolver{binding: tt.binding}
			checker, err := NewPhoneAccountChecker(r)
			if err != nil {
				t.Fatal(err)
			}
			got, err := checker.Decide(context.Background(), "+923001234567")
			if err != nil || got.NextStep != tt.want || got.VerifiedPhone != "+923001234567" || r.got != "+923001234567" {
				t.Fatalf("decision: %+v %v; resolver phone=%q", got, err, r.got)
			}
		})
	}
}

func TestPhoneAccountCheckerRejectsInconsistentBinding(t *testing.T) {
	for _, binding := range []PhoneAccountBinding{
		{Exists: true, IdentitySubject: "kratos-id"},
		{Exists: true, VerifiedPhone: "+923001234567"},
		{Exists: true, IdentitySubject: "kratos-id", VerifiedPhone: "+923001234568"},
		{Exists: false, IdentitySubject: "kratos-id"},
	} {
		checker, err := NewPhoneAccountChecker(&stubPhoneResolver{binding: binding})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := checker.Decide(context.Background(), "+923001234567"); !errors.Is(err, ErrPhoneIdentityBinding) {
			t.Fatalf("binding %+v accepted: %v", binding, err)
		}
	}
}

func TestPhoneAccountCheckerPropagatesLookupFailure(t *testing.T) {
	failure := errors.New("database unavailable")
	r := &stubPhoneResolver{err: failure}
	checker, err := NewPhoneAccountChecker(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := checker.Decide(context.Background(), "+923001234567"); !errors.Is(err, failure) {
		t.Fatalf("fail-open: %v", err)
	}
	if _, err := checker.Decide(context.Background(), "unverified"); !errors.Is(err, ErrPhoneVerificationProof) {
		t.Fatalf("invalid phone: %v", err)
	}
}
