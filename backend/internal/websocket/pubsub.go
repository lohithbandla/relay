package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	redispkg "github.com/lohithbandla/relay/internal/redis"
)

// SubscribeToChannel starts a Redis Pub/Sub subscription for a channel.
// It runs in its own goroutine and forwards messages to the hub's
// local clients via broadcastEvent.
//
// This is the KEY to multi-server scaling:
// - When any server publishes a message to Redis
// - ALL servers receive it here
// - And deliver it to their local WebSocket clients
func (h *Hub) SubscribeToChannel(channelID uuid.UUID) {
	ctx := context.Background()
	channelKey := redispkg.ChannelKey(channelID.String())

	pubsub := redispkg.Subscribe(ctx, channelKey)
	defer pubsub.Close()

	log.Printf("[pubsub] Subscribed to Redis channel: %s", channelKey)

	ch := pubsub.Channel()

	for msg := range ch {
		// The message is already a fully formed OutboundEvent
		// Just decode it directly — no double wrapping
		var outbound OutboundEvent
		if err := json.Unmarshal([]byte(msg.Payload), &outbound); err != nil {
			log.Printf("[pubsub] Failed to parse outbound event: %v", err)
			continue
		}

		// Broadcast to local WebSocket clients
		h.broadcastEvent(channelID, outbound, nil)
	}

	log.Printf("[pubsub] Unsubscribed from Redis channel: %s", channelKey)
}

// StartChannelSubscription starts a Redis subscription for a channel
// if one isn't already running.
// Called when the first client joins a channel.
func (h *Hub) StartChannelSubscription(channelID uuid.UUID) {
	h.subMu.Lock()
	defer h.subMu.Unlock()

	// Don't start duplicate subscriptions
	if h.subscriptions[channelID] {
		return
	}

	h.subscriptions[channelID] = true

	// Run subscription in its own goroutine
	// This goroutine lives as long as there are clients in this channel
	go h.SubscribeToChannel(channelID)
}

// StopChannelSubscription marks a channel subscription as inactive.
// Called when the last client leaves a channel.
func (h *Hub) StopChannelSubscription(channelID uuid.UUID) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	delete(h.subscriptions, channelID)
}
