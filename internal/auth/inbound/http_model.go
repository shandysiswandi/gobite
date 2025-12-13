package inbound

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	MfaRequired  bool   `json:"mfa_required,omitempty"`
	PreAuthToken string `json:"pre_auth_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
