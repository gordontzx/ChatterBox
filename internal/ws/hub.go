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

func (h* Hub) Register(c *Client) {
	h.mut.Lock()
	defer h.mut.Unlock()

	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*Client]struct{})
	}
	h.clients[c.userID][c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mut.Lock()
	defer h.mut.Unlock()

	if conns, ok := h.clients[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.userID)
		}
	}
	close(c.send)
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
