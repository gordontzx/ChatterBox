package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"ChatterBox/internal/message"
)

const (
	writeWait 	   = 10 * time.Second
	pongWait 	   = 60 * time.Second
	pingPeriod	   = pongWait * 9 / 10
	sendBufferSize = 256
)

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
	msgSvc *message.Service
}

func NewClient(userID string, conn *websocket.Conn, hub *Hub, msgSvc *message.Service) *Client {
	return &Client{
		userID: userID,
		conn:	conn,
		send:	make(chan []byte, sendBufferSize),
		hub:	hub,
		msgSvc:	msgSvc,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close for user %s: %v", c.userID, err)
			}
			break
		}

		c.handleIncoming(raw)
	}
}

func (c *Client) handleIncoming(raw []byte) {
	var incoming message.IncomingMessage
	if err := json.Unmarshal(raw, &incoming); err != nil {
		log.Printf("malformed frame from user %s: %v", c.userID, err)
		return
	}

	switch incoming.Type {
	case "message":
		c.handleSendMessage(incoming)
	case "typing":
		c.handleTyping(incoming)
	default:
		log.Printf("unknown frame type %q from user %s", incoming.Type, c.userID)
	}
}

func (c *Client) handleTyping(incoming message.IncomingMessage) {
	otherIDs, err := c.msgSvc.SendTypingIndicator(context.Background(), c.userID, incoming.ChatID)
	if err != nil {
		log.Printf("typing indicator failed for user %s: %v", c.userID, err)
		return
	}

	event := message.TypingEvent{
		Type:   "typing",
		ChatID: incoming.ChatID,
		UserID: c.userID,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal typing event: %v", err)
		return
	}

	for _, memberID := range otherIDs {
		c.hub.SendToUser(memberID, payload)
	}
}

func (c *Client) handleSendMessage(incoming message.IncomingMessage) {
	result, err := c.msgSvc.SendMessage(context.Background(), c.userID, incoming.ChatID, incoming.Content)
	if err != nil {
		log.Printf("send message failed for user %s: %v", c.userID, err)
		return
	}

	event := message.OutgoingEvent{Type: "message", Message: result.Message}
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal outgoing event: %v", err)
		return
	}

	for _, memberID := range result.OtherMemberIDs {
		c.hub.SendToUser(memberID, payload)
	}

	// echo to sender
	c.hub.SendToUser(c.userID, payload)
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
