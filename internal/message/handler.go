package message

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"ChatterBox/internal/middleware"
)

// satisfied by ws Hub
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

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	chatID := r.PathValue("chatID")

	var req MarkAsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	otherIDs, err := h.svc.MarkAsRead(r.Context(), callerID, chatID, req.UpToSeq)
	if err != nil {
		writeMessageError(w, err)
		return
	}

	// Broadcast to other users over ws
	event := ReadEvent{Type: "read", ChatID: chatID, UserID: callerID, UpToSeq: req.UpToSeq}
	if payload, err := json.Marshal(event); err == nil {
		for _, memberIDs := range otherIDs {
			h.hub.SendToUser(memberIDs, payload)
		}
	} else {
		log.Printf("failed to marshall read event: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetReadReceipts(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFromContext(r.Context())
	chatID := r.PathValue("chatID")

	receipts, err := h.svc.GetReadReceipts(r.Context(), callerID, chatID)
	if err != nil {
		writeMessageError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(receipts)
}

func writeMessageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotChatMember):
		http.Error(w, err.Error(), http.StatusForbidden) // 403
	case errors.Is(err, ErrEmptyContent),
		errors.Is(err, ErrContentTooLong),
		errors.Is(err, ErrInvalidSeq):
		http.Error(w, err.Error(), http.StatusBadRequest) // 400
	default:
		log.Printf("unhandled message error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
