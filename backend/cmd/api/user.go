package main

import (
	"net/http"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

func (app application) me(w http.ResponseWriter, r *http.Request) {
	u, ok := app.currentUser(w, r)
	if !ok {
		return
	}
	writeUser(w, http.StatusOK, u)
}

func (app application) enableDriverCapability(w http.ResponseWriter, r *http.Request) {
	p, ok := identity.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	u, err := app.users.EnableCapability(r.Context(), user.ExternalIdentity{Issuer: p.Issuer, Subject: p.Subject}, user.CapabilityDriver)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to enable driver capability"})
		return
	}
	writeUser(w, http.StatusOK, u)
}

func (app application) currentUser(w http.ResponseWriter, r *http.Request) (user.User, bool) {
	p, ok := identity.PrincipalFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return user.User{}, false
	}
	u, err := app.users.GetOrCreate(r.Context(), user.ExternalIdentity{Issuer: p.Issuer, Subject: p.Subject})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load user"})
		return user.User{}, false
	}
	return u, true
}

func writeUser(w http.ResponseWriter, status int, u user.User) {
	writeJSON(w, status, map[string]any{"id": u.ID, "capabilities": u.Capabilities})
}
