package schema

import (
	"time"
)

type Message struct {
	Content       MessageContent `json:"content"`            // The content of the message.
	Timestamp     time.Time      `json:"timestamp"`          // The time when the message was sent.
	Sender        string         `json:"sender"`             // The sender of the message.
	MessageStatus string         `json:"message_status"`     // The status of the message (e.g., sent, delivered, read).
	Reaction      []Reaction     `json:"reaction,omitempty"` // Optional reactions to the message.
}

type MessageContent struct {
	Type  string `json:"type"`  // "text" or "photo"
	Value string `json:"value"` // Text content or image URL
}
