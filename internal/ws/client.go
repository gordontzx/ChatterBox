package ws

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
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
}

func NewClient(userID string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		userID: userID,
		conn:	conn,
		send:	make(chan []byte, sendBufferSize),
		hub:	hub,
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close for user %s: %v", c.userID, err)
			}
			break
		}

		// temp placeholder
		log.Printf("received from user %s: %v", c.userID, message)
	}
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
