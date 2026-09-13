package chat

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"ChatterBox/internal/middleware"
)

// ws.Hub
type Broadcaster interface {
	SendToUser(userID string, payload []byte)
}

type Handler struct {
	svc *Service
	hub Broadcaster
}

func NewHandler(svc *Service, hub Broadcaster) *Handler {
	return &Handler{svc: svc, hub: hub}
}

// POST /chats/direct (protected route)
func (h *Handler) CreateDirect(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())

	var req CreateDirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, wasCreated, err := h.svc.CreateOrGetDirectChat(r.Context(), callerID, req.OtherUserID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	if wasCreated {
		h.broadcastToOthers(c, callerID, ChatEvent{Type: "chat_created", Chat: c})
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

	h.broadcastToOthers(c, callerID, ChatEvent{Type: "chat_created", Chat: c})

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

	// fetch updated chat and broadcast member_added event
	updated, err := h.svc.GetChat(r.Context(), callerID, chatID)
	if err != nil {
		log.Printf("failed to fetch updated chat for broadcast: %v", err)
	} else {
		var addedMember *Member
		for _, m := range updated.Members {
			if m.ID == req.UserID {
				addedMember = &m
				break
			}
		}
		event := ChatEvent{Type: "member_added", Chat: updated, Member: addedMember}
		h.broadcastToOthers(updated, callerID, event)
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /chats/{chatID}/leave (protected route)
func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	chatID := r.PathValue("chatID")

	c, err := h.svc.GetChat(r.Context(), callerID, chatID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	if err := h.svc.LeaveChat(r.Context(), callerID, chatID); err != nil {
		writeChatError(w, err)
		return
	}

	var leftMember *Member
	var afterLeaveMembers []Member
	for _, m := range c.Members {
		if m.ID == callerID {
			leftMember = &m
		} else {
			afterLeaveMembers = append(afterLeaveMembers, m)
		}
	}
	c.Members = afterLeaveMembers

	event := ChatEvent{Type: "member_left", Chat: c, Member: leftMember}
	h.broadcastToOthers(c, callerID, event)

	w.WriteHeader(http.StatusNoContent)
}

// GET /chats (protected route)
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

func (h *Handler) broadcastToOthers(c *Chat, excludeUserID string, event ChatEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal chat event: %v", err)
		return
	}
	for _, m := range c.Members {
		if m.ID != excludeUserID {
			h.hub.SendToUser(m.ID, payload)
		}
	}
}
