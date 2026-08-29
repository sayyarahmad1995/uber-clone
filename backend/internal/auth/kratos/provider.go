package kratos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

type Provider struct { baseURL string; client *http.Client }

func New(baseURL string) (*Provider,error) {
	if strings.TrimSpace(baseURL)=="" { return nil,errors.New("Kratos URL is required") }
	return &Provider{baseURL:strings.TrimRight(baseURL,"/"),client:&http.Client{Timeout:10*time.Second}},nil
}

func (p *Provider) Register(ctx context.Context,c auth.Credentials) error {
	flow,err:=p.createFlow(ctx,"/self-service/registration/api"); if err!=nil{return err}
	body:=map[string]any{"method":"password","password":c.Password,"traits":map[string]string{"email":c.Identifier}}
	return p.submit(ctx,flow.Action,body,nil)
}

func (p *Provider) Login(ctx context.Context,c auth.Credentials)(auth.Session,error) {
	flow,err:=p.createFlow(ctx,"/self-service/login/api"); if err!=nil{return auth.Session{},auth.ErrUnavailable}
	body:=map[string]any{"method":"password","identifier":c.Identifier,"password":c.Password}
	var result struct { SessionToken string `json:"session_token"`; Session struct { ExpiresAt time.Time `json:"expires_at"` } `json:"session"` }
	if err:=p.submit(ctx,flow.Action,body,&result);err!=nil{return auth.Session{},auth.ErrInvalidCredentials}
	if result.SessionToken==""{return auth.Session{},auth.ErrInvalidCredentials}
	return auth.Session{AccessToken:result.SessionToken,ExpiresIn:max(1,int64(time.Until(result.Session.ExpiresAt).Seconds()))},nil
}

func (p *Provider) Refresh(context.Context,string)(auth.Session,error) {
	// Kratos session tokens are not refresh tokens. Refresh is intentionally deferred
	// until the application introduces its own refresh-token/session strategy.
	return auth.Session{},auth.ErrUnavailable
}

func (p *Provider) Logout(ctx context.Context,token string) error {
	token=strings.TrimSpace(strings.TrimPrefix(token,"Bearer "))
	if token=="" { return nil }
	req,err:=http.NewRequestWithContext(ctx,http.MethodDelete,p.baseURL+"/self-service/logout/api",nil);if err!=nil{return err}
	req.Header.Set("X-Session-Token",token)
	resp,err:=p.client.Do(req);if err!=nil{return err}
	defer resp.Body.Close()
	if resp.StatusCode==http.StatusUnauthorized{return nil}
	if resp.StatusCode>=300{return fmt.Errorf("Kratos logout failed: %s",resp.Status)}
	return nil
}

func(p *Provider)createFlow(ctx context.Context,path string)(struct{Action string},error){
	var raw struct{UI struct{Action string `json:"action"`}`json:"ui"`}
	req,err:=http.NewRequestWithContext(ctx,http.MethodGet,p.baseURL+path,nil);if err!=nil{return struct{Action string}{},err}
	req.Header.Set("Accept","application/json");resp,err:=p.client.Do(req);if err!=nil{return struct{Action string}{},err};defer resp.Body.Close()
	if resp.StatusCode>=300{return struct{Action string}{},fmt.Errorf("Kratos flow failed: %s",resp.Status)}
	if err:=json.NewDecoder(resp.Body).Decode(&raw);err!=nil{return struct{Action string}{},err}
	if raw.UI.Action==""{return struct{Action string}{},errors.New("Kratos flow has no action")}
	return struct{Action string}{Action:raw.UI.Action},nil
}
func(p *Provider)submit(ctx context.Context,action string,body any,out any)error{
	payload,err:=json.Marshal(body);if err!=nil{return err}
	parsed,err:=url.Parse(action);if err!=nil{return err}
	if !parsed.IsAbs(){action=p.baseURL+action}
	req,err:=http.NewRequestWithContext(ctx,http.MethodPost,action,bytes.NewReader(payload));if err!=nil{return err}
	req.Header.Set("Content-Type","application/json");req.Header.Set("Accept","application/json")
	resp,err:=p.client.Do(req);if err!=nil{return err};defer resp.Body.Close()
	if resp.StatusCode>=300{var problem any;_ = json.NewDecoder(resp.Body).Decode(&problem);return fmt.Errorf("Kratos request failed: %s: %v",resp.Status,problem)}
	if out!=nil{return json.NewDecoder(resp.Body).Decode(out)};return nil
}
func max(a,b int64)int64{if a>b{return a};return b}
