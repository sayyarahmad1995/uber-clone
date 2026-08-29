package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct{service Service}
func NewHandler(service Service)Handler{return Handler{service:service}}

func(h Handler)Register(w http.ResponseWriter,r *http.Request){
	var req struct{Identifier string `json:"identifier"`;Password string `json:"password"`}
	if err:=decode(r,&req);err!=nil||strings.TrimSpace(req.Identifier)==""||req.Password==""{failure(w,http.StatusBadRequest,"invalid request");return}
	if err:=h.service.Register(r.Context(),Credentials{Identifier:strings.TrimSpace(req.Identifier),Password:req.Password});err!=nil{failure(w,http.StatusBadRequest,"unable to complete registration");return}
	w.WriteHeader(http.StatusCreated)
}
func(h Handler)Login(w http.ResponseWriter,r *http.Request){
	var req struct{Identifier string `json:"identifier"`;Password string `json:"password"`}
	if err:=decode(r,&req);err!=nil||strings.TrimSpace(req.Identifier)==""||req.Password==""{failure(w,http.StatusBadRequest,"invalid request");return}
	session,err:=h.service.Login(r.Context(),Credentials{Identifier:strings.TrimSpace(req.Identifier),Password:req.Password})
	if err!=nil{failure(w,http.StatusUnauthorized,"invalid credentials");return};write(w,http.StatusOK,session)
}
func(h Handler)Refresh(w http.ResponseWriter,r *http.Request){failure(w,http.StatusNotImplemented,"session refresh is not available")}
func(h Handler)Logout(w http.ResponseWriter,r *http.Request){
	token:=strings.TrimSpace(r.Header.Get("Authorization"));if token==""{failure(w,http.StatusBadRequest,"invalid request");return}
	if err:=h.service.Logout(r.Context(),token);err!=nil&&!errors.Is(err,ErrInvalidCredentials){failure(w,http.StatusServiceUnavailable,"unable to logout");return};w.WriteHeader(http.StatusNoContent)
}
func decode(r *http.Request,v any)error{r.Body=http.MaxBytesReader(writableResponseWriter{r},r.Body,16<<10);defer r.Body.Close();d:=json.NewDecoder(r.Body);d.DisallowUnknownFields();return d.Decode(v)}
type writableResponseWriter struct{*http.Request}
func(w writableResponseWriter)Header()http.Header{return http.Header{}}
func(w writableResponseWriter)Write([]byte)(int,error){return 0,nil}
func(w writableResponseWriter)WriteHeader(int){}
func write(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(v)}
func failure(w http.ResponseWriter,status int,message string){write(w,status,map[string]string{"error":message})}
