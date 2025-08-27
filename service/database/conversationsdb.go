package database

import (
	"database/sql"
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) GetMyConversations(userID string) ([]*schema.Conversation, error) {
	query := `
		SELECT c.id, c.name, c.type, c.created_at, c.conversationPhoto
		FROM conversations c
		JOIN conversation_members cm ON cm.conversationId = c.id
		WHERE cm.userId = ?`

	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*schema.Conversation

	for rows.Next() {
		var conv schema.Conversation
		var convPhoto []byte
		if err := rows.Scan(&conv.ConversationID, &conv.DisplayName, &conv.Type, &conv.CreatedAt, &convPhoto); err != nil {
			return nil, err
		}

		// Set the profile photo if it exists
		if conv.Type == "direct" {
			// Get the other user's info
			err = db.c.QueryRow(`
				SELECT u.username, u.photo
				FROM users u
				JOIN conversation_members cm ON cm.userId = u.id
				WHERE cm.conversationId = ? AND cm.userId != ?
			`, conv.ConversationID, userID).Scan(&conv.DisplayName, &conv.ProfilePhoto)
			if err != nil {
				return nil, fmt.Errorf("failed to get direct conversation info: %w", err)
			}
		} else {
			// For group conversations, set the profile photo to the group photo
			conv.ProfilePhoto = convPhoto
		}
		// Fetch members
		memberRows, err := db.c.Query(`SELECT userId FROM conversation_members WHERE conversationId = ?`, conv.ConversationID)
		if err != nil {
			return nil, err
		}
		for memberRows.Next() {
			var memberID string
			if err := memberRows.Scan(&memberID); err != nil {
				return nil, err
			}
			conv.Members = append(conv.Members, memberID)
		}
		memberRows.Close()
		// Fetch last message
		var last schema.LastMessage
		var senderID string
		err = db.c.QueryRow(`
			SELECT content, timestamp, senderId
			FROM messages
			WHERE conversationId = ?
			ORDER BY timestamp DESC LIMIT 1
		`, conv.ConversationID).Scan(&last.Preview, &last.Timestamp, &senderID)
		if err == nil {
			last.MessageType = "text" // adjust for later message types
			conv.LastMessage = &last
		}

		conversations = append(conversations, &conv)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return conversations, nil
}

func (db *appdbimpl) GetConversationByID(userID, conversationID string) (*schema.Conversation, error) {
	query := `
		SELECT c.id, c.name, c.type, c.created_at, c.conversationPhoto
		FROM conversations c
		JOIN conversation_members cm ON cm.conversationId = c.id
		WHERE cm.conversationId = ? AND cm.userId = ?`

	var conv schema.Conversation
	var convPhoto []byte

	err := db.c.QueryRow(query, conversationID, userID).Scan(&conv.ConversationID, &conv.DisplayName, &conv.Type, &conv.CreatedAt, &convPhoto)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if conv.Type == "direct" {
		// Get the other user's info
		err = db.c.QueryRow(`
			SELECT u.username, u.photo
			FROM users u
			JOIN conversation_members cm ON cm.userId = u.id
			WHERE cm.conversationId = ? AND cm.userId != ?
		`, conv.ConversationID, userID).Scan(&conv.DisplayName, &conv.ProfilePhoto)
		if err != nil {
			return nil, fmt.Errorf("failed to get direct conversation info: %w", err)
		}
	} else {
		// For group conversations, set the profile photo to the group photo
		conv.ProfilePhoto = []byte(convPhoto)
	}

	// Fetch members
	memberRows, err := db.c.Query(`SELECT userId FROM conversation_members WHERE conversationId = ?`, conv.ConversationID)
	if err != nil {
		return nil, err
	}
	for memberRows.Next() {
		var memberID string
		if err := memberRows.Scan(&memberID); err != nil {
			return nil, err
		}
		conv.Members = append(conv.Members, memberID)
	}
	memberRows.Close()

	// Fetch last message
	var last schema.LastMessage
	var senderID string
	err = db.c.QueryRow(`
		SELECT content, timestamp, senderId
		FROM messages
		WHERE conversationId = ?
		ORDER BY timestamp DESC LIMIT 1
	`, conv.ConversationID).Scan(&last.Preview, &last.Timestamp, &senderID)
	if err == nil {
		last.MessageType = "text" // adjust for later message types
		conv.LastMessage = &last
	}

	return &conv, nil
}

func (db *appdbimpl) SearchConversationByName(name string) ([]schema.Conversation, error) {
	rows, err := db.c.Query("SELECT id, name, type, created_at, conversationPhoto FROM conversations WHERE name LIKE '%' || ? || '%'", name)
	if err != nil {
		return nil, fmt.Errorf("failed to search conversations by name: %w", err)
	}
	defer rows.Close()

	var conversations []schema.Conversation
	for rows.Next() {
		var conv schema.Conversation
		if err := rows.Scan(&conv.ConversationID, &conv.DisplayName, &conv.Type, &conv.CreatedAt, &conv.ProfilePhoto); err != nil {
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}
		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over conversation rows: %w", err)
	}

	return conversations, nil
}

// CreateConversation inserts a new conversation into the database.
// This method is kept explicitly to be reused by other internal logic
// (e.g., auto-creating direct conversations during SendMessage) or
// for explicit group chat creation if needed in the future.
func (db *appdbimpl) CreateConversation(conversation *schema.Conversation) error {
	query := `
		INSERT INTO conversations (id, name, type, created_at, conversationPhoto)
		VALUES (?, ?, ?, datetime('now'), ?)`

	_, err := db.c.Exec(query, conversation.ConversationID, conversation.DisplayName, conversation.Type, conversation.ProfilePhoto)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	for _, memberID := range conversation.Members {
		_, err = db.c.Exec(`INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)`, conversation.ConversationID, memberID)
		if err != nil {
			return fmt.Errorf("failed to add member to conversation: %w", err)
		}
	}

	return nil
}

func (db *appdbimpl) GetLastMessageByConversationID(conversationID string) (*schema.Message, error) {
	query := `
		SELECT id, content, timestamp, senderId, attachment
		FROM messages
		WHERE conversationId = ?
		ORDER BY timestamp DESC LIMIT 1`

	var msg schema.Message
	var attachment []byte
	err := db.c.QueryRow(query, conversationID).Scan(&msg.ID, &msg.Content.Value, &msg.Timestamp, &msg.SenderID, &attachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no messages found for conversation %s", conversationID)
		}
		return nil, fmt.Errorf("failed to get last message: %w", err)
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

	return &msg, nil
}
