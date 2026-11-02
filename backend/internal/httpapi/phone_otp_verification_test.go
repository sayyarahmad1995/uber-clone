package httpapi

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/google/uuid"
 "github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

type stubProofIssuer struct{ calls int; err error }
func(s *stubProofIssuer) Verify(_ context.Context,_ uuid.UUID,_ string)(string,error) {
 s.calls++
 if s.err!=nil{return "",s.err}
 return "opaque-proof",nil
}
func TestPhoneOTPVerificationRouteDisabledByDefault(t *testing.T) {
 response:=httptest.NewRecorder()
 New(Dependencies{}).Handler().ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/v1/auth/phone/challenges/verify",nil))
 if response.Code!=http.StatusNotFound {t.Fatalf("unexpected status %d",response.Code)}
}
func TestPhoneOTPVerificationRejectsMalformedInputs(t *testing.T) {
 verifier:=&stubProofIssuer{}
 api:=New(Dependencies{PhoneRegistrationEnabled:true,PhoneRegistration:&auth.PhoneRegistrationCoordinator{},PhoneOTPVerifier:verifier})
 for _,body:=range []string{`{}`,`{"challenge_id":"broken","code":"123456"}`,`{"challenge_id":"fd19cba6-d5f9-4c1a-b5e4-81c191baae4f","code":"12345x"}`,`{"challenge_id":"fd19cba6-d5f9-4c1a-b5e4-81c191baae4f","code":"123456","extra":true}`} {
  response:=httptest.NewRecorder()
  api.Handler().ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/v1/auth/phone/challenges/verify",strings.NewReader(body)))
  if response.Code!=http.StatusBadRequest {t.Fatalf("request %s returned %d",body,response.Code)}
 }
 if verifier.calls!=0{t.Fatal("invalid code reached proof issuer")}
}
func TestPhoneOTPVerificationResponseAndUnavailable(t *testing.T) {
 verifier:=&stubProofIssuer{}
 api:=New(Dependencies{PhoneRegistrationEnabled:true,PhoneRegistration:&auth.PhoneRegistrationCoordinator{},PhoneOTPVerifier:verifier})
 body:=`{"challenge_id":"fd19cba6-d5f9-4c1a-b5e4-81c191baae4f","code":"123456"}`
 response:=httptest.NewRecorder()
 api.Handler().ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/v1/auth/phone/challenges/verify",strings.NewReader(body)))
 if response.Code!=http.StatusOK||response.Header().Get("Cache-Control")!="no-store"||!strings.Contains(response.Body.String(),"opaque-proof"){t.Fatalf("unexpected proof response %d %s",response.Code,response.Body.String())}
 verifier.err=errors.New("redis unavailable")
 response=httptest.NewRecorder()
 api.Handler().ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/v1/auth/phone/challenges/verify",strings.NewReader(body)))
 if response.Code!=http.StatusServiceUnavailable||strings.Contains(response.Body.String(),"redis unavailable"){t.Fatalf("leaked internal error: %d %s",response.Code,response.Body.String())}
}
