package requests

type Conversation struct {
	ConversationID string      `json:"conversation_id"`
	DisplayName    string      `json:"display_name"`
	ProfilePhoto   string      `json:"profile_photo"`
	LastMessage    LastMessage `json:"last_message"`
}

type LastMessage struct {
	MessageType string `json:"message_type"` // "text" or "photo"
	Preview     string `json:"preview"`
	Timestamp   string `json:"timestamp"` // ISO 8601 format
}
