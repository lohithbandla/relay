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

	// Inbound — call signaling
	EventTypeCallJoin         EventType = "call_join"
	EventTypeCallLeave        EventType = "call_leave"
	EventTypeCallOffer        EventType = "call_offer"
	EventTypeCallAnswer       EventType = "call_answer"
	EventTypeCallICECandidate EventType = "call_ice_candidate"
	EventTypeCallMute         EventType = "call_mute"
	EventTypeCallCamera       EventType = "call_camera"

	// Outbound — server sends these to clients
	EventTypeNewMessage EventType = "new_message"
	EventTypeUserJoined EventType = "user_joined"
	EventTypeUserLeft   EventType = "user_left"
	EventTypeError      EventType = "error"

	// Outbound — call signaling
	// call_offer / call_answer / call_ice_candidate are reused outbound too —
	// same event type, routed to a single target peer instead of broadcast.
	EventTypeCallParticipants EventType = "call_participants"
	EventTypeCallUserJoined   EventType = "call_user_joined"
	EventTypeCallUserLeft     EventType = "call_user_left"
	EventTypeCallUserMuted    EventType = "call_user_muted"
	EventTypeCallUserCamera   EventType = "call_user_camera"
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

// --- Call signaling payloads ---
//
// Mesh topology: every participant holds one RTCPeerConnection per other
// participant. The backend never touches media — it only routes offer/
// answer/ICE messages between the two peers named in the payload.

// CallParticipant describes one existing participant in an active call.
type CallParticipant struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Muted     bool   `json:"muted"`
	CameraOff bool   `json:"camera_off"`
}

// CallParticipantsPayload is sent once, to a client right after it joins a
// call, listing who's already there. The joining client is responsible for
// initiating an offer to each of them — this avoids both sides racing to
// create an offer for the same pair (SDP glare).
type CallParticipantsPayload struct {
	Participants []CallParticipant `json:"participants"`
}

// CallUserPayload announces a participant joining/leaving/muting/toggling
// their camera to the rest of the call.
type CallUserPayload struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Muted     bool   `json:"muted,omitempty"`
	CameraOff bool   `json:"camera_off,omitempty"`
}

// CallSignalPayload carries an SDP offer or answer between two peers.
// Inbound: TargetUserID says who to route it to. Outbound: FromUserID says
// who it came from.
type CallSignalPayload struct {
	TargetUserID string `json:"target_user_id,omitempty"`
	FromUserID   string `json:"from_user_id,omitempty"`
	SDP          string `json:"sdp"`
}

// CallICEPayload carries one ICE candidate between two peers. Candidate is
// passed through opaque — the backend never inspects it, just forwards it.
type CallICEPayload struct {
	TargetUserID string      `json:"target_user_id,omitempty"`
	FromUserID   string      `json:"from_user_id,omitempty"`
	Candidate    interface{} `json:"candidate"`
}

// CallMutePayload is the inbound payload for EventTypeCallMute.
type CallMutePayload struct {
	Muted bool `json:"muted"`
}

// CallCameraPayload is the inbound payload for EventTypeCallCamera.
type CallCameraPayload struct {
	CameraOff bool `json:"camera_off"`
}

// encode converts an OutboundEvent to JSON bytes for sending
func encode(event OutboundEvent) ([]byte, error) {
	return json.Marshal(event)
}
