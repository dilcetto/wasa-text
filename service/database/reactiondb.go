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
	if reaction.MessageID == "" || reaction.UserID == "" || reaction.Emoji == "" {
		return fmt.Errorf("invalid reaction input")
	}

	query := `INSERT INTO reactions (messageId, userId, emoji) VALUES (?, ?, ?)`
	_, err := db.c.Exec(query, reaction.MessageID, reaction.UserID, reaction.Emoji)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}
	return nil
}

// DeleteReactionFromMessage removes a specific user's reaction from a message.
func (db *appdbimpl) DeleteReactionFromMessage(messageID, userID string) error {
	if messageID == "" || userID == "" {
		return fmt.Errorf("messageID and userID cannot be empty")
	}

	query := `DELETE FROM reactions WHERE messageId = ? AND userId = ?`
	_, err := db.c.Exec(query, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}
	return nil
}
