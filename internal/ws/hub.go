package ws

import "sync"

type Hub struct {
	mut		sync.RWMutex
	clients map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

// Returns true if this is the client's first connection
func (h* Hub) Register(c *Client) bool {
	h.mut.Lock()
	defer h.mut.Unlock()

	isFirstConnection := len(h.clients[c.userID]) == 0
	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*Client]struct{})
	}
	h.clients[c.userID][c] = struct{}{}
	return isFirstConnection
}

// Returns true if the client has 0 remaining connections
func (h *Hub) Unregister(c *Client) bool {
	h.mut.Lock()

	isNowFullyOffline := false
	if conns, ok := h.clients[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.userID)
			isNowFullyOffline = true
		}
	}
	h.mut.Unlock()

	close(c.send)
	return isNowFullyOffline
}

func (h* Hub) SendToUser(userID string, message []byte) {
	h.mut.RLock()
	defer h.mut.RUnlock()

	for c := range h.clients[userID] {
		select {
		case c.send <- message:
		default:
			// Let ping/pong timeout clean up
		}
	}
}

func (h *Hub) IsOnline(userID string) bool {
	h.mut.RLock()
	defer h.mut.RUnlock()
	return len(h.clients[userID]) > 0
}
