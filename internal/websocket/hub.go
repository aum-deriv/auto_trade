package websocket

import "sync"

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for protecting the clients map
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this message type
				client.mu.RLock()
				for _, sub := range client.subscriptions {
					if sub.Type == message.Type {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
						}
						break
					}
				}
				client.mu.RUnlock()
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients that are subscribed to the message type
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}
