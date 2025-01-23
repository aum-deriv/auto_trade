package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/*
WebSocket Client Flow and Memory Structure:

1. Memory Structure:
   Client
   └── hub: *Hub                // Reference to central hub
   └── conn: *websocket.Conn    // WebSocket connection
   └── send: chan Message       // Outbound message queue

2. Connection Flow:
   Browser → WebSocket Server → Client Instance
   a. Browser connects to "/ws"
   b. Server upgrades HTTP to WebSocket
   c. New Client instance created
   d. Client registered with hub
   e. Read/Write pumps started

3. Message Flow Examples:

   a. Subscribe to Ticks:
      → Client Receives:
        {
          "type": "subscribe",
          "payload": {
            "type": "ticks",
            "options": {"symbols": ["AAPL", "GOOGL"]}
          }
        }

      Processing:
      1. handleMessage identifies subscribe request
      2. Generates new UUID for subscription
      3. Routes to registry.HandleSubscribe
      4. Registry forwards to TickHandler
      5. TickHandler starts sending data

      ← Client Sends:
        {
          "type": "subscribe_response",
          "payload": {
            "subscribe_id": "550e8400-e29b-41d4-a716-446655440000",
            "type": "ticks",
            "status": "success"
          }
        }

   b. Receive Tick Data:
      ← Client Sends:
        {
          "type": "ticks",
          "subscribe_id": "550e8400-e29b-41d4-a716-446655440000",
          "payload": {
            "symbol": "AAPL",
            "price": 150.25,
            "volume": 1000,
            "timestamp": "2025-01-23T11:36:00Z"
          }
        }

   c. Unsubscribe:
      → Client Receives:
        {
          "type": "unsubscribe",
          "payload": {
            "subscribe_id": "550e8400-e29b-41d4-a716-446655440000"
          }
        }

      Processing:
      1. handleMessage identifies unsubscribe request
      2. Routes to registry.HandleUnsubscribe
      3. Registry forwards to TickHandler
      4. TickHandler removes subscription

      ← Client Sends:
        {
          "type": "unsubscribe_response",
          "payload": {
            "subscribe_id": "550e8400-e29b-41d4-a716-446655440000",
            "status": "success"
          }
        }

4. Error Handling:
   ← Error Response Example:
     {
       "type": "error",
       "payload": {
         "error": "Invalid subscribe request format"
       }
     }
*/

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a single WebSocket connection
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan Message
	// Track subscriptions
	subscriptions    sync.Map // map[string]map[string]struct{} // msgType -> subscribeIDs
	subscriptionType sync.Map // map[string]string // subscribeID -> msgType
}

// NewClient creates a new client instance
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan Message, 256),
	}
}

// isSubscribed checks if the client is subscribed to a specific message type and subscription ID
func (c *Client) isSubscribed(msgType, subscribeID string) bool {
	if subs, ok := c.subscriptions.Load(msgType); ok {
		if subMap, ok := subs.(map[string]struct{}); ok {
			_, exists := subMap[subscribeID]
			return exists
		}
	}
	return false
}

// addSubscription adds a subscription for a message type
func (c *Client) addSubscription(msgType, subscribeID string) {
	var subMap map[string]struct{}
	if subs, ok := c.subscriptions.Load(msgType); ok {
		subMap = subs.(map[string]struct{})
	} else {
		subMap = make(map[string]struct{})
	}
	subMap[subscribeID] = struct{}{}
	c.subscriptions.Store(msgType, subMap)
}

// removeSubscription removes a subscription
func (c *Client) removeSubscription(msgType, subscribeID string) {
	if subs, ok := c.subscriptions.Load(msgType); ok {
		if subMap, ok := subs.(map[string]struct{}); ok {
			delete(subMap, subscribeID)
			if len(subMap) == 0 {
				c.subscriptions.Delete(msgType)
			} else {
				c.subscriptions.Store(msgType, subMap)
			}
		}
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
				// The hub closed the channel
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
		var subReq SubscribeRequest
		if err := convertPayload(msg.Payload, &subReq); err != nil {
			c.sendError("Invalid subscribe request format")
			return
		}

		subscribeID := uuid.New().String()
		if err := c.hub.registry.HandleSubscribe(subReq.Type, subscribeID, subReq.Options); err != nil {
			c.sendError(fmt.Sprintf("Subscription failed: %v", err))
			return
		}

		// Track the subscription locally
		c.addSubscription(subReq.Type, subscribeID)
		c.subscriptionType.Store(subscribeID, subReq.Type)

		response := Message{
			Type: MessageTypeSubscribeResponse,
			Payload: SubscribeResponse{
				SubscribeID: subscribeID,
				Type:        subReq.Type,
				Status:      StatusSuccess,
			},
		}
		c.send <- response

	case MessageTypeUnsubscribe:
		var unsubReq UnsubscribeRequest
		if err := convertPayload(msg.Payload, &unsubReq); err != nil {
			c.sendError("Invalid unsubscribe request format")
			return
		}

		// Get message type for this subscription
		msgTypeI, ok := c.subscriptionType.Load(unsubReq.SubscribeID)
		if !ok {
			c.sendError("Invalid subscription ID")
			return
		}
		msgType := msgTypeI.(string)

		if err := c.hub.registry.HandleUnsubscribe(msgType, unsubReq.SubscribeID); err != nil {
			c.sendError(fmt.Sprintf("Unsubscribe failed: %v", err))
			return
		}

		// Remove the subscription locally
		c.removeSubscription(msgType, unsubReq.SubscribeID)
		c.subscriptionType.Delete(unsubReq.SubscribeID)

		response := Message{
			Type: MessageTypeUnsubscribeResponse,
			Payload: UnsubscribeResponse{
				SubscribeID: unsubReq.SubscribeID,
				Status:      StatusSuccess,
			},
		}
		c.send <- response

	default:
		c.sendError("Unknown message type")
	}
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
