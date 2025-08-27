package database

import (
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) SendMessage(message *schema.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	var attachment []byte
	if len(message.Attachments) > 0 {
		attachment = []byte(message.Attachments[0])
	}
	_, err := db.c.Exec(query, message.ID, message.ConversationID, message.SenderID, message.Content.Value, message.Timestamp, attachment, message.MessageStatus, message.ForwardedFrom)
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
	  m.attachment, m.status, m.forwardedFrom,
	  u.username, u.photo
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
	var senderName string
	var senderPhoto []byte
	for rows.Next() {
		var msg schema.Message
		var attachment []byte
		// Scan row into msg and senderName, senderPhoto
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Timestamp,
			&attachment, &msg.MessageStatus, &msg.ForwardedFrom,
			&senderName, &senderPhoto,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		if len(attachment) > 0 {
			msg.Attachments = []string{string(attachment)}
			msg.Content.ContentType = schema.Image
			msg.MessageType = string(schema.Image)
		} else {
			msg.Content.ContentType = schema.TextContent
			msg.MessageType = string(schema.TextContent)
		}

		msg.Sender.ID = msg.SenderID
		msg.Sender.Username = senderName
		msg.Sender.Photo = senderPhoto
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading messages: %w", err)
	}

	return messages, nil
}

func (db *appdbimpl) GetMessageByID(messageID string) (*schema.Message, error) {
	query := `SELECT id, conversationId, senderId, content, timestamp, attachment, status, forwardedFrom 
			  FROM messages WHERE id = ?`

	row := db.c.QueryRow(query, messageID)

	var message schema.Message
	var attachment []byte
	err := row.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.Content.Value, &message.Timestamp, &attachment, &message.MessageStatus, &message.ForwardedFrom)
	if err != nil {
		return nil, err
	}

	if len(attachment) > 0 {
		message.Attachments = []string{string(attachment)} // Or a proper file path/identifier
		message.Content.ContentType = schema.Image
		message.MessageType = string(schema.Image)
	} else {
		message.Content.ContentType = schema.TextContent
		message.MessageType = string(schema.TextContent)
	}

	message.Sender = schema.Sender{
		ID:       message.SenderID,
		Username: message.Sender.Username,
		Photo:    message.Sender.Photo,
	}

	return &message, nil
}

func (db *appdbimpl) ForwardMessage(message *schema.Message, userID string) error {
	if message == nil || userID == "" || message.ForwardedFrom == "" {
		return fmt.Errorf("message, user ID, and forwardedFrom cannot be empty")
	}

	// Optionally fetch original content if message.Content is empty
	if len(message.Content.Value) == 0 && message.ForwardedFrom != "" {
		var originalContent string
		query := `SELECT content FROM messages WHERE id = ?`
		err := db.c.QueryRow(query, message.ForwardedFrom).Scan(&originalContent)
		if err != nil {
			return fmt.Errorf("original message not found: %w", err)
		}
		message.Content = schema.MessageContent{
			ContentType: schema.TextContent,
			Value:       []byte(originalContent),
		}

	}

	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	var attachment []byte
	if len(message.Attachments) > 0 {
		attachment = []byte(message.Attachments[0])
	}
	_, err := db.c.Exec(query, message.ID, message.ConversationID, userID, message.Content.Value, message.Timestamp, attachment, message.MessageStatus, message.ForwardedFrom)
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
