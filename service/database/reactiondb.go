package database

import (
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) AddReactionToMessage(reaction *schema.Reaction) error {
	if reaction == nil {
		return fmt.Errorf("reaction is nil")
	}
	if reaction.MessageId == "" || reaction.UserId == "" || reaction.Emoji == "" {
		return fmt.Errorf("invalid reaction input")
	}

	// Upsert: one reaction per (messageId, userId). If it exists, update the emoji.
	query := `
    INSERT INTO reactions (messageId, userId, reaction)
    VALUES (?, ?, ?)
    ON CONFLICT(messageId, userId) DO UPDATE SET
      reaction = excluded.reaction`
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
