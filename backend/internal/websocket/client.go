package websocket

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	redispkg "github.com/lohithbandla/relay/internal/redis"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096

	// HeartbeatInterval is how often client refreshes presence
	heartbeatInterval = 30 * time.Second
)

// Client represents a single WebSocket connection.
type Client struct {
	ID        string
	conn      *websocket.Conn
	UserID    string
	Username  string
	ChannelID uuid.UUID
	send      chan []byte
	hub       *Hub
}

// NewClient creates a new WebSocket client.
func NewClient(conn *websocket.Conn, userID, username string, channelID uuid.UUID, hub *Hub) *Client {
	return &Client{
		ID:        uuid.New().String(),
		conn:      conn,
		UserID:    userID,
		Username:  username,
		ChannelID: channelID,
		send:      make(chan []byte, 256),
		hub:       hub,
	}
}

// ReadPump pumps messages from WebSocket to hub.
func (c *Client) ReadPump() {
	defer func() {
		// Mark user offline when they disconnect
		if err := redispkg.SetOffline(context.Background(), c.UserID); err != nil {
			log.Printf("[presence] Failed to set offline for %s: %v", c.UserID, err)
		}
		log.Printf("[presence] User %s is now offline", c.Username)

		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Mark user online when they connect
	if err := redispkg.SetOnline(context.Background(), c.UserID); err != nil {
		log.Printf("[presence] Failed to set online for %s: %v", c.UserID, err)
	}
	log.Printf("[presence] User %s is now online", c.Username)

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("[websocket] Unexpected close for client %s: %v", c.ID, err)
			}
			break
		}

		c.hub.broadcast <- &BroadcastMessage{
			ChannelID: c.ChannelID,
			SenderID:  c.UserID,
			Username:  c.Username,
			Data:      message,
		}
	}
}

// WritePump pumps messages from hub to WebSocket.
// Also handles presence heartbeat via a ticker.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	// Separate ticker for presence heartbeat
	heartbeat := time.NewTicker(heartbeatInterval)

	defer func() {
		ticker.Stop()
		heartbeat.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-heartbeat.C:
			// Refresh presence TTL every 30 seconds
			// This keeps the user "online" as long as connection is alive
			if err := redispkg.RefreshPresence(context.Background(), c.UserID); err != nil {
				log.Printf("[presence] Failed to refresh presence for %s: %v", c.UserID, err)
			}
		}
	}
}
