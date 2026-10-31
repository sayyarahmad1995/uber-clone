package auth

import (
	"context"
	"errors"
	"strings"
)

// PhoneAccountBinding is durable, authoritative account state. Implementations
// must query the account database, not infer ownership from a submitted phone
// number or from Redis OTP state. An unbound phone returns Exists=false.
type PhoneAccountBinding struct {
	Exists bool
	IdentitySubject string
	VerifiedPhone string
}

// PhoneAccountResolver is deliberately small: it is the boundary between
// ephemeral proof and durable account-to-Kratos identity linkage.
type PhoneAccountResolver interface {
	ResolveVerifiedPhone(context.Context, string) (PhoneAccountBinding, error)
}

// PhoneAuthNextStep is returned only after an OTP has been verified.
// Neither step authorizes a Kratos session.
type PhoneAuthNextStep string

const (
	PhoneAuthLogin PhoneAuthNextStep = "login"
	PhoneAuthRegister PhoneAuthNextStep = "register"
)

// PhoneAuthDecision represents verified possession and an account-state lookup.
// An existing user still needs valid Kratos credentials to log in.
type PhoneAuthDecision struct {
	NextStep PhoneAuthNextStep
	VerifiedPhone string
	IdentitySubject string // server-side only; never return to an unauthenticated client
}

var ErrPhoneIdentityBinding = errors.New("verified phone account binding is inconsistent")

type PhoneAccountChecker struct {
	resolver PhoneAccountResolver
}

func NewPhoneAccountChecker(resolver PhoneAccountResolver) (PhoneAccountChecker,error) {
	if resolver==nil {return PhoneAccountChecker{},ErrPhoneOTPConfiguration}
	return PhoneAccountChecker{resolver:resolver},nil
}

// Decide accepts the phone obtained from a successfully validated OTP, never
// from client request JSON. It does not consume or authorize a session.
// Public handlers must not expose IdentitySubject to the client.
func (s PhoneAccountChecker) Decide(ctx context.Context, verifiedPhone string) (PhoneAuthDecision,error) {
	normalized,err:=(PhoneChallengeRequest{PhoneNumber:verifiedPhone,Purpose:PhoneChallengeSignInOrRegister}).Validate()
	if err!=nil{return PhoneAuthDecision{},ErrPhoneVerificationProof}
	binding,err:=s.resolver.ResolveVerifiedPhone(ctx,normalized.PhoneNumber)
	if err!=nil{return PhoneAuthDecision{},err}
	if !binding.Exists {
		if binding.IdentitySubject!=""||binding.VerifiedPhone!="" {return PhoneAuthDecision{},ErrPhoneIdentityBinding}
		return PhoneAuthDecision{NextStep:PhoneAuthRegister,VerifiedPhone:normalized.PhoneNumber},nil
	}
	if strings.TrimSpace(binding.IdentitySubject)==""||binding.VerifiedPhone!=normalized.PhoneNumber {
		return PhoneAuthDecision{},ErrPhoneIdentityBinding
	}
	return PhoneAuthDecision{
		NextStep:PhoneAuthLogin,VerifiedPhone:normalized.PhoneNumber,
		IdentitySubject:binding.IdentitySubject,
	},nil
}
