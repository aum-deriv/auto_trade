package handler

import (
	"fmt"
	"sync"

	"github.com/aumbhatt/auto_trade/internal/websocket"
)

var _ websocket.MessageTypeRegistry = (*Registry)(nil) // Ensure Registry implements MessageTypeRegistry

/*
Registry Memory Structure:

Registry
└── handlers (map[string]MessageHandler)
    │
    ├── "ticks" → TickHandler
    │   └── subscribers: map[string]struct{}
    │       ├── "uuid1": {}  // Subscriber 1
    │       ├── "uuid2": {}  // Subscriber 2
    │       └── "uuid3": {}  // Subscriber 3
    │
    ├── "orderbook" → OrderbookHandler
    │   └── subscribers: map[string]struct{}
    │       ├── "uuid4": {}  // Subscriber 1
    │       └── "uuid5": {}  // Subscriber 2
    │
    └── "trades" → TradesHandler
        └── subscribers: map[string]struct{}
            └── "uuid6": {}  // Subscriber 1

Example subscription flow:
1. Client subscribes to "ticks"
2. Registry.HandleSubscribe("ticks", "new-uuid", options)
3. Registry looks up "ticks" handler
4. TickHandler.HandleSubscribe adds "new-uuid" to its subscribers
5. When new tick arrives, TickHandler broadcasts to all its subscribers

This structure allows:
- Independent handling of different message types
- Efficient message routing
- Clean separation of concerns
- Easy addition of new message types
*/

// Registry manages all message type handlers
type Registry struct {
	handlers map[string]MessageHandler
	mutex    sync.RWMutex
}

// NewRegistry creates a new Registry instance
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]MessageHandler),
	}
}

// Register adds a new message handler for a message type
func (r *Registry) Register(msgType string, handler MessageHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.handlers[msgType]; exists {
		return fmt.Errorf("handler for message type '%s' already registered", msgType)
	}

	r.handlers[msgType] = handler
	return nil
}

// HandleSubscribe routes subscription requests to appropriate handler
func (r *Registry) HandleSubscribe(msgType string, subscribeID string, options map[string]interface{}) error {
	r.mutex.RLock()
	handler, exists := r.handlers[msgType]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for message type '%s'", msgType)
	}

	return handler.HandleSubscribe(subscribeID, options)
}

// HandleUnsubscribe routes unsubscribe requests to appropriate handler
func (r *Registry) HandleUnsubscribe(msgType string, subscribeID string) error {
	r.mutex.RLock()
	handler, exists := r.handlers[msgType]
	r.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for message type '%s'", msgType)
	}

	return handler.HandleUnsubscribe(subscribeID)
}

// StartAll starts all registered handlers
func (r *Registry) StartAll() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for msgType, handler := range r.handlers {
		if err := handler.Start(); err != nil {
			return fmt.Errorf("failed to start handler for '%s': %w", msgType, err)
		}
	}
	return nil
}

// StopAll stops all registered handlers
func (r *Registry) StopAll() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for msgType, handler := range r.handlers {
		if err := handler.Stop(); err != nil {
			return fmt.Errorf("failed to stop handler for '%s': %w", msgType, err)
		}
	}
	return nil
}
