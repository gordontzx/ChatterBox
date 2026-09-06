package ws

import (
	"log"
	"net/http"

	"ChatterBox/internal/auth"
	"ChatterBox/internal/message"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub    *Hub
	secret []byte
	msgSvc *message.Service
}

func NewHandler(hub *Hub, secret []byte, msgSvc *message.Service) *Handler {
	return &Handler{hub: hub, secret: secret, msgSvc: msgSvc}
}

// GET /ws?token=<access_token>
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	claims, err := auth.VerifyAccessToken(tokenStr, h.secret)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := NewClient(claims.UserID, conn, h.hub, h.msgSvc)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
