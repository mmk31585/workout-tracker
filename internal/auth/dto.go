package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=32"`
}

type SignupRequest struct {
	LoginRequest
	DisplayName string `json:"display_name" validate:"required,min=3,max=50"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
