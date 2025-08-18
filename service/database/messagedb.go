package database

import (
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) SendMessage(message *schema.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, replyTo, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.c.Exec(query, message.ID, message.ConversationID, message.SenderID, message.Content, message.Timestamp, message.Attachment, message.Status, message.ReplyTo, message.ForwardedFrom)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (db *appdbimpl) GetMessagesByConversationID(conversationID string) ([]*schema.Message, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("conversation ID cannot be empty")
	}

	query := `
	SELECT 
	  m.id, m.conversationId, m.senderId, m.content, m.timestamp, 
	  m.attachment, m.status, m.replyTo, m.forwardedFrom,
	  u.name, u.photo
	FROM messages m
	JOIN users u ON m.senderId = u.id
	WHERE m.conversationId = ?
	ORDER BY m.timestamp ASC`
	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*schema.Message
	for rows.Next() {
		var msg schema.Message
		var senderName, senderPhoto string
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Timestamp,
			&msg.Attachment, &msg.Status, &msg.ReplyTo, &msg.ForwardedFrom,
			&senderName, &senderPhoto,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		// Optionally set these if supported by schema.Message
		msg.SenderName = senderName
		msg.SenderPhoto = senderPhoto
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading messages: %w", err)
	}

	return messages, nil
}

func (db *appdbimpl) ForwardMessage(message *schema.Message, userID string) error {
	if message == nil || userID == "" || message.ForwardedFrom == "" {
		return fmt.Errorf("message, user ID, and forwardedFrom cannot be empty")
	}

	// Optionally fetch original content if message.Content is empty
	if message.Content == "" {
		var originalContent string
		query := `SELECT content FROM messages WHERE id = ?`
		err := db.c.QueryRow(query, message.ForwardedFrom).Scan(&originalContent)
		if err != nil {
			return fmt.Errorf("original message not found: %w", err)
		}
		message.Content = originalContent
	}

	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, replyTo, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.c.Exec(query, message.ID, message.ConversationID, userID, message.Content, message.Timestamp, message.Attachment, message.Status, message.ReplyTo, message.ForwardedFrom)
	if err != nil {
		return fmt.Errorf("failed to forward message: %w", err)
	}
	return nil
}

func (db *appdbimpl) DeleteMessage(conversationID, messageID, userID string) error {
	if conversationID == "" || messageID == "" || userID == "" {
		return fmt.Errorf("conversation ID, message ID, and user ID cannot be empty")
	}

	query := `DELETE FROM messages WHERE id = ? AND conversationId = ? AND senderId = ?`
	result, err := db.c.Exec(query, messageID, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no message deleted, possibly due to non-matching sender or invalid ID")
	}

	return nil
}

func (db *appdbimpl) MarkMessageStatus(messageID, userID, status string) error {
	if messageID == "" || userID == "" || (status != "delivered" && status != "read") {
		return fmt.Errorf("invalid input")
	}

	query := `
	INSERT INTO message_receipts (message_id, user_id, status)
	VALUES (?, ?, ?)
	ON CONFLICT(message_id, user_id) DO UPDATE SET
	status = excluded.status,
	timestamp = CURRENT_TIMESTAMP`

	_, err := db.c.Exec(query, messageID, userID, status)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	return nil
}
