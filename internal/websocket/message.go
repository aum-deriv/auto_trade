package websocket

// Message represents a WebSocket message
type Message struct {
	Type        string      `json:"type"`
	SubscribeID string      `json:"subscribe_id,omitempty"`
	Payload     interface{} `json:"payload"`
}

// SubscribeRequest represents a subscription request from client
type SubscribeRequest struct {
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// SubscribeResponse represents a subscription response to client
type SubscribeResponse struct {
	SubscribeID string `json:"subscribe_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

// UnsubscribeRequest represents an unsubscribe request from client
type UnsubscribeRequest struct {
	SubscribeID string `json:"subscribe_id"`
}

// UnsubscribeResponse represents an unsubscribe response to client
type UnsubscribeResponse struct {
	SubscribeID string `json:"subscribe_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

// Message types
const (
	MessageTypeSubscribe          = "subscribe"
	MessageTypeSubscribeResponse  = "subscribe_response"
	MessageTypeUnsubscribe       = "unsubscribe"
	MessageTypeUnsubscribeResponse = "unsubscribe_response"
	MessageTypeError             = "error"
)

// Status types
const (
	StatusSuccess = "success"
	StatusError   = "error"
)
