package messages

import "github.com/google/uuid"

// SendMessageRequest is the body for POST /api/v1/channels/:channelID/messages
type SendMessageRequest struct {
	Content string      `json:"content"`
	Type    MessageType `json:"type"`
}

// MessageResponse is what we return — message + sender info combined
type MessageResponse struct {
	Message
	Sender MessageSender `json:"sender"`
}

// PaginatedMessages wraps a list of messages with pagination metadata
type PaginatedMessages struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
}

// GetMessagesQuery holds query params for fetching messages
type GetMessagesQuery struct {
	ChannelID uuid.UUID
	Limit     int
	Offset    int
}
