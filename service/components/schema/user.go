package schema

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Photo    string `json:"photo_url"`
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
	PhotoURL string `json:"photo_url"`
}
