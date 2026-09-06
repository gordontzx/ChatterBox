package user

import (
	"encoding/json"
	"net/http"
	"strings"

	"ChatterBox/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /users/search?email=<search_email>
func (h *Handler) SearchByEmail(w http.ResponseWriter, r *http.Request) {
	_ = middleware.UserIDFromContext(r.Context())

	rawEmail := r.URL.Query().Get("email")
	if rawEmail == "" {
		http.Error(w, "missing 'email' query parameter", http.StatusBadRequest)
		return
	}

	email, err := normalizeEmail(r.URL.Query().Get("email"))
	if err != nil {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	
	userSummary, err := h.svc.FindByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
	if userSummary == nil {
		http.Error(w, "user not found", http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(userSummary)
}

func normalizeEmail(email string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(email))
	if trimmed == "" || !strings.Contains(trimmed, "@") || strings.ContainsAny(trimmed, " \t\n") {
		return "", ErrInvalidEmail
	}
	return trimmed, nil
}
