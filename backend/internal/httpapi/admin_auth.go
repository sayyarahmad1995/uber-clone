package httpapi

import (
	"crypto/subtle"
	"net/http"
)

func (api *API) adminReviewConfigured() bool {
	return api.adminReviewUsername != "" && api.adminReviewPassword != ""
}

func (api *API) adminAuthenticated(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		usernameMatches := subtle.ConstantTimeCompare([]byte(username), []byte(api.adminReviewUsername)) == 1
		passwordMatches := subtle.ConstantTimeCompare([]byte(password), []byte(api.adminReviewPassword)) == 1
		if !ok || !usernameMatches || !passwordMatches {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Driver onboarding review\"")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		handler(w, r)
	})
}

func (api *API) adminMutation(handler http.HandlerFunc) http.Handler {
	return api.adminAuthenticated(func(w http.ResponseWriter, r *http.Request) {
		if !adminSameOrigin(r) {
			http.Error(w, "Invalid request origin", http.StatusForbidden)
			return
		}
		handler(w, r)
	})
}

func adminReviewer(r *http.Request) string {
	username, _, _ := r.BasicAuth()
	return username
}
