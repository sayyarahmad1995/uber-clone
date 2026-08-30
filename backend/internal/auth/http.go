package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Handler struct{ service Service }

func NewHandler(service Service) Handler { return Handler{service: service} }

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := decode(r, &req); err != nil || strings.TrimSpace(req.Identifier) == "" || req.Password == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	challenge, err := h.service.Register(r.Context(), Credentials{Identifier: strings.TrimSpace(req.Identifier), Password: req.Password})
	if err != nil {
		authFailure(w, err)
		return
	}
	write(w, http.StatusCreated, map[string]string{"verification_id": challenge.ChallengeID})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := decode(r, &req); err != nil || strings.TrimSpace(req.Identifier) == "" || req.Password == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	session, err := h.service.Login(r.Context(), Credentials{Identifier: strings.TrimSpace(req.Identifier), Password: req.Password})
	if err != nil {
		authFailure(w, err)
		return
	}
	write(w, http.StatusOK, session)
}

func (h Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct{ Email string `json:"email"` }
	if err := decode(r, &req); err != nil || strings.TrimSpace(req.Email) == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	challenge, err := h.service.StartVerification(r.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		authFailure(w, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"verification_id": challenge.ChallengeID})
}

func (h Handler) CompleteVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VerificationID string `json:"verification_id"`
		Code           string `json:"code"`
	}
	if err := decode(r, &req); err != nil || strings.TrimSpace(req.VerificationID) == "" || strings.TrimSpace(req.Code) == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.service.CompleteVerification(r.Context(), strings.TrimSpace(req.VerificationID), strings.TrimSpace(req.Code)); err != nil {
		authFailure(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) ExtendSession(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	session, err := h.service.ExtendSession(r.Context(), token)
	if err != nil {
		authFailure(w, err)
		return
	}
	write(w, http.StatusOK, session)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		failure(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.service.Logout(r.Context(), token); err != nil && !errors.Is(err, ErrInvalidCredentials) {
		authFailure(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	d := json.NewDecoder(io.LimitReader(r.Body, 16<<10))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func failure(w http.ResponseWriter, status int, message string) {
	write(w, status, map[string]string{"error": message})
}

func authFailure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		failure(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ErrIdentifierConflict):
		failure(w, http.StatusConflict, "identifier already exists")
	case errors.Is(err, ErrPasswordRejected):
		failure(w, http.StatusBadRequest, "password does not meet requirements")
	case errors.Is(err, ErrRegistrationInvalid):
		failure(w, http.StatusBadRequest, "registration request is invalid")
	case errors.Is(err, ErrVerificationInvalid):
		failure(w, http.StatusBadRequest, "verification is invalid or expired")
	case errors.Is(err, ErrUnavailable):
		failure(w, http.StatusServiceUnavailable, "authentication service unavailable")
	default:
		failure(w, http.StatusInternalServerError, "unable to process authentication request")
	}
}
