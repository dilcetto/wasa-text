package requests

import (
	"net/url"
	"regexp"
)

// LoginRequest contains the username submitted during login
type LoginRequest struct {
	Username string `json:"username"`
}

// IsValid checks if the username meets required format constraints
func (u *LoginRequest) IsValid() bool {
	// Accepts alphanumeric usernames with underscores, 3–16 characters
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,16}$`, u.Username)
	return match
}

type RegisterRequest struct {
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

func (u *RegisterRequest) IsValid() bool {
	// Accepts alphanumeric usernames with underscores, 3–16 characters
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,16}$`, u.Username)
	if !match {
		return false
	}
	// Photo URL must be a valid URL if provided
	if u.PhotoURL != "" {
		_, err := url.ParseRequestURI(u.PhotoURL)
		return err == nil
	}
	return true
}

type SearchRequest struct {
	User         string `json:"user,omitempty"`
	Conversation string `json:"conversation,omitempty"`
}

func (s *SearchRequest) IsValid() bool {
	if s.User != "" {
		match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{1,100}$`, s.User)
		if !match {
			return false
		}
	}
	if s.Conversation != "" {
		match, _ := regexp.MatchString(`^[a-zA-Z0-9_ ]{1,100}$`, s.Conversation)
		if !match {
			return false
		}
	}
	return true
}
