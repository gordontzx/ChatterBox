package presence

import (
	"encoding/json"
	"net/http"
)

// ws.hub
type OnlineChecker interface {
	IsOnline(userID string) bool
}

type Handler struct {
	svc		*Service
	checker OnlineChecker
}

func NewHandler(svc *Service, checker OnlineChecker) *Handler {
	return &Handler{svc: svc, checker: checker}
}

// GET /users/{userID}/presence (protected route)
func (h *Handler) GetUserPresence(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	online := h.checker.IsOnline(userID)

	info, err := h.svc.GetPresence(r.Context(), userID, online)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}
