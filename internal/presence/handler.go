package presence

import (
	"ChatterBox/internal/middleware"
	"context"
	"encoding/json"
	"net/http"
)

// ws.hub
type OnlineChecker interface {
	IsOnline(userID string) bool
}

// chat.service
type ContactChecker interface {
	ShareAnyChat(ctx context.Context, user1, user2 string) (bool, error)
}

type Handler struct {
	svc		 *Service
	checker  OnlineChecker
	contacts ContactChecker
}

func NewHandler(svc *Service, checker OnlineChecker, contacts ContactChecker) *Handler {
	return &Handler{svc: svc, checker: checker, contacts: contacts}
}

// GET /users/{userID}/presence (protected route)
// Viewing presence of any user requires sharing at least one chat with them
func (h *Handler) GetUserPresence(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	targetUserID := r.PathValue("userID")

	if callerID != targetUserID {
		shared, err := h.contacts.ShareAnyChat(r.Context(), callerID, targetUserID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !shared {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	online := h.checker.IsOnline(targetUserID)
	info, err := h.svc.GetPresence(r.Context(), targetUserID, online)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}
