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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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
		authError(w, http.StatusBadRequest, "invalid_request", "Invalid request.")
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

func authError(w http.ResponseWriter, status int, code, message string) {
	write(w, status, map[string]string{"error": code, "message": message})
}

func authFailure(w http.ResponseWriter, err error) {
	status, code, message := httpAuthError(err)
	var publicErr *PublicError
	if errors.As(err, &publicErr) {
		if strings.TrimSpace(publicErr.Code) != "" {
			code = publicErr.Code
		}
		if strings.TrimSpace(publicErr.Message) != "" {
			message = publicErr.Message
		}
	}
	authError(w, status, code, message)
}

func httpAuthError(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid_credentials", "Invalid credentials."
	case errors.Is(err, ErrVerificationRequired):
		return http.StatusForbidden, "verification_required", "Account verification is required."
	case errors.Is(err, ErrIdentifierConflict):
		return http.StatusConflict, "identifier_already_exists", "An account with this identifier already exists."
	case errors.Is(err, ErrPasswordRejected):
		return http.StatusBadRequest, "password_rejected", "Password does not meet requirements."
	case errors.Is(err, ErrRegistrationInvalid):
		return http.StatusBadRequest, "registration_invalid", "Registration request is invalid."
	case errors.Is(err, ErrVerificationInvalid):
		return http.StatusBadRequest, "verification_invalid", "Verification is invalid or expired."
	case errors.Is(err, ErrUnavailable):
		return http.StatusServiceUnavailable, "authentication_unavailable", "Authentication service is unavailable."
	default:
		return http.StatusInternalServerError, "authentication_failed", "Unable to process authentication request."
	}
}
