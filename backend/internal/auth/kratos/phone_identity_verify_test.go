package kratos

import (
 "context"
 "net/http"
 "net/http/httptest"
 "testing"
)

func TestVerifyPhoneIdentityUsesKratosAdmin(t *testing.T){
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/admin/identities/identity-1" {t.Errorf("unexpected URL %s",r.URL.Path)}
  w.Header().Set("Content-Type","application/json")
  _,_=w.Write([]byte(`{"id":"identity-1","traits":{"phone":"+923001234567"}}`))
 }))
 defer server.Close()
 provider:=&Provider{adminURL:server.URL,client:server.Client()}
 matched,err:=provider.VerifyPhoneIdentity(context.Background(),"identity-1","+923001234567")
 if err!=nil||!matched{t.Fatalf("valid: %v %v",matched,err)}
 matched,err=provider.VerifyPhoneIdentity(context.Background(),"identity-1","+923001234568")
 if err!=nil||matched{t.Fatalf("phone mismatch accepted: %v %v",matched,err)}
}
