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

	// subscriptions tracks which channels have active Redis subscriptions
	// Protected by subMu since StartChannelSubscription runs in goroutines
	subscriptions map[uuid.UUID]bool
	subMu         sync.Mutex
}

// NewHub creates a new Hub.
func NewHub(messageService *messages.Service) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID]map[*Client]bool),
		broadcast:      make(chan *BroadcastMessage, 256),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		messageService: messageService,
		subscriptions:  make(map[uuid.UUID]bool),
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
			log.Printf("[hub] Client %s registered to channel %s", client.ID, client.ChannelID)

			// Start Redis subscription when first client joins channel
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
					log.Printf("[hub] Client %s unregistered from channel %s", client.ID, client.ChannelID)

					if len(clients) == 0 {
						delete(h.clients, client.ChannelID)
						// Stop Redis subscription when last client leaves
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
	default:
		log.Printf("[hub] Unknown event type: %s", event.Type)
	}
}

// handleSendMessage persists and publishes the message via Redis.
func (h *Hub) handleSendMessage(msg *BroadcastMessage, event InboundEvent) {
	var payload SendMessagePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return
	}

	if payload.Content == "" {
		return
	}

	senderID, err := uuid.Parse(msg.SenderID)
	if err != nil {
		return
	}

	// Save to PostgreSQL
	savedMessage, err := h.messageService.SendMessage(
		messages.SendMessageRequest{Content: payload.Content},
		msg.ChannelID,
		senderID,
	)
	if err != nil {
		log.Printf("[hub] Failed to save message: %v", err)
		return
	}

	// Build outbound event
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

	// Publish to Redis — all server instances will receive this
	// and deliver to their local clients
	data, err := json.Marshal(outbound)
	if err != nil {
		return
	}

	if err := redispkg.Publish(
		context.Background(),
		redispkg.ChannelKey(msg.ChannelID.String()),
		redispkg.PubSubMessage{
			Type:    string(EventTypeNewMessage),
			Payload: data,
		},
	); err != nil {
		log.Printf("[hub] Failed to publish to Redis: %v", err)
		h.broadcastEvent(msg.ChannelID, outbound, nil)
	}
}

// handleTyping broadcasts typing indicators locally only.
// Typing indicators are ephemeral — no need for Redis Pub/Sub.
func (h *Hub) handleTyping(msg *BroadcastMessage, event InboundEvent) {
	outbound := OutboundEvent{
		Type: event.Type,
		Payload: map[string]string{
			"user_id":  msg.SenderID,
			"username": msg.Username,
		},
	}
	h.broadcastEventExcludingSender(msg.ChannelID, outbound, msg.SenderID)
}

// broadcastEvent sends to all clients in a channel, optionally excluding one.
func (h *Hub) broadcastEvent(channelID uuid.UUID, event OutboundEvent, exclude *Client) {
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

// broadcastEventExcludingSender sends to all except the sender.
func (h *Hub) broadcastEventExcludingSender(channelID uuid.UUID, event OutboundEvent, senderID string) {
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
