package schema

import "time"

type Conversation struct {
	ConversationId string       `json:"conversationId"`
	DisplayName    string       `json:"displayName"`
	ProfilePhoto   string       `json:"profilePhoto"`
	LastMessage    *LastMessage `json:"lastMessage"`
}

type LastMessage struct {
	MessageType string    `json:"messageType"`
	Preview     string    `json:"preview"`
	Timestamp   time.Time `json:"timestamp"`
}
