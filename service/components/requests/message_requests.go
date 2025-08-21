package requests

type SendMessageRequest struct {
	ConversationID string         `json:"conversation_id"`
	Content        MessageContent `json:"content"`
	Sender         string         `json:"sender"`
}

func (s *SendMessageRequest) IsValid() bool {
	return len(s.ConversationID) >= 1 && len(s.ConversationID) <= 36 &&
		s.Content.IsValid() &&
		len(s.Sender) >= 3 && len(s.Sender) <= 16
}

type MessageContent struct {
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

func (c *MessageContent) IsValid() bool {
	if c.Type == "text" {
		return len(c.Value) >= 1 && len(c.Value) <= 500
	}
	if c.Type == "photo" {
		return len(c.Value) >= 1 && len(c.Value) <= (5*1024*1024) // e.g., up to 5MB
	}
	return false
}

type ForwardMessageRequest struct {
	ConversationID       string `json:"conversation_id"`
	MessageID            string `json:"message_id"`
	TargetConversationID string `json:"target_conversation_id"`
}

func (f *ForwardMessageRequest) IsValid() bool {
	return len(f.ConversationID) >= 1 && len(f.ConversationID) <= 36 &&
		len(f.MessageID) >= 1 && len(f.MessageID) <= 36 &&
		len(f.TargetConversationID) >= 1 && len(f.TargetConversationID) <= 36
}

type DeleteMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	UserID         string `json:"user_id"`
}

func (d *DeleteMessageRequest) IsValid() bool {
	return len(d.ConversationID) >= 1 && len(d.ConversationID) <= 36 &&
		len(d.MessageID) >= 1 && len(d.MessageID) <= 36 &&
		len(d.UserID) >= 3 && len(d.UserID) <= 16
}
