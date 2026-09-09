package message

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"ChatterBox/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /chats/{chatID}/messages?since=<seq> (protected route)
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	chatID := r.PathValue("chatID")

	since, ok := parseOptionalInt64(r.URL.Query().Get("since"), 0)
	if !ok {
		http.Error(w, "invalid 'since' parameter", http.StatusBadRequest)
		return
	}

	limit, ok := parseOptionalInt64(r.URL.Query().Get("limit"), 0)
	if !ok {
		http.Error(w, "invalid 'limit' parameter", http.StatusBadRequest)
		return
	}

	message, err := h.svc.GetMessagesSince(r.Context(), callerID, chatID, since, int(limit))
	if err != nil {
		writeMessageError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}

func parseOptionalInt64(raw string, fallback int64) (int64, bool) {
	if raw == "" {
		return fallback, true
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed < 0 {
		return 0, false
	}
	return parsed, true
}

func writeMessageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotChatMember):
		http.Error(w, err.Error(), http.StatusForbidden) // 403
	default:
		log.Printf("unhandled message error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
