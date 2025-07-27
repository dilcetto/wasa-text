package requests

import (
	"regexp"
)

type AddReactionRequest struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Emoji          string `json:"emoji"`
	Username       string `json:"username"`
}

func (r *AddReactionRequest) IsValid() bool {
	usernameMatch, _ := regexp.MatchString(`^.*?$`, r.Username)
	return len(r.ConversationID) > 0 && len(r.MessageID) > 0 && len(r.Emoji) > 0 && len(r.Emoji) <= 10 && usernameMatch
}

type RemoveReactionRequest struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Emoji          string `json:"emoji"`
	Username       string `json:"username"`
}

func (r *RemoveReactionRequest) IsValid() bool {
	usernameMatch, _ := regexp.MatchString(`^.*?$`, r.Username)
	return len(r.ConversationID) > 0 && len(r.MessageID) > 0 && len(r.Emoji) > 0 && len(r.Emoji) <= 10 && usernameMatch
}
