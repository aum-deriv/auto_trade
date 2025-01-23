package handler

// MessageHandler defines the interface for handling different message types
type MessageHandler interface {
	// HandleSubscribe handles subscription requests for this message type
	HandleSubscribe(subscribeID string, options map[string]interface{}) error

	// HandleUnsubscribe handles unsubscribe requests for this message type
	HandleUnsubscribe(subscribeID string) error

	// Start starts the handler's processing
	Start() error

	// Stop stops the handler's processing
	Stop() error
}
