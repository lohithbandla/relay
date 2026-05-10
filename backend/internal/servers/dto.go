package servers

// CreateServerRequest is the body for POST /api/v1/servers
type CreateServerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateServerResponse wraps the created server + its default channel
type CreateServerResponse struct {
	Server Server `json:"server"`
}
