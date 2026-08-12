package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lohithbandla/relay/internal/messages"
	redispkg "github.com/lohithbandla/relay/internal/redis"
)

// BroadcastMessage is what gets sent through the hub's broadcast channel.
type BroadcastMessage struct {
	ChannelID uuid.UUID
	SenderID  string
	Username  string
	Data      []byte
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client

	messageService *messages.Service

	// subscriptions tracks which channels have active Redis subscriptions.
	// Protected by subMu since StartChannelSubscription runs in goroutines.
	subscriptions map[uuid.UUID]bool
	subMu         sync.Mutex

	// callParticipants tracks users currently in a voice/video call.
	// This is separate from clients because a user can be connected
	// to a text channel without being in a call.
	//
	// Only accessed from the Run() goroutine.
	callParticipants map[uuid.UUID]map[string]*Client

	// Call state is kept inside the Hub instead of Client.
	// This avoids coupling the WebSocket Client to call-specific state.
	callMuted     map[uuid.UUID]map[string]bool
	callCameraOff map[uuid.UUID]map[string]bool
}

// NewHub creates a new Hub.
func NewHub(messageService *messages.Service) *Hub {
	return &Hub{
		clients:          make(map[uuid.UUID]map[*Client]bool),
		broadcast:        make(chan *BroadcastMessage, 256),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		messageService:   messageService,
		subscriptions:    make(map[uuid.UUID]bool),
		callParticipants: make(map[uuid.UUID]map[string]*Client),
		callMuted:        make(map[uuid.UUID]map[string]bool),
		callCameraOff:    make(map[uuid.UUID]map[string]bool),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	log.Println("[hub] Starting WebSocket hub")

	for {
		select {
		case client := <-h.register:
			if h.clients[client.ChannelID] == nil {
				h.clients[client.ChannelID] = make(map[*Client]bool)
			}

			h.clients[client.ChannelID][client] = true

			log.Printf(
				"[hub] Client %s registered to channel %s",
				client.ID,
				client.ChannelID,
			)

			// Start Redis subscription when first client joins channel.
			h.StartChannelSubscription(client.ChannelID)

			h.broadcastEvent(client.ChannelID, OutboundEvent{
				Type: EventTypeUserJoined,
				Payload: map[string]string{
					"user_id":  client.UserID,
					"username": client.Username,
				},
			}, client)

		case client := <-h.unregister:
			if clients, ok := h.clients[client.ChannelID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)

					log.Printf(
						"[hub] Client %s unregistered from channel %s",
						client.ID,
						client.ChannelID,
					)

					if len(clients) == 0 {
						delete(h.clients, client.ChannelID)

						// Stop Redis subscription when last client leaves.
						h.StopChannelSubscription(client.ChannelID)
					}

					h.broadcastEvent(client.ChannelID, OutboundEvent{
						Type: EventTypeUserLeft,
						Payload: map[string]string{
							"user_id":  client.UserID,
							"username": client.Username,
						},
					}, nil)
				}
			}

			// A dropped WebSocket connection also removes the user
			// from any active call.
			h.removeFromCall(client.ChannelID, client.UserID, true)

		case message := <-h.broadcast:
			h.handleInboundMessage(message)
		}
	}
}

// handleInboundMessage parses and routes inbound WebSocket events.
func (h *Hub) handleInboundMessage(msg *BroadcastMessage) {
	var event InboundEvent

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("[hub] Failed to parse event: %v", err)
		return
	}

	switch event.Type {
	case EventTypeSendMessage:
		h.handleSendMessage(msg, event)

	case EventTypeTypingStart, EventTypeTypingStop:
		h.handleTyping(msg, event)

	case EventTypeCallJoin:
		h.handleCallJoin(msg)

	case EventTypeCallLeave:
		h.removeFromCall(msg.ChannelID, msg.SenderID, true)

	case EventTypeCallOffer,
		EventTypeCallAnswer,
		EventTypeCallICECandidate:
		h.handleCallSignal(msg, event)

	case EventTypeCallMute:
		h.handleCallMute(msg, event)

	case EventTypeCallCamera:
		h.handleCallCamera(msg, event)

	default:
		log.Printf("[hub] Unknown event type: %s", event.Type)
	}
}

