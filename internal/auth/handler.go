package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Register(r.Context(), req); err != nil {
		writeAuthError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tokens, err := h.svc.Login(r.Context(), req)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		log.Printf("failed to encode login response; %v", err)
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailAlreadyRegistered):
		http.Error(w, err.Error(), http.StatusConflict) // 409
	case errors.Is(err, ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized) // 401
	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong):
		http.Error(w, err.Error(), http.StatusBadRequest) // 400
	case errors.Is(err, ErrInvalidEmail):
		http.Error(w, err.Error(), http.StatusBadRequest) // 400
	default:
		log.Printf("unhandled auth error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
