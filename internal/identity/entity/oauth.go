package entity

import "time"

type UserConnection struct {
	ID             int64
	UserID         int64
	Provider       string
	ProviderUserID string
}

type OAuthState struct {
	ID           int64
	State        string
	Provider     string
	CodeVerifier string
	RedirectPath string
	ExpiresAt    time.Time
}
