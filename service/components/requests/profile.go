package requests

import (
	"net/url"
	"regexp"
)

type UsernameUpdateRequest struct {
	Username string `json:"username"`
}

func (u *UsernameUpdateRequest) IsValid() bool {
	match, _ := regexp.MatchString(`^.*?$`, u.Username)
	return len(u.Username) >= 3 && len(u.Username) <= 16 && match
}

type ProfilePhotoUpdateRequest struct {
	PhotoURL string `json:"photoUrl"`
}

func (u *ProfilePhotoUpdateRequest) IsValid() bool {
	_, err := url.ParseRequestURI(u.PhotoURL)
	return err == nil
}
