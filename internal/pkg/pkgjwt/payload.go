package pkgjwt

// Access token payload
type AccessTokenPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// Refresh token payload
type RefreshTokenPayload struct {
	Message string `json:"message"`
}
