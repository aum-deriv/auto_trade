package websocket

// MessageTypeRegistry defines the interface for managing message type handlers
type MessageTypeRegistry interface {
	// HandleSubscribe routes subscription requests to appropriate handler
	HandleSubscribe(msgType string, subscribeID string, options map[string]interface{}) error

	// HandleUnsubscribe routes unsubscribe requests to appropriate handler
	HandleUnsubscribe(msgType string, subscribeID string) error

	// StartAll starts all registered handlers
	StartAll() error

	// StopAll stops all registered handlers
	StopAll() error
}
