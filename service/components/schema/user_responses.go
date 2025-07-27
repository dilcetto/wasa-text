package schema

type LoginRequest struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

type UsernameUpdateResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

type ProfilePhotoUpdateResponse struct {
	PhotoURL string `json:"photo_url"`
}
