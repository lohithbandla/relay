package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// PubSubMessage is the envelope for all Redis Pub/Sub messages.
// Every message published to Redis uses this structure.
type PubSubMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Publish sends a message to a Redis channel.
// channelKey is typically "chat:<channelID>"
func Publish(ctx context.Context, channelKey string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return client.Publish(ctx, channelKey, data).Err()
}

// Subscribe returns a Redis subscription to a channel.
// The caller is responsible for reading from the subscription
// and closing it when done.
func Subscribe(ctx context.Context, channelKey string) *redis.PubSub {
	return client.Subscribe(ctx, channelKey)
}

// ChannelKey generates a consistent Redis channel key for a chat channel.
// All instances use the same key format so they can find each other.
func ChannelKey(channelID string) string {
	return "chat:" + channelID
}
