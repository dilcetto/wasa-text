package schema

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Photo    []byte `json:"photo"`
}

type LoginRequest struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	User
	Token string `json:"token"`
}

type UsernameUpdateResponse = User

type ProfilePhotoUpdateResponse struct {
	Photo []byte `json:"photo"`
}
