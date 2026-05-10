package channels

// CreateChannelRequest is the body for POST /api/v1/servers/:serverID/channels
type CreateChannelRequest struct {
	Name      string      `json:"name"`
	Topic     string      `json:"topic"`
	Type      ChannelType `json:"type"`
	IsPrivate bool        `json:"is_private"`
}
