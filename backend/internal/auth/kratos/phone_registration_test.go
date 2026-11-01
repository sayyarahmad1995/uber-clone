package kratos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

func TestRegisterPhoneIdentityNoEmailPlaceholder(t *testing.T) {
	posted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		w.Header().Set("Content-Type","application/json")
		switch r.URL.Path {
		case "/self-service/registration/api":
			_,_=w.Write([]byte(`{"id":"flow-1","ui":{"method":"POST"}}`))
		case "/self-service/registration":
			if r.URL.Query().Get("flow")!="flow-1" {t.Errorf("missing flow")}
			var body struct {
				Method string `json:"method"`
				Traits map[string]string `json:"traits"`
			}
			if err:=json.NewDecoder(r.Body).Decode(&body);err!=nil {t.Error(err)}
			if body.Method!="password"||body.Traits["phone"]!="+923001234567"||len(body.Traits)!=1 {
				t.Errorf("unexpected registration traits: %+v",body)
			}
			posted=true
			_,_=w.Write([]byte(`{"identity":{"id":"kratos-registered-identity"}}`))
		default:
			t.Errorf("unexpected endpoint: %s",r.URL.Path)
		}
	}))
	defer server.Close()
	p:=&Provider{baseURL:server.URL,client:server.Client()}
	created,err:=p.RegisterPhoneIdentity(context.Background(),"+923001234567","valid-test-password")
	if err!=nil||created.IdentitySubject!="kratos-registered-identity"||!posted {
		t.Fatalf("registration: %+v %v posted=%v",created,err,posted)
	}
}

func TestRegisterPhoneIdentityFailsWithoutIdentityID(t *testing.T) {
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		w.Header().Set("Content-Type","application/json")
		if r.URL.Path=="/self-service/registration/api" {
			_,_=w.Write([]byte(`{"id":"flow-1"}`))
		} else {
			_,_=w.Write([]byte(`{"continue_with":[]}`))
		}
	}))
	defer server.Close()
	p:=&Provider{baseURL:server.URL,client:server.Client()}
	_,err:=p.RegisterPhoneIdentity(context.Background(),"+923001234567","valid-test-password")
	if !errors.Is(err,auth.ErrUnavailable){t.Fatalf("missing Kratos identity accepted: %v",err)}
}
