package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

var ErrUserDoesNotExist = errors.New("user does not exist")

func (db *appdbimpl) SearchUserByName(username string) ([]schema.User, error) {
	rows, err := db.c.Query("SELECT id, username, photoURL FROM users WHERE username LIKE '%' || ? || '%'", username)
	if err != nil {
		return nil, fmt.Errorf("failed to search users by name: %w", err)
	}
	defer rows.Close()

	var users []schema.User
	for rows.Next() {
		var user schema.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Photo); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over user rows: %w", err)
	}

	return users, nil
}

func (db *appdbimpl) CreateUser(u *schema.User) error {
	// Check if user with the same name already exists
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", u.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return fmt.Errorf("username %s already exists", u.Username)
	}

	// Attempt to insert the new user
	_, err = db.c.Exec("INSERT INTO users(id, username, photo) VALUES (?, ?, ?)", u.ID, u.Username, u.Photo)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", u.Username, err)
	}
	return nil
}

func (db *appdbimpl) GetUserByName(name string) (*schema.User, error) {
	var u schema.User
	err := db.c.QueryRow("SELECT id, username, photoURL FROM users WHERE username = ?", name).Scan(&u.ID, &u.Username, &u.Photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserDoesNotExist
		}
		return nil, err
	}
	return &u, nil
}

func (db *appdbimpl) GetUserById(userID string) (*schema.User, error) {
	var user schema.User
	err := db.c.QueryRow("SELECT id, username, photoURL FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username, &user.Photo)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *appdbimpl) UpdateUsername(userId, newName string) error {
	res, err := db.c.Exec(`UPDATE users SET username=? WHERE id=?`, newName, userId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrUserDoesNotExist
	}
	return nil
}

func (db *appdbimpl) UpdateUserPhoto(userID string, photoURL string) error {
	var exists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserDoesNotExist
	}
	_, err = db.c.Exec(`UPDATE users SET photoURL=? WHERE id=?`, photoURL, userID)
	return err
}
