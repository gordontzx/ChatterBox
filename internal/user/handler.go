package user

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"ChatterBox/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /users/me (protected route)
func (h* Handler) Me(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())

	u, err := h.svc.FindByID(r.Context(), callerID)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

// GET /users/search?email=<search_email> (protected route)
func (h *Handler) SearchByEmail(w http.ResponseWriter, r *http.Request) {
	_ = middleware.UserIDFromContext(r.Context())

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "missing 'email' query parameter", http.StatusBadRequest)
		return
	}

	u, err := h.svc.FindByEmail(r.Context(), email)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		log.Printf("unhandled user error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
