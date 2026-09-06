package chat

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

// POST /chats/direct (protected route)
func (h *Handler) CreateDirect(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())

	var req CreateDirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.svc.CreateOrGetDirectChat(r.Context(), callerID, req.OtherUserID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// GET chats (protected route)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())

	chats, err := h.svc.ListChats(r.Context(), callerID)
	if err != nil {
		log.Printf("failed to list chats: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chats)
}

func writeChatError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCannotChatWithSelf):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		log.Printf("unhandled chat error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
