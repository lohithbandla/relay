package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	// PresenceTTL is how long a presence key lives without a heartbeat
	PresenceTTL = 90 * time.Second

	// PresenceKeyPrefix is the Redis key prefix for presence
	presenceKeyPrefix = "presence:"
)

// presenceKey generates the Redis key for a user's presence
func presenceKey(userID string) string {
	return fmt.Sprintf("%s%s", presenceKeyPrefix, userID)
}

// SetOnline marks a user as online in Redis with a TTL.
// Must be refreshed every PresenceTTL or user goes offline.
func SetOnline(ctx context.Context, userID string) error {
	return client.Set(ctx, presenceKey(userID), "online", PresenceTTL).Err()
}

// SetOffline immediately removes a user's presence key.
// Called when user explicitly disconnects.
func SetOffline(ctx context.Context, userID string) error {
	return client.Del(ctx, presenceKey(userID)).Err()
}

// IsOnline checks if a user is currently online.
func IsOnline(ctx context.Context, userID string) bool {
	result, err := client.Exists(ctx, presenceKey(userID)).Result()
	if err != nil {
		return false
	}
	return result > 0
}

// RefreshPresence resets the TTL for a user — called on heartbeat.
func RefreshPresence(ctx context.Context, userID string) error {
	return client.Expire(ctx, presenceKey(userID), PresenceTTL).Err()
}

// GetOnlineUsers checks which users from a given list are online.
// Used to get presence for all members of a server.
func GetOnlineUsers(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return []string{}, nil
	}

	onlineUsers := make([]string, 0)

	for _, userID := range userIDs {
		if IsOnline(ctx, userID) {
			onlineUsers = append(onlineUsers, userID)
		}
	}

	return onlineUsers, nil
}
