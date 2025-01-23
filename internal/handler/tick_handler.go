package handler

import (
	"errors"
	"sync"
	"time"

	"github.com/aumbhatt/auto_trade/internal/source"
	"github.com/aumbhatt/auto_trade/internal/websocket"
)

/*
TickHandler Flow:

1. Initialization:
   TickHandler
   └── source: TickSource        // Provides tick data
   └── subs: map[string]struct{} // Stores active subscriptions
   └── hub: *websocket.Hub       // For broadcasting messages
   └── done: chan struct{}       // For graceful shutdown
   └── running: bool             // Handler state
   └── tickDelay: time.Duration  // Interval between ticks

2. Subscription Flow:
   Client → WebSocket → Registry → TickHandler
   a. Client sends subscribe request for "ticks"
   b. Registry routes to TickHandler.HandleSubscribe
   c. TickHandler adds subscription ID to subs map
   d. Client receives subscription confirmation

3. Data Flow:
   TickSource → TickHandler → Hub → Subscribers
   a. Ticker triggers every tickDelay
   b. TickHandler calls source.GetTick()
   c. For each subscribeID in subs map:
      - Creates Message with tick data
      - Adds subscribeID to Message
      - Broadcasts via Hub

4. Unsubscribe Flow:
   Client → WebSocket → Registry → TickHandler
   a. Client sends unsubscribe request
   b. Registry routes to TickHandler.HandleUnsubscribe
   c. TickHandler removes subscription ID from subs map
   d. Client stops receiving tick messages

5. Shutdown Flow:
   a. Stop() is called
   b. done channel is closed
   c. Ticker goroutine exits
   d. Resources are cleaned up

Example Message Flow:
1. Subscribe:
   → Client: {"type": "subscribe", "payload": {"type": "ticks"}}
   ← Server: {"type": "subscribe_response", "subscribe_id": "uuid1", "status": "success"}

2. Tick Data:
   ← Server: {
        "type": "ticks",
        "subscribe_id": "uuid1",
        "payload": {
            "symbol": "AAPL",
            "price": 150.25,
            "volume": 1000,
            "timestamp": "2025-01-23T11:34:23Z"
        }
     }

3. Unsubscribe:
   → Client: {"type": "unsubscribe", "payload": {"subscribe_id": "uuid1"}}
   ← Server: {"type": "unsubscribe_response", "subscribe_id": "uuid1", "status": "success"}
*/

// TickHandler handles tick message subscriptions and broadcasting
type TickHandler struct {
	hub        *websocket.Hub
	source     source.TickSource
	subs       map[string]struct{} // Map of subscribeID to empty struct (set implementation)
	mutex      sync.RWMutex
	done       chan struct{}
	running    bool
	tickDelay  time.Duration // Delay between ticks
}

// NewTickHandler creates a new TickHandler instance
func NewTickHandler(hub *websocket.Hub, source source.TickSource) *TickHandler {
	return &TickHandler{
		hub:       hub,
		source:    source,
		subs:      make(map[string]struct{}),
		tickDelay: time.Second, // Default to 1 second between ticks
	}
}

// HandleSubscribe adds a new subscription
func (h *TickHandler) HandleSubscribe(subscribeID string, options map[string]interface{}) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.subs[subscribeID] = struct{}{}
	return nil
}

// HandleUnsubscribe removes a subscription
func (h *TickHandler) HandleUnsubscribe(subscribeID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.subs, subscribeID)
	return nil
}

// Start begins the tick processing
func (h *TickHandler) Start() error {
	if h.running {
		return errors.New("handler already running")
	}

	h.done = make(chan struct{})
	h.running = true

	go func() {
		ticker := time.NewTicker(h.tickDelay)
		defer ticker.Stop()

		for {
			select {
			case <-h.done:
				return
			case <-ticker.C:
				h.processTick()
			}
		}
	}()

	return nil
}

// Stop halts the tick processing
func (h *TickHandler) Stop() error {
	if !h.running {
		return nil
	}

	close(h.done)
	h.running = false
	return nil
}

// processTick gets a new tick and broadcasts it to subscribers
func (h *TickHandler) processTick() {
	tick, err := h.source.GetTick()
	if err != nil {
		// Log error or handle it appropriately
		return
	}

	// Only broadcast if there are subscribers
	h.mutex.RLock()
	if len(h.subs) > 0 {
		// Send a separate message for each subscription
		for subID := range h.subs {
			msg := websocket.Message{
				Type:        "ticks",
				SubscribeID: subID,
				Payload:     tick,
			}
			h.hub.Broadcast(msg)
		}
	}
	h.mutex.RUnlock()
}
