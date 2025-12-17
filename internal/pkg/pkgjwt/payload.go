package pkgjwt

// AccessTokenPayload is the custom payload stored inside access tokens.
type AccessTokenPayload struct {
	UserID int64  `json:"uid,string"`
	Email  string `json:"email"`
}

// RefreshTokenPayload is the custom payload stored inside refresh tokens.
type RefreshTokenPayload struct {
	UserID int64 `json:"uid,string"`
}
