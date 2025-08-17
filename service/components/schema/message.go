package schema

type Sender struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Message struct {
	ID            string         `json:"id"`
	Sender        Sender         `json:"sender"`
	Content       MessageContent `json:"content"`
	Timestamp     string         `json:"timestamp"`
	MessageStatus string         `json:"message_status"` // "sent", "delivered", "read"
	Reaction      []Reaction     `json:"reaction,omitempty"`
	Attachments   []string       `json:"attachments,omitempty"`
	ForwardedFrom string         `json:"forwarded_from,omitempty"`
}

type ContentType string

const (
	TextContent ContentType = "text"
	Image       ContentType = "image"
)

type MessageContent struct {
	ContentType ContentType `json:"type"`  // Type of content (text, image, video, etc.)
	Value       string      `json:"value"` // Text content or image
}

type Reaction struct {
	Emoji    string `json:"emoji"`    // The emoji used for the reaction.
	Username string `json:"username"` // The username of the user who reacted.
}