// handleSendMessage persists and publishes the message via Redis.
func (h *Hub) handleSendMessage(
	msg *BroadcastMessage,
	event InboundEvent,
) {
	var payload SendMessagePayload

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("[hub] Failed to parse message payload: %v", err)
		return
	}

	if payload.Content == "" {
		return
	}

	senderID, err := uuid.Parse(msg.SenderID)
	if err != nil {
		log.Printf("[hub] Invalid sender ID: %v", err)
		return
	}

	// Save to PostgreSQL.
	savedMessage, err := h.messageService.SendMessage(
		messages.SendMessageRequest{
			Content: payload.Content,
		},
		msg.ChannelID,
		senderID,
	)

	if err != nil {
		log.Printf("[hub] Failed to save message: %v", err)
		return
	}

	// Build outbound event.
	outbound := OutboundEvent{
		Type: EventTypeNewMessage,
		Payload: NewMessagePayload{
			ID:        savedMessage.ID,
			Content:   savedMessage.Content,
			ChannelID: savedMessage.ChannelID,
			Sender: MessageSender{
				ID:       msg.SenderID,
				Username: msg.Username,
			},
			CreatedAt: time.Now().UTC(),
		},
	}

	// Publish to Redis.
	//
	// IMPORTANT:
	// Do NOT json.Marshal(outbound) here.
	// Publish() already marshals the supplied value.
	if err := redispkg.Publish(
		context.Background(),
		redispkg.ChannelKey(msg.ChannelID.String()),
		outbound,
	); err != nil {
		log.Printf("[hub] Failed to publish to Redis: %v", err)

		// Redis failed, so at least broadcast locally.
		h.broadcastEvent(msg.ChannelID, outbound, nil)
	}
}

// handleTyping broadcasts typing indicators locally only.
//
// Typing indicators are ephemeral, so there is no need for Redis Pub/Sub.
func (h *Hub) handleTyping(
	msg *BroadcastMessage,
	event InboundEvent,
) {
	outbound := OutboundEvent{
		Type: event.Type,
		Payload: map[string]string{
			"user_id":  msg.SenderID,
			"username": msg.Username,
		},
	}

	h.broadcastEventExcludingSender(
		msg.ChannelID,
		outbound,
		msg.SenderID,
	)
}

