package pkgjwt

// Access token payload
type AccessTokenPayload struct {
	UserID int64  `json:"uid,string"`
	Email  string `json:"email"`
}

// Refresh token payload
type RefreshTokenPayload struct {
	UserID int64 `json:"uid,string"`
}
