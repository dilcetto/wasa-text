package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) GetGroupByID(groupID string) (*schema.Group, error) {
	// fetch conversation as a group
	var g schema.Group
	err := db.c.QueryRow(`SELECT id, name, conversationPhoto, created_at FROM conversations WHERE id = ? AND type = 'group'`, groupID).
		Scan(&g.ID, &g.GroupName, &g.GroupPhoto, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("group with ID %s not found", groupID)
		}
		return nil, fmt.Errorf("error retrieving group: %w", err)
	}
	return &g, nil
}

func (db *appdbimpl) GetMyGroups(userID string) ([]*schema.Group, error) {
	rows, err := db.c.Query(`
        SELECT c.id, c.name, c.conversationPhoto, c.created_at
        FROM conversations c
        JOIN conversation_members cm ON cm.conversationId = c.id
        WHERE cm.userId = ? AND c.type = 'group'`, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving groups for user %s: %w", userID, err)
	}
	defer rows.Close()

	var groups []*schema.Group
	for rows.Next() {
		var g schema.Group
		if err := rows.Scan(&g.ID, &g.GroupName, &g.GroupPhoto, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning group row: %w", err)
		}
		groups = append(groups, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over groups: %w", err)
	}
	return groups, nil
}

func (db *appdbimpl) CreateGroup(group *schema.Group) error {
	// mapping schema.Group to schema.Conversation and reuse CreateConversation
	conv := &schema.Conversation{
		ConversationID: group.ID,
		DisplayName:    group.GroupName,
		Type:           "group",
		CreatedAt:      group.CreatedAt,
		Members:        group.Members,
	}
	// store photo if present
	conv.ProfilePhoto = group.GroupPhoto
	return db.CreateConversation(conv)
}

func (db *appdbimpl) UpdateGroupName(groupID, newName string) error {
	_, err := db.c.Exec(`UPDATE conversations SET name = ? WHERE id = ? AND type = 'group'`, newName, groupID)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}
	return nil
}

func (db *appdbimpl) UpdateGroupPhoto(groupID string, photo []byte) error {
	_, err := db.c.Exec(`UPDATE conversations SET conversationPhoto = ? WHERE id = ? AND type = 'group'`, photo, groupID)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}
	return nil
}

func (db *appdbimpl) AddUserToGroup(groupID, userID string) error {
	_, err := db.c.Exec(`INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)`, groupID, userID)
	if err != nil {
		return fmt.Errorf("error adding user %s to group %s: %w", userID, groupID, err)
	}
	return nil
}

func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	result, err := db.c.Exec(`DELETE FROM conversation_members WHERE conversationId = ? AND userId = ?`, groupID, userID)
	if err != nil {
		return fmt.Errorf("error leaving group %s: %w", groupID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user %s is not a member of group %s", userID, groupID)
	}
	return nil
}
