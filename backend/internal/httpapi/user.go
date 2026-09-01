package httpapi

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

func (api *API) me(w http.ResponseWriter, r *http.Request) {
	u, ok := api.currentUser(w, r)
	if !ok {
		return
	}
	writeUser(w, http.StatusOK, u)
}

func (api *API) enableDriverCapability(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	u, err := api.users.EnableCapability(r.Context(), user.ExternalIdentity{Issuer: p.Issuer, Subject: p.Subject}, user.CapabilityDriver)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to enable driver capability"})
		return
	}
	writeUser(w, http.StatusOK, u)
}

func (api *API) currentUser(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	p, ok := identity.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return user.User{}, false
	}
	u, err := api.users.GetOrCreate(r.Context(), user.ExternalIdentity{Issuer: p.Issuer, Subject: p.Subject})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load user"})
		return user.User{}, false
	}
	return u, true
}

func (api *API) requireRiderCapability(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	u, ok := api.currentUser(w, r)
	if !ok {
		return user.User{}, false
	}
	for _, capability := range u.Capabilities {
		if capability == user.CapabilityRider {
			return u, true
		}
	}
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "rider capability required"})
	return user.User{}, false
}

func (api *API) requireDriverCapability(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	u, ok := api.currentUser(w, r)
	if !ok {
		return user.User{}, false
	}
	for _, capability := range u.Capabilities {
		if capability == user.CapabilityDriver {
			return u, true
		}
	}
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "driver capability required"})
	return user.User{}, false
}

func writeUser(w http.ResponseWriter, status int, u user.User) {
	writeJSON(w, status, map[string]any{"id": u.ID, "capabilities": u.Capabilities})
}
