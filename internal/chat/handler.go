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

// POST /chats/group (protected route)
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())

	var req CreateGroupChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, err := h.svc.CreateGroupChat(r.Context(), callerID, req)
	if err != nil {
		writeChatError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c)
}

// POST /chats/{chatID}/members (protected route)
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	chatID := r.PathValue("chatID")

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.AddMember(r.Context(), callerID, chatID, req.UserID); err != nil {
		writeChatError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	case errors.Is(err, ErrCannotChatWithSelf),
		errors.Is(err, ErrGroupNameRequired),
		errors.Is(err, ErrGroupTooSmall):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrChatNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrNotChatMember):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, ErrNotGroupChat), errors.Is(err, ErrAlreadyMember):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		log.Printf("unhandled chat error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
