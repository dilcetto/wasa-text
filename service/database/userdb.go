package database

import (
	"database/sql"
	"errors"

	"github.com/dilcetto/wasa/service/components/schema"
)

var ErrUserDoesNotExist = errors.New("User does not exist")

func (db *appdbimpl) CreateUser(u schema.User) (schema.User, error) {
	_, err := db.c.Exec("INSERT INTO users(id, name, photo) VALUES (?, ?, ?)", u.Id, u.Name, u.Photo)
	if err != nil {
		var existing schema.User
		if errCheck := db.c.QueryRow("SELECT id, name FROM users WHERE name = ?", u.Name).Scan(&existing.Id, &existing.Name); errCheck != nil {
			if errCheck == sql.ErrNoRows {
				return u, err
			}
		}
		return existing, nil
	}
	return u, nil
}

func (db *appdbimpl) GetUserByName(name string) (schema.User, error) {
	var u schema.User
	if err := db.c.QueryRow("SELECT id, name FROM users WHERE name = ?", name).Scan(&u.Id, &u.Name); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrUserDoesNotExist
		}
		return u, err
	}
	return u, nil
}

func (db *appdbimpl) GetUserById(id string) (schema.User, error) {
	var u schema.User
	if err := db.c.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&u.Id, &u.Name); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrUserDoesNotExist
		}
		return u, err
	}
	return u, nil
}

func (db *appdbimpl) UpdateUsername(userId, newName string) (schema.User, error) {
	res, err := db.c.Exec(`UPDATE users SET name=? WHERE id=?`, newName, userId)
	if err != nil {
		return schema.User{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return schema.User{}, err
	} else if affected == 0 {
		return schema.User{}, ErrUserDoesNotExist
	}
	return db.GetUserById(userId)
}

func (db *appdbimpl) UpdateUserPhoto(userID string, photo []byte) error {
	var exists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserDoesNotExist
	}
	_, err = db.c.Exec(`UPDATE users SET photo=? WHERE id=?`, photo, userID)
	if err != nil {
		return err
	}
	return nil
}