// broadcastEvent sends an event to all clients in a channel,
// optionally excluding one client.
func (h *Hub) broadcastEvent(
	channelID uuid.UUID,
	event OutboundEvent,
	exclude *Client,
) {
	data, err := encode(event)
	if err != nil {
		return
	}

	clients := h.clients[channelID]

	for client := range clients {
		if exclude != nil && client == exclude {
			continue
		}

		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}

// broadcastEventExcludingSender sends to all clients except the sender.
func (h *Hub) broadcastEventExcludingSender(
	channelID uuid.UUID,
	event OutboundEvent,
	senderID string,
) {
	data, err := encode(event)
	if err != nil {
		return
	}

	clients := h.clients[channelID]

	for client := range clients {
		if client.UserID == senderID {
			continue
		}

		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}

// GetChannelClientCount returns active client count for a channel.
func (h *Hub) GetChannelClientCount(channelID uuid.UUID) int {
	if clients, ok := h.clients[channelID]; ok {
		return len(clients)
	}

	return 0
}

// -----------------------------------------------------------------------------
// Call signaling
// -----------------------------------------------------------------------------

// The backend never handles actual audio/video.
//
// WebRTC media flows directly between browsers.
//
// The backend only handles:
//   - call join
//   - call leave
//   - SDP offer
//   - SDP answer
//   - ICE candidates
//   - mute state
//   - camera state
//
// IMPORTANT:
// Call state is currently local to this Hub instance.
// For a multi-instance backend, this would eventually need Redis routing.

// handleCallJoin adds the sender to the channel's call.
//
// The joining client receives a snapshot of existing participants.
// The joining client can then create WebRTC offers to those peers.
//
// Existing participants receive a call_user_joined event.
func (h *Hub) handleCallJoin(msg *BroadcastMessage) {
	client := h.findClient(
		msg.ChannelID,
		msg.SenderID,
	)

	if client == nil {
		return
	}

	participants := h.callParticipants[msg.ChannelID]

	if participants == nil {
		participants = make(map[string]*Client)
		h.callParticipants[msg.ChannelID] = participants
	}

	// Already in call.
	if _, already := participants[msg.SenderID]; already {
		return
	}

	// Make sure call state maps exist.
	if h.callMuted[msg.ChannelID] == nil {
		h.callMuted[msg.ChannelID] = make(map[string]bool)
	}

	if h.callCameraOff[msg.ChannelID] == nil {
		h.callCameraOff[msg.ChannelID] = make(map[string]bool)
	}

	// Build list of existing participants.
	existing := make(
		[]CallParticipant,
		0,
		len(participants),
	)

	for userID, c := range participants {
		existing = append(existing, CallParticipant{
			UserID:    c.UserID,
			Username:  c.Username,
			Muted:     h.callMuted[msg.ChannelID][userID],
			CameraOff: h.callCameraOff[msg.ChannelID][userID],
		})
	}

	// Add new participant.
	participants[msg.SenderID] = client

	h.callMuted[msg.ChannelID][msg.SenderID] = false
	h.callCameraOff[msg.ChannelID][msg.SenderID] = false

	// Send current call participants to the joining client.
	h.sendToClient(client, OutboundEvent{
		Type: EventTypeCallParticipants,
		Payload: CallParticipantsPayload{
			Participants: existing,
		},
	})

	// Tell existing participants that a new user joined.
	h.broadcastToCallExcluding(
		msg.ChannelID,
		msg.SenderID,
		OutboundEvent{
			Type: EventTypeCallUserJoined,
			Payload: CallUserPayload{
				UserID:   msg.SenderID,
				Username: msg.Username,
			},
		},
	)
}

// removeFromCall removes a user from the call.
//
// If notify is true, remaining participants receive call_user_left.
func (h *Hub) removeFromCall(
	channelID uuid.UUID,
	userID string,
	notify bool,
) {
	participants := h.callParticipants[channelID]

	if participants == nil {
		return
	}

	if _, ok := participants[userID]; !ok {
		return
	}

	delete(participants, userID)

	// Clean up call state.
	if h.callMuted[channelID] != nil {
		delete(h.callMuted[channelID], userID)
	}

	if h.callCameraOff[channelID] != nil {
		delete(h.callCameraOff[channelID], userID)
	}

	// If call is empty, clean everything up.
	if len(participants) == 0 {
		delete(h.callParticipants, channelID)
		delete(h.callMuted, channelID)
		delete(h.callCameraOff, channelID)
	}

	if notify {
		h.broadcastToCallExcluding(
			channelID,
			userID,
			OutboundEvent{
				Type: EventTypeCallUserLeft,
				Payload: CallUserPayload{
					UserID: userID,
				},
			},
		)
	}
}

// handleCallSignal forwards an offer, answer, or ICE candidate
// to exactly the target peer.
//
// These signaling messages must NOT be broadcast to the whole channel.
func (h *Hub) handleCallSignal(
	msg *BroadcastMessage,
	event InboundEvent,
) {
	var target string
	var outbound interface{}

	switch event.Type {

	case EventTypeCallOffer,
		EventTypeCallAnswer:

		var payload CallSignalPayload

		if err := json.Unmarshal(
			event.Payload,
			&payload,
		); err != nil {
			return
		}

		if payload.TargetUserID == "" {
			return
		}

		target = payload.TargetUserID

		outbound = CallSignalPayload{
			FromUserID: msg.SenderID,
			SDP:        payload.SDP,
		}

	case EventTypeCallICECandidate:

		var payload CallICEPayload

		if err := json.Unmarshal(
			event.Payload,
			&payload,
		); err != nil {
			return
		}

		if payload.TargetUserID == "" {
			return
		}

		target = payload.TargetUserID

		outbound = CallICEPayload{
			FromUserID: msg.SenderID,
			Candidate:  payload.Candidate,
		}

	default:
		return
	}

	targetClient := h.findClient(
		msg.ChannelID,
		target,
	)

	if targetClient == nil {
		log.Printf(
			"[hub] Call target %s not found in channel %s",
			target,
			msg.ChannelID,
		)
		return
	}

	h.sendToClient(
		targetClient,
		OutboundEvent{
			Type:    event.Type,
			Payload: outbound,
		},
	)
}

// handleCallMute updates the sender's mute state
// and informs the other call participants.
func (h *Hub) handleCallMute(
	msg *BroadcastMessage,
	event InboundEvent,
) {
	var payload CallMutePayload

	if err := json.Unmarshal(
		event.Payload,
		&payload,
	); err != nil {
		return
	}

	client := h.findClient(
		msg.ChannelID,
		msg.SenderID,
	)

	if client == nil {
		return
	}

	participants := h.callParticipants[msg.ChannelID]

	if participants == nil {
		return
	}

	if _, ok := participants[msg.SenderID]; !ok {
		return
	}

	if h.callMuted[msg.ChannelID] == nil {
		h.callMuted[msg.ChannelID] = make(map[string]bool)
	}

	h.callMuted[msg.ChannelID][msg.SenderID] = payload.Muted

	h.broadcastToCallExcluding(
		msg.ChannelID,
		msg.SenderID,
		OutboundEvent{
			Type: EventTypeCallUserMuted,
			Payload: CallUserPayload{
				UserID: msg.SenderID,
				Muted:  payload.Muted,
			},
		},
	)
}

// handleCallCamera updates camera state
// and informs the other call participants.
func (h *Hub) handleCallCamera(
	msg *BroadcastMessage,
	event InboundEvent,
) {
	var payload CallCameraPayload

	if err := json.Unmarshal(
		event.Payload,
		&payload,
	); err != nil {
		return
	}

	client := h.findClient(
		msg.ChannelID,
		msg.SenderID,
	)

	if client == nil {
		return
	}

	participants := h.callParticipants[msg.ChannelID]

	if participants == nil {
		return
	}

	if _, ok := participants[msg.SenderID]; !ok {
		return
	}

	if h.callCameraOff[msg.ChannelID] == nil {
		h.callCameraOff[msg.ChannelID] = make(map[string]bool)
	}

	h.callCameraOff[msg.ChannelID][msg.SenderID] = payload.CameraOff

	h.broadcastToCallExcluding(
		msg.ChannelID,
		msg.SenderID,
		OutboundEvent{
			Type: EventTypeCallUserCamera,
			Payload: CallUserPayload{
				UserID:    msg.SenderID,
				CameraOff: payload.CameraOff,
			},
		},
	)
}

// findClient locates a connected client by channel + user ID.
//
// This is a linear scan, which is fine because a chat channel
// normally has a relatively small number of connected clients.
func (h *Hub) findClient(
	channelID uuid.UUID,
	userID string,
) *Client {
	for client := range h.clients[channelID] {
		if client.UserID == userID {
			return client
		}
	}

	return nil
}

// sendToClient delivers an event to exactly one client.
func (h *Hub) sendToClient(
	client *Client,
	event OutboundEvent,
) {
	data, err := encode(event)

	if err != nil {
		return
	}

	select {
	case client.send <- data:
	default:
		log.Printf(
			"[hub] Failed to send event to client %s",
			client.ID,
		)
	}
}

// broadcastToCallExcluding sends an event to every current
// call participant except one user.
func (h *Hub) broadcastToCallExcluding(
	channelID uuid.UUID,
	excludeUserID string,
	event OutboundEvent,
) {
	participants := h.callParticipants[channelID]

	if participants == nil {
		return
	}

	data, err := encode(event)

	if err != nil {
		return
	}

	for userID, client := range participants {
		if userID == excludeUserID {
			continue
		}

		select {
		case client.send <- data:
		default:
			// Don't block the Hub if a client isn't reading.
		}
	}
}
