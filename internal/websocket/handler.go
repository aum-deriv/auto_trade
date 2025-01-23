package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler represents the WebSocket handler
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeHTTP handles WebSocket requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	client := NewClient(h.hub, conn)
	client.hub.register <- client

	// Start the client's read and write pumps in separate goroutines
	go client.writePump()
	go client.readPump()
}

// HandleWebSocket returns an http.HandlerFunc for the WebSocket endpoint
func HandleWebSocket(hub *Hub) http.HandlerFunc {
	handler := NewHandler(hub)
	return handler.ServeHTTP
}
