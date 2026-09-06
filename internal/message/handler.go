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

	sinceStr := r.URL.Query().Get("since")
	var since int64
	if sinceStr != "" {
		parsed, err := strconv.ParseInt(sinceStr, 10, 64)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid 'since' parameter", http.StatusBadRequest)
			return
		}
		since = parsed
	}

	message, err := h.svc.GetMessagesSince(r.Context(), callerID, chatID, since)
	if err != nil {
		writeMessageError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
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
