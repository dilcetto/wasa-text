package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

func (db *appdbimpl) GetGroupByID(groupID string) (*schema.Group, error) {
	query := `SELECT id, name, photo, created_at FROM groups WHERE id = ?`
	row := db.c.QueryRow(query, groupID)

	var group schema.Group
	if err := row.Scan(&group.ID, &group.GroupName, &group.GroupPhoto, &group.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("group with ID %s not found", groupID)
		}
		return nil, fmt.Errorf("error retrieving group: %w", err)
	}

	return &group, nil
}

func (db *appdbimpl) GetMyGroups(userID string) ([]*schema.Group, error) {
	query := `SELECT g.id, g.name, g.photo, g.created_at 
			  FROM groups g 
			  JOIN group_members gm ON g.id = gm.group_id 
			  WHERE gm.user_id = ?`
	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving groups for user %s: %w", userID, err)
	}
	defer rows.Close()

	var groups []*schema.Group
	for rows.Next() {
		var group schema.Group
		if err := rows.Scan(&group.ID, &group.GroupName, &group.GroupPhoto, &group.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning group row: %w", err)
		}
		groups = append(groups, &group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over groups: %w", err)
	}

	return groups, nil
}

func (db *appdbimpl) CreateGroup(group *schema.Group) error {
	query := `INSERT INTO groups (id, name, photo, created_at) VALUES (?, ?, ?, ?)`
	_, err := db.c.Exec(query, group.ID, group.GroupName, group.GroupPhoto, group.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	return nil
}

func (db *appdbimpl) UpdateGroupName(groupID, newName string) error {
	query := `UPDATE groups SET name = ? WHERE id = ?`
	_, err := db.c.Exec(query, newName, groupID)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}

	return nil
}

func (db *appdbimpl) UpdateGroupPhoto(groupID string, photo []byte) error {
	query := `UPDATE groups SET photo = ? WHERE id = ?`
	_, err := db.c.Exec(query, photo, groupID)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}

	return nil
}

func (db *appdbimpl) AddUserToGroup(groupID, userID string) error {
	query := `INSERT INTO group_members (group_id, user_id) VALUES (?, ?)`
	_, err := db.c.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("error adding user %s to group %s: %w", userID, groupID, err)
	}

	return nil
}

func (db *appdbimpl) LeaveGroup(groupID, userID string) error {
	query := `DELETE FROM group_members WHERE group_id = ? AND user_id = ?`
	result, err := db.c.Exec(query, groupID, userID)
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
