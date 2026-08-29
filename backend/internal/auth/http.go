package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Handler struct{service Service}
func NewHandler(service Service)Handler{return Handler{service:service}}

func(h Handler)Register(w http.ResponseWriter,r *http.Request){
	var req struct{Identifier string `json:"identifier"`;Password string `json:"password"`}
	if err:=decode(r,&req);err!=nil||strings.TrimSpace(req.Identifier)==""||req.Password==""{failure(w,http.StatusBadRequest,"invalid request");return}
	if err:=h.service.Register(r.Context(),Credentials{Identifier:strings.TrimSpace(req.Identifier),Password:req.Password});err!=nil{providerFailure(w,err,"unable to process your request. Please try again later.");return}
	w.WriteHeader(http.StatusCreated)
}
func(h Handler)Login(w http.ResponseWriter,r *http.Request){
	var req struct{Identifier string `json:"identifier"`;Password string `json:"password"`}
	if err:=decode(r,&req);err!=nil||strings.TrimSpace(req.Identifier)==""||req.Password==""{failure(w,http.StatusBadRequest,"invalid request");return}
	session,err:=h.service.Login(r.Context(),Credentials{Identifier:strings.TrimSpace(req.Identifier),Password:req.Password})
	if err!=nil{providerFailure(w,err,"unable to process your request. Please try again later.");return};write(w,http.StatusOK,session)
}
func(h Handler)Verify(w http.ResponseWriter,r *http.Request){
	var req struct{Email string `json:"email"`}
	if err:=decode(r,&req);err!=nil||strings.TrimSpace(req.Email)==""{failure(w,http.StatusBadRequest,"invalid request");return}
	if err:=h.service.Verify(r.Context(),strings.TrimSpace(req.Email));err!=nil{providerFailure(w,err,"unable to process your request. Please try again later.");return}
	w.WriteHeader(http.StatusNoContent)
}
func(h Handler)Refresh(w http.ResponseWriter,r *http.Request){failure(w,http.StatusNotImplemented,"session refresh is not available")}
func(h Handler)Logout(w http.ResponseWriter,r *http.Request){
	token:=strings.TrimSpace(r.Header.Get("Authorization"));if token==""{failure(w,http.StatusBadRequest,"invalid request");return}
	if err:=h.service.Logout(r.Context(),token);err!=nil&&!errors.Is(err,ErrInvalidCredentials){failure(w,http.StatusServiceUnavailable,"unable to logout");return};w.WriteHeader(http.StatusNoContent)
}
func decode(r *http.Request,v any)error{defer r.Body.Close();d:=json.NewDecoder(io.LimitReader(r.Body,16<<10));d.DisallowUnknownFields();return d.Decode(v)}
func write(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func failure(w http.ResponseWriter,status int,message string){write(w,status,map[string]string{"error":message})}
func providerFailure(w http.ResponseWriter, err error, general string) {
	type clientError interface { ClientError() (int, string) }
	var ce clientError
	if errors.As(err, &ce) {
		status, message := ce.ClientError()
		if status >= 400 && status < 500 {
			failure(w, status, message)
			return
		}
	}
	failure(w, http.StatusInternalServerError, general)
}
