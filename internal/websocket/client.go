package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Subscription represents a client subscription to a message type
type Subscription struct {
	ID      string                 // UUID for this subscription
	Type    string                 // Message type subscribed to
	Options map[string]interface{} // Subscription options
}

// Client represents a single WebSocket connection
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan Message
	subscriptions map[string]Subscription // Map of subscribeID to Subscription
	mu            sync.RWMutex           // Protects subscriptions
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan Message, 256),
		subscriptions: make(map[string]Subscription),
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.handleMessage(msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
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

// handleMessage processes incoming messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(msg)
	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(msg)
	default:
		c.sendError("Unknown message type")
	}
}

// handleSubscribe processes subscription requests
func (c *Client) handleSubscribe(msg Message) {
	var subReq SubscribeRequest
	if err := convertPayload(msg.Payload, &subReq); err != nil {
		c.sendError("Invalid subscribe request format")
		return
	}

	subscribeID := uuid.New().String()
	subscription := Subscription{
		ID:      subscribeID,
		Type:    subReq.Type,
		Options: subReq.Options,
	}

	c.mu.Lock()
	c.subscriptions[subscribeID] = subscription
	c.mu.Unlock()

	response := Message{
		Type: MessageTypeSubscribeResponse,
		Payload: SubscribeResponse{
			SubscribeID: subscribeID,
			Type:        subReq.Type,
			Status:      StatusSuccess,
		},
	}

	c.send <- response
}

// handleUnsubscribe processes unsubscribe requests
func (c *Client) handleUnsubscribe(msg Message) {
	var unsubReq UnsubscribeRequest
	if err := convertPayload(msg.Payload, &unsubReq); err != nil {
		c.sendError("Invalid unsubscribe request format")
		return
	}

	c.mu.Lock()
	delete(c.subscriptions, unsubReq.SubscribeID)
	c.mu.Unlock()

	response := Message{
		Type: MessageTypeUnsubscribeResponse,
		Payload: UnsubscribeResponse{
			SubscribeID: unsubReq.SubscribeID,
			Status:      StatusSuccess,
		},
	}

	c.send <- response
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type: MessageTypeError,
		Payload: map[string]string{
			"error": errMsg,
		},
	}
	c.send <- msg
}

// convertPayload converts a payload interface to a specific type
func convertPayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
