package schema

type Sender struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Photo    []byte `json:"photo,omitempty"`
}

type Message struct {
	ID             string         `json:"id"`
	Sender         Sender         `json:"sender"`
	SenderID       string         `json:"senderId"`
	ConversationID string         `json:"conversationId"`
	MessageType    string         `json:"messageType"`
	Content        MessageContent `json:"content"`
	Timestamp      string         `json:"timestamp"`
	MessageStatus  string         `json:"message_status"`
	Reaction       []Reaction     `json:"reaction,omitempty"`
	Attachments    []string       `json:"attachments,omitempty"`
	ForwardedFrom  string         `json:"forwarded_from,omitempty"`
}

type ContentType string

const (
	TextContent ContentType = "text"
	// Keep constant name for backward compatibility, but align value to OpenAPI ('photo')
	Image ContentType = "photo"
)

type MessageContent struct {
	ContentType ContentType `json:"type"`
	Value       []byte      `json:"value"`
}

type Reaction struct {
	MessageId string `json:"message_id"` // ID of the message this reaction belongs to.
	UserId    string `json:"user_id"`    // ID of the user who reacted.
	Emoji     string `json:"emoji"`      // The emoji used for the reaction.
	Username  string `json:"username"`   // The username of the user who reacted.
}
