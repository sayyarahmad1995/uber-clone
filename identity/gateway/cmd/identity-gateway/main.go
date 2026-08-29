package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type config struct {
	Port string
	HydraAdminURL string
	KratosPublicURL string
}

func main() {
	cfg := config{
		Port: getenv("PORT", "8081"),
		HydraAdminURL: getenv("HYDRA_ADMIN_URL", "http://hydra:4445"),
		KratosPublicURL: getenv("KRATOS_PUBLIC_URL", "http://kratos:4433"),
	}
	app := gateway{cfg: cfg, client: http.DefaultClient}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status":"ok"}) })
	mux.HandleFunc("GET /oauth2/login", app.login)
	mux.HandleFunc("GET /oauth2/consent", app.consent)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

type gateway struct { cfg config; client *http.Client }

func (g gateway) login(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("login_challenge")
	if challenge == "" { http.Error(w, "missing login_challenge", 400); return }

	if r.URL.Query().Get("identity") == "" {
		// Explicit development boundary: the gateway delegates interactive login to Kratos.
		// A browser UI must complete the Kratos flow and return to this endpoint.
		returnTo := "http://localhost:8081/oauth2/login?login_challenge="+url.QueryEscape(challenge)
		target := g.cfg.KratosPublicURL+"/self-service/login/browser?return_to="+url.QueryEscape(returnTo)
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	http.Error(w, "direct identity injection is disabled; Kratos session integration UI is required", http.StatusNotImplemented)
}

func (g gateway) consent(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("consent_challenge")
	if challenge == "" { http.Error(w, "missing consent_challenge", 400); return }

	req, err := g.client.Get(g.cfg.HydraAdminURL+"/admin/oauth2/auth/requests/consent?consent_challenge="+url.QueryEscape(challenge))
	if err != nil { http.Error(w, "hydra unavailable", 502); return }
	defer req.Body.Close()
	if req.StatusCode >= 300 { http.Error(w, "unable to load consent request", 502); return }

	var body struct{ RequestedScope []string `json:"requested_scope"` }
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil { http.Error(w, "invalid hydra response", 502); return }

	payload, _ := json.Marshal(map[string]any{"grant_scope": body.RequestedScope, "remember": true, "remember_for": 3600})
	accept, err := http.NewRequest(http.MethodPut, g.cfg.HydraAdminURL+"/admin/oauth2/auth/requests/consent/accept?consent_challenge="+url.QueryEscape(challenge), strings.NewReader(string(payload)))
	if err != nil { http.Error(w, "request creation failed", 500); return }
	accept.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(accept)
	if err != nil { http.Error(w, "hydra unavailable", 502); return }
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { http.Error(w, "consent acceptance failed", 502); return }

	var accepted struct{ RedirectTo string `json:"redirect_to"` }
	_ = json.NewDecoder(resp.Body).Decode(&accepted)
	http.Redirect(w, r, accepted.RedirectTo, http.StatusFound)
}

func getenv(k,d string) string { if v:=os.Getenv(k); v!="" { return v }; return d }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); _=json.NewEncoder(w).Encode(value) }
var _ = io.EOF
