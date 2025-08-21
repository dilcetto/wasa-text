package schema

import "time"

type Conversation struct {
	ConversationID string       `json:"conversationId"`
	DisplayName    string       `json:"displayName"`
	ProfilePhoto   []byte       `json:"profilePhoto,omitempty"`
	Type           string       `json:"type"`
	CreatedAt      string       `json:"createdAt"`
	Members        []string     `json:"membersIds"`
	Messages       []*Message   `json:"messages,omitempty"`
	LastMessage    *LastMessage `json:"lastMessage,omitempty"`
}

type LastMessage struct {
	MessageType string    `json:"messageType"`
	Preview     string    `json:"preview"`
	Timestamp   time.Time `json:"timestamp"`
}

type Group struct {
	ID         string   `json:"id"`
	GroupName  string   `json:"group_name"`
	GroupPhoto []byte   `json:"group_photo"`
	Members    []string `json:"members"`
	CreatedAt  string   `json:"createdAt"`
}
