package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeProvider struct{}
func (fakeProvider) Register(context.Context,Credentials) error{return nil}
func (fakeProvider) Login(context.Context,Credentials)(Session,error){return Session{AccessToken:"token",ExpiresIn:3600},nil}
func (fakeProvider) Refresh(context.Context,string)(Session,error){return Session{AccessToken:"new",ExpiresIn:3600},nil}
func (fakeProvider) Logout(context.Context,string)error{return nil}

func TestLogin(t *testing.T){
	h:=NewHandler(NewService(fakeProvider{}))
	r:=httptest.NewRequest(http.MethodPost,"/",strings.NewReader(`{"identifier":"a","password":"b"}`))
	w:=httptest.NewRecorder(); h.Login(w,r)
	if w.Code!=http.StatusOK{t.Fatalf("got %d",w.Code)}
}
