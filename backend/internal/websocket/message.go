package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType defines all possible WebSocket event types.
// Using constants prevents typos and makes the protocol explicit.
type EventType string

const (
	// Inbound — client sends these to server
	EventTypeSendMessage EventType = "message"
	EventTypeTypingStart EventType = "typing_start"
	EventTypeTypingStop  EventType = "typing_stop"

	// Outbound — server sends these to clients
	EventTypeNewMessage EventType = "new_message"
	EventTypeUserJoined EventType = "user_joined"
	EventTypeUserLeft   EventType = "user_left"
	EventTypeError      EventType = "error"
)

// InboundEvent is what the client sends to the server.
// We decode the Payload lazily using json.RawMessage —
// this lets us decode the outer envelope first,
// then decode the payload based on the event type.
type InboundEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// OutboundEvent is what the server sends to clients.
type OutboundEvent struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// SendMessagePayload is the payload for EventTypeSendMessage
type SendMessagePayload struct {
	Content string `json:"content"`
}

// NewMessagePayload is the payload for EventTypeNewMessage
// This is what gets broadcast to all clients in the channel
type NewMessagePayload struct {
	ID        uuid.UUID     `json:"id"`
	Content   string        `json:"content"`
	ChannelID uuid.UUID     `json:"channel_id"`
	Sender    MessageSender `json:"sender"`
	CreatedAt time.Time     `json:"created_at"`
}

// MessageSender carries sender info inside a broadcast
type MessageSender struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// ErrorPayload is sent when something goes wrong
type ErrorPayload struct {
	Message string `json:"message"`
}

// encode converts an OutboundEvent to JSON bytes for sending
func encode(event OutboundEvent) ([]byte, error) {
	return json.Marshal(event)
}
