package database

import (
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

// AddReaction adds a reaction to a message.
func (db *appdbimpl) AddReactionToMessage(reaction *schema.Reaction) error {
	if reaction == nil {
		return fmt.Errorf("reaction is nil")
	}
	if reaction.MessageId == "" || reaction.UserId == "" || reaction.Emoji == "" {
		return fmt.Errorf("invalid reaction input")
	}

	query := `INSERT INTO reactions (messageId, userId, emoji) VALUES (?, ?, ?)`
	_, err := db.c.Exec(query, reaction.MessageId, reaction.UserId, reaction.Emoji)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

func (db *appdbimpl) DeleteReactionFromMessage(messageId, userId string) error {
	if messageId == "" || userId == "" {
		return fmt.Errorf("messageID and userID cannot be empty")
	}

	query := `DELETE FROM reactions WHERE messageId = ? AND userId = ?`
	_, err := db.c.Exec(query, messageId, userId)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	return nil
}
