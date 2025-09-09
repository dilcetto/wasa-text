package database

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) SendMessage(message *schema.Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	var attachment []byte
	if len(message.Attachments) > 0 {
		if dec, derr := base64.StdEncoding.DecodeString(message.Attachments[0]); derr == nil {
			attachment = dec
		} else {
			attachment = []byte(message.Attachments[0])
		}
	}
	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.c.Exec(query, message.ID, message.ConversationID, message.SenderID, string(message.Content.Value), message.Timestamp, attachment, message.MessageStatus, message.ForwardedFrom)
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
		var content string
		var attachment []byte
		// Scan row into vars and then populate msg
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &content, &msg.Timestamp,
			&attachment, &msg.MessageStatus, &msg.ForwardedFrom,
			&senderName, &senderPhoto,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		// Content
		// always keep text content
		msg.Content.Value = []byte(content)
		if len(attachment) > 0 {
			msg.Content.ContentType = schema.Image
			msg.MessageType = string(schema.Image)
			msg.Attachments = []string{base64.StdEncoding.EncodeToString(attachment)}
		} else {
			msg.Content.ContentType = schema.TextContent
			msg.MessageType = string(schema.TextContent)
		}
		// Sender
		msg.Sender.ID = msg.SenderID
		msg.Sender.Username = senderName
		msg.Sender.Photo = senderPhoto
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading messages: %w", err)
	}

	// load reactions in batch
	if len(messages) > 0 {
		idx := make(map[string]*schema.Message, len(messages))
		placeholders := make([]string, 0, len(messages))
		args := make([]interface{}, 0, len(messages))
		for _, m := range messages {
			idx[m.ID] = m
			placeholders = append(placeholders, "?")
			args = append(args, m.ID)
		}
		q := "SELECT r.messageId, r.userId, r.reaction, u.username FROM reactions r JOIN users u ON u.id = r.userId WHERE r.messageId IN (" + strings.Join(placeholders, ",") + ")"
		rr, err := db.c.Query(q, args...)
		if err == nil {
			defer rr.Close()
			for rr.Next() {
				var mid, uid, emoji, uname string
				if err := rr.Scan(&mid, &uid, &emoji, &uname); err == nil {
					if m := idx[mid]; m != nil {
						m.Reaction = append(m.Reaction, schema.Reaction{MessageId: mid, UserId: uid, Emoji: emoji, Username: uname})
					}
				}
			}
			if err := rr.Err(); err != nil {
				// ignore reaction load errors to not fail the whole call
			}
		}
		// compute aggregate delivery/read based on all recipients in the conversation
		// recipients are all members except the sender
		recipients := 0
		if row := db.c.QueryRow(`SELECT COUNT(*) FROM conversation_members WHERE conversationId = ?`, conversationID); row != nil {
			_ = row.Scan(&recipients)
			if recipients > 0 {
				recipients -= 1
			}
		}
		if recipients < 0 {
			recipients = 0
		}

		qs := "SELECT message_id, " +
			"SUM(CASE WHEN status = 'read' THEN 1 ELSE 0 END) AS rc, " +
			"SUM(CASE WHEN status IN ('read','delivered') THEN 1 ELSE 0 END) AS dc " +
			"FROM message_receipts WHERE message_id IN (" + strings.Join(placeholders, ",") + ") GROUP BY message_id"
		rs, err := db.c.Query(qs, args...)
		if err == nil {
			defer rs.Close()
			for rs.Next() {
				var mid string
				var rc, dc int
				if err := rs.Scan(&mid, &rc, &dc); err == nil {
					if m := idx[mid]; m != nil && recipients > 0 {
						if rc >= recipients {
							m.MessageStatus = "read"
						} else if dc >= recipients {
							m.MessageStatus = "delivered"
						}
					}
				}
			}
			_ = rs.Err()
		}
	}

	return messages, nil
}

func (db *appdbimpl) GetMessageByID(messageID string) (*schema.Message, error) {
	query := `SELECT m.id, m.conversationId, m.senderId, m.content, m.timestamp, m.attachment, m.status, m.forwardedFrom,
					 u.username, u.photo
			FROM messages m
			JOIN users u ON u.id = m.senderId
			WHERE m.id = ?`

	row := db.c.QueryRow(query, messageID)

	var message schema.Message
	var attachment []byte
	var senderName string
	var senderPhoto []byte
	err := row.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.Content.Value, &message.Timestamp, &attachment, &message.MessageStatus, &message.ForwardedFrom, &senderName, &senderPhoto)
	if err != nil {
		return nil, err
	}

	if len(attachment) > 0 {
		message.Attachments = []string{base64.StdEncoding.EncodeToString(attachment)}
		message.Content.ContentType = schema.Image
		message.MessageType = string(schema.Image)
	} else {
		message.Content.ContentType = schema.TextContent
		message.MessageType = string(schema.TextContent)
	}

	message.Sender = schema.Sender{
		ID:       message.SenderID,
		Username: senderName,
		Photo:    senderPhoto,
	}

	// load reactions for this message
	rr, rerr := db.c.Query(`SELECT r.userId, r.reaction, u.username FROM reactions r JOIN users u ON u.id = r.userId WHERE r.messageId = ?`, messageID)
	if rerr == nil {
		defer rr.Close()
		for rr.Next() {
			var uid, emoji, uname string
			if err := rr.Scan(&uid, &emoji, &uname); err == nil {
				message.Reaction = append(message.Reaction, schema.Reaction{MessageId: messageID, UserId: uid, Emoji: emoji, Username: uname})
			}
		}
		// Ensure any iteration error is surfaced
		if err := rr.Err(); err != nil {
			return nil, fmt.Errorf("error reading reactions: %w", err)
		}
	}

	return &message, nil
}

func (db *appdbimpl) ForwardMessage(message *schema.Message, userID string) error {
	if message == nil || userID == "" || message.ForwardedFrom == "" {
		return fmt.Errorf("message, user ID, and forwardedFrom cannot be empty")
	}

	// optionally fetch original content if message.Content is empty
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

	var attachment []byte
	if len(message.Attachments) > 0 {
		if dec, derr := base64.StdEncoding.DecodeString(message.Attachments[0]); derr == nil {
			attachment = dec
		} else {
			attachment = []byte(message.Attachments[0])
		}
	}
	query := `INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, status, forwardedFrom) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.c.Exec(query, message.ID, message.ConversationID, userID, string(message.Content.Value), message.Timestamp, attachment, message.MessageStatus, message.ForwardedFrom)
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
