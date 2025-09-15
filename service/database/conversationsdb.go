package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/dilcetto/wasa/service/components/schema"
	"github.com/gofrs/uuid"
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
				return nil, fmt.Errorf("failed to get private conversation info: %w", err)
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
		if err := memberRows.Err(); err != nil {
			return nil, err
		}
		_ = memberRows.Close()
		// Fetch last message
		var last schema.LastMessage
		var senderID string
		var ts string
		var attLen int
		err = db.c.QueryRow(`
			SELECT content, timestamp, senderId,
			       CASE WHEN attachment IS NOT NULL AND LENGTH(attachment) > 0 THEN 1 ELSE 0 END as attlen
			FROM messages
			WHERE conversationId = ?
			ORDER BY timestamp DESC LIMIT 1
		`, conv.ConversationID).Scan(&last.Preview, &ts, &senderID, &attLen)
		if err == nil {
			if t, perr := time.Parse(time.RFC3339, ts); perr == nil {
				last.Timestamp = t
			}
			if attLen > 0 {
				last.MessageType = "photo"
				if last.Preview == "" {
					last.Preview = "Photo"
				}
			} else {
				last.MessageType = "text"
			}
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
		WHERE c.id = ? AND cm.userId = ?`

	var conv schema.Conversation
	var convPhoto []byte

	err := db.c.QueryRow(query, conversationID, userID).Scan(&conv.ConversationID, &conv.DisplayName, &conv.Type, &conv.CreatedAt, &convPhoto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
			return nil, fmt.Errorf("failed to get private conversation info: %w", err)
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
	// check for errors during iteration and close rows
	if err := memberRows.Err(); err != nil {
		return nil, err
	}
	_ = memberRows.Close()

	// Fetch last message
	var last schema.LastMessage
	var senderID string
	var ts2 string
	var attLen2 int
	err = db.c.QueryRow(`
		SELECT content, timestamp, senderId,
		       CASE WHEN attachment IS NOT NULL AND LENGTH(attachment) > 0 THEN 1 ELSE 0 END as attlen
		FROM messages
		WHERE conversationId = ?
		ORDER BY timestamp DESC LIMIT 1
	`, conv.ConversationID).Scan(&last.Preview, &ts2, &senderID, &attLen2)
	if err == nil {
		if t, perr := time.Parse(time.RFC3339, ts2); perr == nil {
			last.Timestamp = t
		}
		if attLen2 > 0 {
			last.MessageType = "photo"
			if last.Preview == "" {
				last.Preview = "Photo"
			}
		} else {
			last.MessageType = "text"
		}
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
	var content string
	var senderID string
	var attachment []byte
	err := db.c.QueryRow(query, conversationID).Scan(&msg.ID, &content, &msg.Timestamp, &senderID, &attachment)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no messages found for conversation %s", conversationID)
		}
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}
	msg.SenderID = senderID
	msg.Content = schema.MessageContent{ContentType: schema.TextContent, Value: []byte(content)}
	if len(attachment) > 0 {
		msg.Attachments = []string{base64.StdEncoding.EncodeToString(attachment)}
	}
	return &msg, nil
}

// return the list of users in the conversation.
func (db *appdbimpl) GetConversationMembers(conversationID string) ([]schema.User, error) {
	rows, err := db.c.Query(`
        SELECT u.id, u.username, u.photo
        FROM users u
        JOIN conversation_members cm ON cm.userId = u.id
        WHERE cm.conversationId = ?
    `, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversation members: %w", err)
	}
	defer rows.Close()

	var users []schema.User
	for rows.Next() {
		var u schema.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Photo); err != nil {
			return nil, fmt.Errorf("failed to scan conversation member: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over conversation members: %w", err)
	}
	return users, nil
}

// EnsureDirectConversation returns an existing direct conversation between userID and other user,
// or creates a new one if none exists.
func (db *appdbimpl) EnsureDirectConversation(userID, peerUserID string) (*schema.Conversation, error) {
	// find existing direct conversation between the two users
	var conversationID string
	err := db.c.QueryRow(`
        SELECT c.id
        FROM conversations c
        JOIN conversation_members cm1 ON cm1.conversationId = c.id AND cm1.userId = ?
        JOIN conversation_members cm2 ON cm2.conversationId = c.id AND cm2.userId = ?
        WHERE c.type = 'direct'
        LIMIT 1
    `, userID, peerUserID).Scan(&conversationID)
	if err == nil {
		// return the full conversation
		return db.GetConversationByID(userID, conversationID)
	}

	// create a new direct conversation via the single write path
	convID, cerr := uuid.NewV4()
	if cerr != nil {
		return nil, fmt.Errorf("failed generating conversation id: %w", cerr)
	}
	conv := &schema.Conversation{
		ConversationID: convID.String(),
		DisplayName:    "direct",
		Type:           "direct",
		Members:        []string{userID, peerUserID},
	}
	if cerr = db.CreateConversation(conv); cerr != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", cerr)
	}
	return db.GetConversationByID(userID, convID.String())
}
