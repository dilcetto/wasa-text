package requests

import (
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
	Photo []byte `json:"photo"`
}

func (u *ProfilePhotoUpdateRequest) IsValid() bool {
	return len(u.Photo) > 0
}
