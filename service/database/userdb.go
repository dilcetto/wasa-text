package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dilcetto/wasa/service/components/schema"
)

var ErrUserDoesNotExist = errors.New("User does not exist")

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

func (db *appdbimpl) GetUserById(id string) (schema.User, error) {
	var u schema.User
	if err := db.c.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Username); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrUserDoesNotExist
		}
		return u, err
	}
	return u, nil
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
