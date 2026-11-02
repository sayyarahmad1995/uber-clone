package httpapi

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "net/http"

 "github.com/google/uuid"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

// PhoneOTPVerifier exchanges a valid single-use SMS code for a short-lived
// registration proof. It must not issue an authenticated session.
type PhoneOTPVerifier interface {
 VerifyAndIssueProof(context.Context,uuid.UUID,string,auth.PhoneChallengePurpose,auth.PhoneOTPCodec)(string,error)
}

// PhoneOTPVerification holds both the Redis store and its HMAC codec.
type PhoneOTPVerification struct {
 Store PhoneOTPVerifier
 Codec auth.PhoneOTPCodec
}
func (v PhoneOTPVerification) Verify(ctx context.Context,id uuid.UUID,code string)(string,error) {
 return v.Store.VerifyAndIssueProof(ctx,id,code,auth.PhoneChallengeSignInOrRegister,v.Codec)
}

type PhoneProofIssuer interface { Verify(context.Context,uuid.UUID,string)(string,error) }

func (api *API) verifyPhoneChallenge(w http.ResponseWriter,r *http.Request) {
 if api.phoneOTPVerifier==nil {http.NotFound(w,r);return}
 var input struct {
  ChallengeID string `json:"challenge_id"`
  Code string `json:"code"`
 }
 decoder:=json.NewDecoder(io.LimitReader(r.Body,16<<10))
 decoder.DisallowUnknownFields()
 if err:=decoder.Decode(&input);err!=nil {writePhoneError(w,http.StatusBadRequest,"invalid_request");return}
 var trailing any
 if err:=decoder.Decode(&trailing);err!=io.EOF {writePhoneError(w,http.StatusBadRequest,"invalid_request");return}
 id,err:=uuid.Parse(input.ChallengeID)
 if err!=nil || len(input.Code)!=6 {writePhoneError(w,http.StatusBadRequest,"invalid_request");return}
 for _,digit:=range input.Code {if digit<'0'||digit>'9' {writePhoneError(w,http.StatusBadRequest,"invalid_request");return}}
 token,err:=api.phoneOTPVerifier.Verify(r.Context(),id,input.Code)
 if err!=nil {
  if errors.Is(err,auth.ErrPhoneChallengeInvalid)||errors.Is(err,auth.ErrPhoneChallengeLimited)||errors.Is(err,auth.ErrPhoneVerificationProof){
   writePhoneError(w,http.StatusBadRequest,"invalid_or_expired_code")
  } else {writePhoneError(w,http.StatusServiceUnavailable,"verification_unavailable")}
  return
 }
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"proof_token":token})
}
