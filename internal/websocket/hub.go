package websocket

import "sync"

/*
Hub Memory Structure and Message Flow:

1. Memory Structure:
   Hub
   ├── clients: map[*Client]bool        // Active client connections
   ├── broadcast: chan Message          // Channel for broadcasting messages
   ├── register: chan *Client           // Channel for new client registration
   ├── unregister: chan *Client         // Channel for client disconnection
   ├── mu: sync.RWMutex                // Protects clients map
   └── registry: *handler.Registry      // Message type handlers
       └── handlers: map[string]MessageHandler
           ├── "ticks" → TickHandler
           ├── "orderbook" → OrderbookHandler
           └── "trades" → TradesHandler

2. Message Flow Examples:

   a. Client Registration:
      1. New WebSocket connection established
      2. Client instance created
      3. Client sent to Hub's register channel
      4. Hub adds client to clients map

   b. Message Broadcasting:
      Tick Data Example:
      1. TickHandler generates new tick
      2. Calls Hub.Broadcast with message:
         {
           "type": "ticks",
           "subscribe_id": "uuid-123",
           "payload": {
             "symbol": "AAPL",
             "price": 150.25
           }
         }
      3. Hub sends to all relevant clients

   c. Client Disconnection:
      1. Client connection closes
      2. Client sent to Hub's unregister channel
      3. Hub removes client from clients map
      4. Hub closes client's send channel

3. Concurrent Operations:
   - Multiple clients can connect/disconnect simultaneously
   - Messages can be broadcast while clients connect/disconnect
   - Thread-safe operations on clients map using mutex
   - Non-blocking message sending using select

4. Error Handling:
   - Graceful handling of client disconnections
   - Channel closing on client removal
   - Mutex protection for shared resources
   - Non-blocking message broadcasts

5. Integration with Registry:
   - Registry manages message type handlers
   - Hub provides broadcasting capability
   - Handlers use hub to send messages
   - Clean separation of concerns
*/

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

	// Registry for message type handlers
	registry MessageTypeRegistry
}

// NewHub creates a new Hub instance
func NewHub(registry MessageTypeRegistry) *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		registry:   registry,
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
				// Only send to clients subscribed to this message type
				if client.isSubscribed(message.Type, message.SubscribeID) {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}
