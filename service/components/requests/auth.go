package requests

import "regexp"

// LoginRequest represents a login request with a username
type LoginRequest struct {
	Username string `json:"username"` // Add JSON tag for proper serialization
}

// IsValid validates the LoginRequest
func (u *LoginRequest) IsValid() bool {
	// Regex to allow alphanumeric usernames with underscores, 3-16 characters
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,16}$`, u.Username)
	return match
}
